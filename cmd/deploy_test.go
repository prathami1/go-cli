package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/prathami1/go-cli/internal/config"
)

func TestPersistentTerraformDir(t *testing.T) {
	// Test that we can create the persistent terraform directory
	terraformDir := ".clouddeploy-tf"

	// Clean up any existing directory first
	os.RemoveAll(terraformDir)

	// Test creating persistent directory
	if err := os.MkdirAll(terraformDir, 0755); err != nil {
		t.Fatalf("Failed to create terraform directory: %v", err)
	}

	// Verify directory exists
	if _, err := os.Stat(terraformDir); os.IsNotExist(err) {
		t.Errorf("Terraform directory was not created: %s", terraformDir)
	}

	// Verify we can write to the directory
	testFile := filepath.Join(terraformDir, "test.tf")
	if err := os.WriteFile(testFile, []byte("# test terraform file"), 0644); err != nil {
		t.Errorf("Failed to write to terraform directory: %v", err)
	}

	// Clean up
	defer os.RemoveAll(terraformDir)
}

func TestDisplayDeploymentOutputs(t *testing.T) {
	cfg := &config.DeploymentConfig{
		ProjectName:   "test-project",
		AppType:       config.NodeJS,
		CloudProvider: config.AWS,
		Region:        "us-east-1",
		Services: config.Services{
			Database:     true,
			Storage:      false,
			LoadBalancer: true,
		},
		LastDeployment: "2023-12-01T10:00:00Z",
	}

	outputs := map[string]interface{}{
		"public_ip":         "192.168.1.1",
		"database_url":      "postgresql://user:pass@host:5432/db",
		"load_balancer_url": "https://lb.example.com",
	}

	// This is more of an integration test to ensure the function doesn't panic
	// In a real test environment, we'd want to capture the output
	displayDeploymentOutputs(cfg, outputs)
}

func TestDisplayDeploymentOutputsEmptyOutputs(t *testing.T) {
	cfg := &config.DeploymentConfig{
		ProjectName:   "test-project",
		AppType:       config.StaticSite,
		CloudProvider: config.GCP,
		Region:        "us-central1",
	}

	// Test with empty outputs
	outputs := map[string]interface{}{}
	displayDeploymentOutputs(cfg, outputs)
}

func TestDisplayDeploymentOutputsNilOutputs(t *testing.T) {
	cfg := &config.DeploymentConfig{
		ProjectName:   "test-project",
		AppType:       config.Flask,
		CloudProvider: config.Azure,
		Region:        "eastus",
	}

	// Test with nil outputs
	displayDeploymentOutputs(cfg, nil)
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Integration test helpers - these would require more complex setup in real scenarios

func TestDeployCommandFlags(t *testing.T) {
	// Test that the deploy command has the expected flags
	if deployCmd.Flags().Lookup("auto-approve") == nil {
		t.Error("deploy command should have auto-approve flag")
	}

	if deployCmd.Flags().Lookup("plan-only") == nil {
		t.Error("deploy command should have plan-only flag")
	}
}

func TestDeployCommandBasicProperties(t *testing.T) {
	// Test basic command properties
	if deployCmd.Use != "deploy" {
		t.Errorf("Expected command use 'deploy', got '%s'", deployCmd.Use)
	}

	if deployCmd.Short == "" {
		t.Error("Deploy command should have a short description")
	}

	if deployCmd.Long == "" {
		t.Error("Deploy command should have a long description")
	}

	if deployCmd.Run == nil {
		t.Error("Deploy command should have a run function")
	}
}
