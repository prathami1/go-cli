package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/prathami1/go-cli/internal/config"
	"github.com/prathami1/go-cli/internal/logger"
	"github.com/prathami1/go-cli/internal/project"
	"github.com/prathami1/go-cli/internal/providers"
	"github.com/prathami1/go-cli/internal/terraform"
	"github.com/prathami1/go-cli/internal/utils"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new deployment configuration and deploy",
	Long: `Start a new deployment configuration by prompting for:

- Project name
- Application type (auto-detected if possible)
- Cloud provider (aws, gcp, azure)
- Region
- Services to enable (database, storage, load balancer)

This will create a .clouddeploy.json file in the current directory
with your deployment configuration and then automatically deploy your application.

This is the complete one-command workflow from setup to deployment.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("🚀 Starting deployment configuration setup...")

		if err := runStart(cmd); err != nil {
			logger.Fatalf("Configuration setup failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Add flags for start command
	startCmd.Flags().BoolP("config-only", "c", false, "Only setup configuration, skip deployment")
	startCmd.Flags().BoolP("auto-approve", "y", false, "Skip interactive approval during deployment")
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func runStart(cmd *cobra.Command) error {
	// Check if config already exists
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	configPath := filepath.Join(currentDir, ".clouddeploy.json")
	if fileExists(configPath) {
		logger.Info("⚠️  Configuration file already exists")
		confirm, err := utils.PromptYesNo("Do you want to overwrite the existing configuration?")
		if err != nil {
			return err
		}
		if !confirm {
			logger.Info("Configuration setup cancelled.")
			return nil
		}
	}

	logger.Info("🚀 Welcome to CloudDeploy! Let's set up your deployment configuration.")

	// Step 1: Analyze project
	logger.Info("🔍 Analyzing your project...")
	analysis, err := project.AnalyzeProject(currentDir)
	if err != nil {
		logger.Errorf("Failed to analyze project: %v", err)
		analysis = nil
	}

	// Step 2: Get project details
	var projectName string
	if analysis != nil && analysis.Path != "" {
		// Extract project name from directory
		dirName := filepath.Base(currentDir)
		confirm, err := utils.PromptYesNo(fmt.Sprintf("Use '%s' as project name?", dirName))
		if err != nil {
			return err
		}
		if confirm {
			projectName = dirName
		} else {
			projectName, err = utils.PromptString("Enter project name", utils.ValidateProjectName)
			if err != nil {
				return err
			}
		}
	} else {
		projectName, err = utils.PromptString("Enter project name", utils.ValidateProjectName)
		if err != nil {
			return err
		}
	}

	// Step 3: Confirm or select app type
	var appType string
	if analysis != nil && analysis.IsHighConfidence() {
		logger.Infof("🎯 Detected %s project with %s confidence", analysis.AppType, analysis.Confidence)

		confirm, err := utils.PromptYesNo(fmt.Sprintf("Use detected app type '%s'?", analysis.GetAppTypeString()))
		if err != nil {
			return err
		}
		if confirm {
			appType = analysis.GetAppTypeString()
		} else {
			appType, err = promptAppType()
			if err != nil {
				return err
			}
		}
	} else {
		logger.Info("🤔 Could not auto-detect project type")
		appType, err = promptAppType()
		if err != nil {
			return err
		}
	}

	// Step 4: Cloud provider selection
	cloudProvider, err := utils.PromptSelect("Select cloud provider", []string{"aws", "gcp", "azure"})
	if err != nil {
		return err
	}

	// Step 5: Region selection
	regions := getRegionsForProvider(cloudProvider)
	region, err := utils.PromptSelect("Select region", regions)
	if err != nil {
		return err
	}

	// Step 6: Services configuration
	logger.Info("🔧 Configure services:")
	enableDatabase, err := utils.PromptYesNo("Enable database?")
	if err != nil {
		return err
	}
	enableStorage, err := utils.PromptYesNo("Enable storage?")
	if err != nil {
		return err
	}
	enableLoadBalancer, err := utils.PromptYesNo("Enable load balancer?")
	if err != nil {
		return err
	}

	// Step 7: Authentication check and auto-install
	logger.Info("🔐 Checking cloud provider authentication...")

	var providerType config.CloudProvider
	switch cloudProvider {
	case "aws":
		providerType = config.AWS
	case "gcp":
		providerType = config.GCP
	case "azure":
		providerType = config.Azure
	default:
		return fmt.Errorf("unsupported cloud provider: %s", cloudProvider)
	}

	if err := providers.CheckAuthentication(providerType); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	logger.Infof("✅ %s authentication verified", cloudProvider)

	// Step 8: Create configuration
	cfg := &config.DeploymentConfig{
		ProjectName:   projectName,
		AppType:       config.AppType(appType),
		CloudProvider: config.CloudProvider(cloudProvider),
		Region:        region,
		Services: config.Services{
			Database:     enableDatabase,
			Storage:      enableStorage,
			LoadBalancer: enableLoadBalancer,
		},
	}

	// Step 9: Save configuration
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	logger.Info("✅ Configuration saved successfully!")
	logger.Info("")
	logger.Info("📋 Configuration Summary:")
	logger.Infof("   Project: %s", projectName)
	logger.Infof("   Type: %s", appType)
	logger.Infof("   Provider: %s", cloudProvider)
	logger.Infof("   Region: %s", region)
	logger.Infof("   Database: %v", enableDatabase)
	logger.Infof("   Storage: %v", enableStorage)
	logger.Infof("   Load Balancer: %v", enableLoadBalancer)
	logger.Info("")

	// Step 10: Check if we should proceed to deployment
	configOnly, _ := cmd.Flags().GetBool("config-only")
	if configOnly {
		logger.Info("💡 Configuration complete. Run 'cdeploy deploy' to deploy your application")
		return nil
	}

	// Step 11: Proceed to deployment automatically
	logger.Info("🚀 Proceeding to deployment...")

	// Create a mock deploy command with the auto-approve flag if needed
	autoApprove, _ := cmd.Flags().GetBool("auto-approve")

	return runDeploymentFromStart(cfg, autoApprove)
}

func promptAppType() (string, error) {
	appTypes := []string{"nodejs", "python", "flask", "docker", "static"}
	return utils.PromptSelect("Select application type", appTypes)
}

func getRegionsForProvider(provider string) []string {
	switch provider {
	case "aws":
		return []string{"us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1"}
	case "gcp":
		return []string{"us-central1", "us-east1", "europe-west1", "asia-southeast1"}
	case "azure":
		return []string{"eastus", "westus2", "westeurope", "southeastasia"}
	default:
		return []string{"us-east-1"}
	}
}

// runDeploymentFromStart runs the deployment logic inline
func runDeploymentFromStart(cfg *config.DeploymentConfig, autoApprove bool) error {
	logger.Info("🚀 Starting deployment process...")
	logger.Infof("📝 Configuration for project: %s", cfg.ProjectName)
	logger.Infof("☁️  Target: %s (%s)", cfg.CloudProvider, cfg.Region)

	// Verify authentication again (in case something changed)
	logger.Info("🔐 Verifying cloud provider authentication...")
	if err := providers.CheckAuthentication(cfg.CloudProvider); err != nil {
		return fmt.Errorf("authentication failed: %w. The authentication process was attempted automatically. If you're a Bloomberg employee, make sure your CORP credentials and B-Unit are ready", err)
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

			if detectionResult.IsHighConfidence() && !autoApprove {
				useDetected, err := utils.PromptYesNo(fmt.Sprintf(
					"Use detected app type (%s) instead of configured (%s)?",
					detectionResult.AppType, cfg.AppType))
				if err != nil {
					return fmt.Errorf("failed to get user input: %w", err)
				}
				if useDetected {
					cfg.AppType = detectionResult.AppType
					logger.Infof("✅ Using detected app type: %s", cfg.AppType)
					// Save the updated configuration
					if err := config.SaveConfig(cfg); err != nil {
						logger.Errorf("Warning: failed to update config with new app type: %v", err)
					}
				}
			}
		} else {
			logger.Infof("✅ Project analysis confirmed app type: %s", cfg.AppType)
		}
	}

	// Create persistent directory for Terraform files
	terraformDir := terraform.TerraformWorkingDir
	if err := os.MkdirAll(terraformDir, 0755); err != nil {
		return fmt.Errorf("failed to create terraform directory: %w", err)
	}

	// Generate Terraform files in persistent directory
	logger.Info("📄 Generating Terraform configuration...")
	if err := terraform.GenerateConfigInDir(cfg, terraformDir); err != nil {
		return fmt.Errorf("failed to generate Terraform config: %w", err)
	}
	logger.Info("✅ Terraform configuration generated")

	// Initialize Terraform in persistent directory
	logger.Info("🔧 Initializing Terraform...")
	if err := terraform.InitInDir(terraformDir); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}
	logger.Info("✅ Terraform initialized")

	// Run Terraform plan
	logger.Info("📋 Running Terraform plan...")
	planOutput, err := terraform.PlanInDir(terraformDir)
	if err != nil {
		return fmt.Errorf("terraform plan failed: %w", err)
	}

	// Display plan output
	fmt.Println(planOutput)

	// Ask for user confirmation unless auto-approve is enabled
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
	applyOutput, err := terraform.ApplyInDir(terraformDir)
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
	outputs, err := terraform.GetOutputsInDir(terraformDir)
	if err != nil {
		logger.Errorf("Warning: failed to get Terraform outputs: %v", err)
	} else {
		displayDeploymentOutputs(cfg, outputs)
	}

	logger.Info("🎉 Deployment completed successfully!")
	logger.Info("")
	logger.Info("✨ Your application has been configured and deployed in one command!")
	logger.Info("💡 You can now access your application using the URLs shown above")
	return nil
}
