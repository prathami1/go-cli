package cmd

import (
	"fmt"
	"time"

	"github.com/prathami1/go-cli/internal/config"
	"github.com/prathami1/go-cli/internal/logger"
	"github.com/prathami1/go-cli/internal/project"
	"github.com/prathami1/go-cli/internal/providers"
	"github.com/prathami1/go-cli/internal/utils"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new deployment configuration",
	Long: `Initialize a new deployment configuration by prompting for:
- Project name
- Application type (static-site, nodejs, flask, docker)
- Cloud provider (aws, gcp, azure)
- Region
- Optional services (database, storage, load balancer)

This will create a .clouddeploy.json file in the current directory
with your configuration and check if you're authenticated with the
selected cloud provider.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runInit(); err != nil {
			logger.Fatalf("Initialization failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit() error {
	logger.Info("🚀 Welcome to CloudDeploy! Let's set up your deployment configuration.")

	// Check if config already exists
	if _, err := config.LoadConfig(); err == nil {
		overwrite, err := utils.PromptYesNo("Configuration file already exists. Overwrite?")
		if err != nil {
			return err
		}
		if !overwrite {
			logger.Info("Initialization cancelled.")
			return nil
		}
	}

	// Prompt for project name
	projectName, err := utils.PromptString("Project name", utils.ValidateProjectName)
	if err != nil {
		return err
	}

	// Analyze current directory to detect project type
	logger.Info("🔍 Analyzing current directory to detect project type...")
	detectionResult, err := project.AnalyzeProject(".")
	if err != nil {
		logger.Errorf("Failed to analyze project: %v", err)
		logger.Info("Proceeding with manual selection...")
		detectionResult = nil
	}

	var appType string
	if detectionResult != nil {
		// Show detection results
		logger.Infof("🎯 Detected: %s", detectionResult.String())

		if detectionResult.IsHighConfidence() {
			// High confidence - ask if user wants to use detected type
			useDetected, err := utils.PromptYesNo(fmt.Sprintf("Use detected app type (%s)?", detectionResult.AppType))
			if err != nil {
				return err
			}
			if useDetected {
				appType = detectionResult.GetAppTypeString()
			}
		}

		if appType == "" {
			// Either low/medium confidence or user declined - show manual selection with suggestion
			logger.Infof("💡 Suggestion: %s (confidence: %s)", detectionResult.AppType, detectionResult.Confidence)
			appType, err = utils.PromptSelect("Application type", config.GetAppTypes())
			if err != nil {
				return err
			}
		}
	} else {
		// No detection result - manual selection
		appType, err = utils.PromptSelect("Application type", config.GetAppTypes())
		if err != nil {
			return err
		}
	}

	// Prompt for cloud provider
	cloudProvider, err := utils.PromptSelect("Cloud provider", config.GetCloudProviders())
	if err != nil {
		return err
	}

	// Check if user is authenticated with the selected cloud provider
	if err := providers.CheckAuthentication(config.CloudProvider(cloudProvider)); err != nil {
		logger.Errorf("Authentication check failed: %v", err)
		logger.Info("Please authenticate with your cloud provider and try again.")
		return err
	}

	// Get regions for the selected cloud provider
	regions, err := providers.GetRegions(config.CloudProvider(cloudProvider))
	if err != nil {
		return fmt.Errorf("failed to get regions: %w", err)
	}

	// Prompt for region
	region, err := utils.PromptSelect("Region", regions)
	if err != nil {
		return err
	}

	// Prompt for optional services
	database, err := utils.PromptYesNo("Include database?")
	if err != nil {
		return err
	}

	storage, err := utils.PromptYesNo("Include storage?")
	if err != nil {
		return err
	}

	loadBalancer, err := utils.PromptYesNo("Include load balancer?")
	if err != nil {
		return err
	}

	// Create configuration
	cfg := &config.DeploymentConfig{
		ProjectName:   projectName,
		AppType:       config.AppType(appType),
		CloudProvider: config.CloudProvider(cloudProvider),
		Region:        region,
		Services: config.Services{
			Database:     database,
			Storage:      storage,
			LoadBalancer: loadBalancer,
		},
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// Save configuration
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	logger.Info("✅ Configuration saved successfully!")
	logger.Infof("📝 Project: %s", cfg.ProjectName)
	logger.Infof("🔧 App Type: %s", cfg.AppType)
	logger.Infof("☁️  Provider: %s", cfg.CloudProvider)
	logger.Infof("🌍 Region: %s", cfg.Region)
	logger.Info("💡 Next step: run 'clouddeploy deploy' to deploy your application")

	return nil
}
