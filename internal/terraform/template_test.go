package terraform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prathami1/go-cli/internal/config"
)

func TestRenderTemplate(t *testing.T) {
	// Create a temporary template file for testing
	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "test.tf.tpl")

	templateContent := `# Test template
resource "test_resource" "example" {
  name   = "{{.ProjectName}}"
  region = "{{.Region}}"
  type   = "{{.AppType}}"
}

{{if .EnableDatabase}}
resource "test_database" "db" {
  name = "{{.ProjectName}}-db"
}
{{end}}
`

	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}

	// Test data
	data := &TemplateData{
		ProjectName:    "test-project",
		Region:         "us-east-1",
		AppType:        "nodejs",
		EnableDatabase: true,
	}

	// Render template
	result, err := RenderTemplate(templatePath, data)
	if err != nil {
		t.Fatalf("Template rendering failed: %v", err)
	}

	// Verify template was rendered correctly
	if !strings.Contains(result, "us-east-1") {
		t.Errorf("Template did not render region correctly, got: %s", result)
	}

	if !strings.Contains(result, "test-project") {
		t.Errorf("Template did not render project name correctly, got: %s", result)
	}

	if !strings.Contains(result, "nodejs") {
		t.Errorf("Template did not render app type correctly, got: %s", result)
	}

	if !strings.Contains(result, "test_database") {
		t.Errorf("Template did not render conditional database block, got: %s", result)
	}
}

func TestRenderTemplateWithoutDatabase(t *testing.T) {
	// Create a temporary template file for testing
	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "test.tf.tpl")

	templateContent := `# Test template
resource "test_resource" "example" {
  name = "{{.ProjectName}}"
}

{{if .EnableDatabase}}
resource "test_database" "db" {
  name = "{{.ProjectName}}-db"
}
{{end}}
`

	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}

	// Test data without database
	data := &TemplateData{
		ProjectName:    "test-project",
		EnableDatabase: false,
	}

	// Render template
	result, err := RenderTemplate(templatePath, data)
	if err != nil {
		t.Fatalf("Template rendering failed: %v", err)
	}

	// Verify conditional block was not rendered
	if strings.Contains(result, "test_database") {
		t.Errorf("Template should not render database block when disabled, got: %s", result)
	}
}

func TestGenerateConfigFromTemplates(t *testing.T) {
	// Create test configuration
	cfg := &config.DeploymentConfig{
		ProjectName:   "test-app",
		AppType:       config.NodeJS,
		CloudProvider: config.AWS,
		Region:        "us-east-1",
		Services: config.Services{
			Database:     true,
			Storage:      false,
			LoadBalancer: true,
		},
	}

	// This will fail because we don't have actual template files in test environment
	// But we can test the function structure
	err := GenerateConfigFromTemplates(cfg)
	if err == nil {
		t.Error("Expected error due to missing template directory, but got none")
	}

	// Verify error message mentions template directory
	if !strings.Contains(err.Error(), "template directory not found") {
		t.Errorf("Expected error about template directory, got: %v", err)
	}
}

func TestGetTemplateDir(t *testing.T) {
	tests := []struct {
		provider config.CloudProvider
		expected string
	}{
		{config.AWS, "assets/templates/aws"},
		{config.GCP, "assets/templates/gcp"},
		{config.Azure, "assets/templates/azure"},
	}

	for _, test := range tests {
		result := getTemplateDir(test.provider)
		if result != test.expected {
			t.Errorf("getTemplateDir(%s) = %s, expected %s", test.provider, result, test.expected)
		}
	}
}

func TestTemplateFuncs(t *testing.T) {
	funcs := templateFuncs()

	// Test replace function
	if replaceFn, ok := funcs["replace"]; ok {
		result := replaceFn.(func(string, string, string) string)("hello-world", "-", "_")
		if result != "hello_world" {
			t.Errorf("Replace function failed: got %s, expected hello_world", result)
		}
	} else {
		t.Error("Replace function not found in template functions")
	}

	// Test lower function
	if lowerFn, ok := funcs["lower"]; ok {
		result := lowerFn.(func(string) string)("HELLO")
		if result != "hello" {
			t.Errorf("Lower function failed: got %s, expected hello", result)
		}
	} else {
		t.Error("Lower function not found in template functions")
	}
}
