package terraform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/prathami1/go-cli/internal/config"
)

func TestGenerateConfigInDir(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "terraform-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test configuration
	cfg := &config.DeploymentConfig{
		ProjectName:   "test-project",
		AppType:       config.NodeJS,
		CloudProvider: config.AWS,
		Region:        "us-east-1",
		Services: config.Services{
			Database:     true,
			Storage:      true,
			LoadBalancer: true,
		},
	}

	// Generate config in temp directory
	err = GenerateConfigInDir(cfg, tempDir)
	if err != nil {
		t.Fatalf("GenerateConfigInDir failed: %v", err)
	}

	// Check that all expected files were created
	expectedFiles := []string{"provider.tf", "compute.tf", "variables.tf", "outputs.tf", "terraform.tfvars"}
	for _, filename := range expectedFiles {
		filePath := filepath.Join(tempDir, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", filename)
		}
	}
}

func TestGenerateConfigInDirInvalidProvider(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "terraform-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.DeploymentConfig{
		ProjectName:   "test-project",
		AppType:       config.NodeJS,
		CloudProvider: "invalid-provider",
		Region:        "us-east-1",
	}

	err = GenerateConfigInDir(cfg, tempDir)
	if err == nil {
		t.Error("Expected error for invalid cloud provider")
	}

	if !contains(err.Error(), "template directory not found") {
		t.Errorf("Expected error message about template directory, got: %v", err)
	}
}

func TestWriteFileInDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "terraform-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testContent := "# Test Terraform file\nresource \"test\" \"example\" {}"
	filename := "test.tf"

	err = writeFileInDir(tempDir, filename, testContent)
	if err != nil {
		t.Fatalf("writeFileInDir failed: %v", err)
	}

	// Verify file was created and has correct content
	filePath := filepath.Join(tempDir, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("File content mismatch.\nExpected: %s\nGot: %s", testContent, string(content))
	}
}

func TestGenerateConfigInDirWithTemplates(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "terraform-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name     string
		provider config.CloudProvider
		wantErr  bool
	}{
		{"AWS provider", config.AWS, false},
		{"GCP provider", config.GCP, false},
		{"Azure provider", config.Azure, false},
		{"Invalid provider", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.DeploymentConfig{
				ProjectName:   "test-project",
				AppType:       config.NodeJS,
				CloudProvider: tt.provider,
				Region:        "test-region",
				Services: config.Services{
					Database:     true,
					Storage:      false,
					LoadBalancer: true,
				},
			}

			err := GenerateConfigInDir(cfg, tempDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateConfigInDir() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify core files were created
				expectedFiles := []string{"variables.tf", "outputs.tf", "terraform.tfvars"}
				for _, file := range expectedFiles {
					filePath := filepath.Join(tempDir, file)
					if _, err := os.Stat(filePath); os.IsNotExist(err) {
						t.Errorf("%s file was not created for provider %s", file, tt.provider)
					}
				}
			}
		})
	}
}

func TestGenerateVariablesTFInDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "terraform-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.DeploymentConfig{
		ProjectName:   "test-project",
		AppType:       config.Flask,
		CloudProvider: config.AWS,
		Region:        "us-west-2",
	}

	err = generateVariablesTFInDir(cfg, tempDir)
	if err != nil {
		t.Fatalf("generateVariablesTFInDir failed: %v", err)
	}

	// Verify file was created
	filePath := filepath.Join(tempDir, "variables.tf")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("variables.tf file was not created")
	}

	// Read and verify content contains expected variables
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read variables.tf: %v", err)
	}

	expectedVars := []string{"project_name", "region", "app_type", "enable_database", "enable_storage", "enable_load_balancer"}
	for _, variable := range expectedVars {
		if !contains(string(content), variable) {
			t.Errorf("variables.tf should contain variable %s", variable)
		}
	}
}

func TestGenerateOutputsTFInDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "terraform-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name     string
		provider config.CloudProvider
		wantErr  bool
	}{
		{"AWS provider", config.AWS, false},
		{"GCP provider", config.GCP, false},
		{"Azure provider", config.Azure, false},
		{"Invalid provider", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.DeploymentConfig{
				ProjectName:   "test-project",
				AppType:       config.Docker,
				CloudProvider: tt.provider,
				Region:        "test-region",
			}

			err := generateOutputsTFInDir(cfg, tempDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateOutputsTFInDir() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify file was created
				filePath := filepath.Join(tempDir, "outputs.tf")
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Errorf("outputs.tf file was not created for provider %s", tt.provider)
				}
			}
		})
	}
}

func TestGenerateTFVarsInDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "terraform-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.DeploymentConfig{
		ProjectName:   "my-awesome-project",
		AppType:       config.StaticSite,
		CloudProvider: config.GCP,
		Region:        "us-central1",
		Services: config.Services{
			Database:     false,
			Storage:      true,
			LoadBalancer: false,
		},
	}

	err = generateTFVarsInDir(cfg, tempDir)
	if err != nil {
		t.Fatalf("generateTFVarsInDir failed: %v", err)
	}

	// Verify file was created
	filePath := filepath.Join(tempDir, "terraform.tfvars")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("terraform.tfvars file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read terraform.tfvars: %v", err)
	}

	contentStr := string(content)

	// Check that configuration values are present
	expectedValues := map[string]string{
		"my-awesome-project":           cfg.ProjectName,
		"us-central1":                  cfg.Region,
		"static-site":                  string(cfg.AppType),
		"enable_database = false":      "database setting",
		"enable_storage = true":        "storage setting",
		"enable_load_balancer = false": "load balancer setting",
	}

	for expected, description := range expectedValues {
		if !contains(contentStr, expected) {
			t.Errorf("terraform.tfvars should contain %s for %s", expected, description)
		}
	}
}

func TestWriteFileInDirInvalidPath(t *testing.T) {
	// Test writing to invalid directory
	invalidDir := "/invalid/nonexistent/path"
	err := writeFileInDir(invalidDir, "test.tf", "content")
	if err == nil {
		t.Error("Expected error when writing to invalid directory")
	}
}

// Helper function for testing
func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsSubstring(s, substr)
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Integration tests for directory operations
func TestTerraformDirectoryOperationsIntegration(t *testing.T) {
	// Create a temporary directory for full integration test
	tempDir, err := os.MkdirTemp("", "terraform-integration-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test full config generation for each cloud provider
	providers := []config.CloudProvider{config.AWS, config.GCP, config.Azure}
	appTypes := []config.AppType{config.NodeJS, config.Flask, config.Docker, config.StaticSite}

	for _, provider := range providers {
		for _, appType := range appTypes {
			t.Run(string(provider)+"_"+string(appType), func(t *testing.T) {
				// Create subdirectory for this test
				testDir := filepath.Join(tempDir, string(provider)+"_"+string(appType))

				cfg := &config.DeploymentConfig{
					ProjectName:   "test-project",
					AppType:       appType,
					CloudProvider: provider,
					Region:        "test-region",
					Services: config.Services{
						Database:     true,
						Storage:      true,
						LoadBalancer: true,
					},
				}

				// Test full generation
				err := GenerateConfigInDir(cfg, testDir)
				if err != nil {
					t.Errorf("Failed to generate config for %s/%s: %v", provider, appType, err)
					return
				}

				// Verify all files exist and have content
				expectedFiles := []string{"provider.tf", "compute.tf", "variables.tf", "outputs.tf", "terraform.tfvars"}
				for _, filename := range expectedFiles {
					filePath := filepath.Join(testDir, filename)
					if stat, err := os.Stat(filePath); os.IsNotExist(err) {
						t.Errorf("File %s missing for %s/%s", filename, provider, appType)
					} else if stat.Size() == 0 {
						t.Errorf("File %s is empty for %s/%s", filename, provider, appType)
					}
				}
			})
		}
	}
}
