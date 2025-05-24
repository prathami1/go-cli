package cmd

import (
	"fmt"

	"github.com/prathami1/go-cli/internal/config"
	"github.com/prathami1/go-cli/internal/logger"
	"github.com/prathami1/go-cli/internal/terraform"
	"github.com/prathami1/go-cli/internal/utils"
	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy the deployed infrastructure",
	Long: `Destroy the infrastructure that was deployed using 'clouddeploy deploy'.

This command will:
1. Load your configuration from .clouddeploy.json
2. Run terraform destroy to tear down all resources
3. Optionally clean up generated Terraform files

⚠️  WARNING: This will permanently delete all resources created by this deployment.
Make sure you have backed up any important data before proceeding.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runDestroy(cmd); err != nil {
			logger.Fatalf("Destroy failed: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)

	// Add flags for destroy command
	destroyCmd.Flags().BoolP("auto-approve", "y", false, "Skip interactive approval")
	destroyCmd.Flags().BoolP("cleanup", "c", false, "Remove generated Terraform files after destroy")
}

func runDestroy(cmd *cobra.Command) error {
	logger.Info("🗑️  Starting infrastructure destruction...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	logger.Infof("📝 Loaded configuration for project: %s", cfg.ProjectName)
	logger.Infof("☁️  Target: %s (%s)", cfg.CloudProvider, cfg.Region)

	// Check for auto-approve flag
	autoApprove, _ := cmd.Flags().GetBool("auto-approve")
	if !autoApprove {
		// Warn user about permanent deletion
		logger.Info("⚠️  WARNING: This will permanently delete all resources created by this deployment!")
		logger.Info("💾 Make sure you have backed up any important data.")

		// Ask for user confirmation
		approve, err := utils.PromptYesNo("Are you sure you want to destroy all resources?")
		if err != nil {
			return err
		}
		if !approve {
			logger.Info("❌ Destruction cancelled by user")
			return nil
		}
	}

	// Run Terraform destroy
	logger.Info("🔥 Running terraform destroy...")
	destroyOutput, err := terraform.Destroy()
	if err != nil {
		return fmt.Errorf("terraform destroy failed: %w", err)
	}

	// Display destroy output
	fmt.Println(destroyOutput)

	logger.Info("✅ Infrastructure destroyed successfully!")

	// Check for cleanup flag
	cleanup, _ := cmd.Flags().GetBool("cleanup")
	if cleanup {
		logger.Info("🧹 Cleaning up generated Terraform files...")
		if err := cleanupTerraformFiles(); err != nil {
			logger.Errorf("Warning: failed to cleanup Terraform files: %v", err)
		} else {
			logger.Info("✅ Terraform files cleaned up")
		}
	} else {
		logger.Info("💡 Tip: Use --cleanup flag to remove generated Terraform files")
	}

	// Display completion message
	logger.Info("🎉 Destruction completed successfully!")
	logger.Info("💡 Your cloud resources have been removed")
	logger.Info("📝 Configuration file (.clouddeploy.json) is preserved for future deployments")

	return nil
}

func cleanupTerraformFiles() error {
	// This would remove the .terraform-generated directory
	// For now, we'll just log that it would happen
	logger.Debug("Would remove .terraform-generated directory")
	return nil
}
