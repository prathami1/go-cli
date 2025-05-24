package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/prathami1/go-cli/internal/config"
	"github.com/prathami1/go-cli/internal/logger"
	"github.com/prathami1/go-cli/internal/project"
	"github.com/prathami1/go-cli/internal/providers"
	"github.com/prathami1/go-cli/internal/terraform"
	"github.com/prathami1/go-cli/internal/utils"
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy your application to the cloud",
	Long: `Deploy your application to the configured cloud provider using Terraform.

This command will:
1. Load your configuration from .clouddeploy.json
2. Verify cloud provider authentication
3. Generate appropriate Terraform files
4. Run terraform init, plan, and apply
5. Output deployment information (URLs, credentials, etc.)

Make sure you have run 'clouddeploy init' first to create your configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDeploy(cmd); err != nil {
			logger.Fatalf("Deployment failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	// Add flags for deploy command
	deployCmd.Flags().BoolP("auto-approve", "y", false, "Skip interactive approval of plan")
	deployCmd.Flags().BoolP("plan-only", "p", false, "Only run terraform plan, don't apply")
}

func runDeploy(cmd *cobra.Command) error {
	logger.Info("🚀 Starting deployment process...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w. Run 'clouddeploy init' first", err)
	}

	logger.Infof("📝 Loaded configuration for project: %s", cfg.ProjectName)
	logger.Infof("☁️  Target: %s (%s)", cfg.CloudProvider, cfg.Region)

	// Verify authentication
	logger.Info("🔐 Verifying cloud provider authentication...")
	if err := providers.CheckAuthentication(cfg.CloudProvider); err != nil {
		return fmt.Errorf("authentication failed: %w. Please authenticate with your cloud provider", err)
	}
	logger.Info("✅ Authentication verified")

	// Analyze current project to verify app type
	logger.Info("🔍 Analyzing current project...")
	detectionResult, err := project.AnalyzeProject(".")
	if err != nil {
		logger.Errorf("Warning: project analysis failed: %v", err)
		logger.Info("Proceeding with configured app type...")
	} else {
		// Compare detected type with configured type
		if detectionResult.AppType != cfg.AppType {
			logger.Infof("⚠️  Detected app type (%s) differs from configured type (%s)",
				detectionResult.AppType, cfg.AppType)

			if detectionResult.IsHighConfidence() {
				useDetected, err := utils.PromptYesNo(fmt.Sprintf(
					"Use detected app type (%s) instead of configured (%s)?",
					detectionResult.AppType, cfg.AppType))
				if err != nil {
					return fmt.Errorf("failed to get user input: %w", err)
				}
				if useDetected {
					cfg.AppType = detectionResult.AppType
					logger.Infof("✅ Using detected app type: %s", cfg.AppType)
				}
			}
		} else {
			logger.Infof("✅ Project analysis confirmed app type: %s", cfg.AppType)
		}
	}

	// Create temporary directory for Terraform files
	tempDir, err := createTempTerraformDir(cfg.ProjectName)
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Setup cleanup function
	defer func() {
		if err := cleanupTempDir(tempDir); err != nil {
			logger.Errorf("Warning: failed to cleanup temporary directory: %v", err)
		}
	}()

	// Generate Terraform files in temp directory
	logger.Info("📄 Generating Terraform configuration...")
	if err := terraform.GenerateConfigInDir(cfg, tempDir); err != nil {
		return fmt.Errorf("failed to generate Terraform config: %w", err)
	}
	logger.Info("✅ Terraform configuration generated")

	// Initialize Terraform in temp directory
	logger.Info("🔧 Initializing Terraform...")
	if err := terraform.InitInDir(tempDir); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}
	logger.Info("✅ Terraform initialized")

	// Run Terraform plan
	logger.Info("📋 Running Terraform plan...")
	planOutput, err := terraform.PlanInDir(tempDir)
	if err != nil {
		return fmt.Errorf("terraform plan failed: %w", err)
	}

	// Display plan output
	fmt.Println(planOutput)

	// Check if this is plan-only mode
	planOnly, _ := cmd.Flags().GetBool("plan-only")
	if planOnly {
		logger.Info("📋 Plan-only mode: deployment stopped after plan")
		return nil
	}

	// Check for auto-approve flag
	autoApprove, _ := cmd.Flags().GetBool("auto-approve")
	if !autoApprove {
		// Ask for user confirmation
		approve, err := utils.PromptYesNo("Do you want to apply these changes?")
		if err != nil {
			return fmt.Errorf("failed to get user confirmation: %w", err)
		}
		if !approve {
			logger.Info("❌ Deployment cancelled by user")
			return nil
		}
	} else {
		logger.Info("🔄 Auto-approve enabled, proceeding with deployment...")
	}

	// Apply Terraform configuration
	logger.Info("🚀 Applying Terraform configuration...")
	applyOutput, err := terraform.ApplyInDir(tempDir)
	if err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}

	// Display apply output
	fmt.Println(applyOutput)

	// Update configuration with deployment timestamp
	cfg.LastDeployment = time.Now().Format(time.RFC3339)
	if err := config.SaveConfig(cfg); err != nil {
		logger.Errorf("Warning: failed to update config with deployment timestamp: %v", err)
	}

	// Get outputs from Terraform
	logger.Info("📊 Retrieving deployment outputs...")
	outputs, err := terraform.GetOutputsInDir(tempDir)
	if err != nil {
		logger.Errorf("Warning: failed to get Terraform outputs: %v", err)
	} else {
		displayDeploymentOutputs(cfg, outputs)
	}

	logger.Info("🎉 Deployment completed successfully!")
	return nil
}

func displayDeploymentOutputs(cfg *config.DeploymentConfig, outputs map[string]interface{}) {
	logger.Info("📊 Deployment Information:")
	logger.Info("=" + fmt.Sprintf("%50s", "="))

	// Display basic info
	logger.Infof("Project: %s", cfg.ProjectName)
	logger.Infof("App Type: %s", cfg.AppType)
	logger.Infof("Provider: %s", cfg.CloudProvider)
	logger.Infof("Region: %s", cfg.Region)

	if cfg.LastDeployment != "" {
		logger.Infof("Deployed: %s", cfg.LastDeployment)
	}

	// Display Terraform outputs
	if len(outputs) > 0 {
		logger.Info("\n🔗 Outputs:")
		for key, value := range outputs {
			logger.Infof("%s: %v", key, value)
		}
	}

	// Display next steps
	logger.Info("\n💡 Next Steps:")
	logger.Info("- Visit the provided URLs to access your application")
	logger.Info("- Check the outputs above for important credentials and endpoints")
	logger.Info("- Run 'clouddeploy destroy' when you want to tear down the infrastructure")

	if cfg.Services.Database {
		logger.Info("- Database credentials are available in the outputs above")
	}
}

// createTempTerraformDir creates a temporary directory for Terraform files
func createTempTerraformDir(projectName string) (string, error) {
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("clouddeploy-%s-%d", projectName, time.Now().Unix()))

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	logger.Debugf("Created temporary directory: %s", tempDir)
	return tempDir, nil
}

// cleanupTempDir removes the temporary directory
func cleanupTempDir(tempDir string) error {
	if tempDir == "" {
		return nil
	}

	logger.Debugf("Cleaning up temporary directory: %s", tempDir)
	return os.RemoveAll(tempDir)
}
