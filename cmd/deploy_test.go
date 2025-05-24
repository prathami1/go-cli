package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/prathami1/go-cli/internal/config"
)

func TestCreateTempTerraformDir(t *testing.T) {
	projectName := "test-project"

	// Test creating temp directory
	tempDir, err := createTempTerraformDir(projectName)
	if err != nil {
		t.Fatalf("createTempTerraformDir failed: %v", err)
	}

	// Verify directory exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Temporary directory was not created: %s", tempDir)
	}

	// Verify directory name contains project name
	if !contains(tempDir, projectName) {
		t.Errorf("Temporary directory name should contain project name: %s", tempDir)
	}

	// Clean up
	defer func() {
		if err := cleanupTempDir(tempDir); err != nil {
			t.Errorf("Failed to cleanup temp directory: %v", err)
		}
	}()

	// Verify we can write to the directory
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Errorf("Failed to write to temp directory: %v", err)
	}
}

func TestCleanupTempDir(t *testing.T) {
	// Create a temporary directory
	tempDir := filepath.Join(os.TempDir(), "test-cleanup")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a test file in it
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test cleanup
	if err := cleanupTempDir(tempDir); err != nil {
		t.Errorf("cleanupTempDir failed: %v", err)
	}

	// Verify directory is gone
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Errorf("Temporary directory still exists after cleanup: %s", tempDir)
	}
}

func TestCleanupTempDirEmptyPath(t *testing.T) {
	// Test cleanup with empty path (should not error)
	if err := cleanupTempDir(""); err != nil {
		t.Errorf("cleanupTempDir with empty path should not error: %v", err)
	}
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
