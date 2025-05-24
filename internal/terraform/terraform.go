package terraform

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/prathami1/go-cli/internal/config"
	"github.com/prathami1/go-cli/internal/logger"
)

const terraformDir = ".terraform-generated"

// GenerateConfig generates Terraform configuration files based on the deployment config
func GenerateConfig(cfg *config.DeploymentConfig) error {
	logger.Debug("Generating Terraform configuration...")

	// Create terraform directory if it doesn't exist
	if err := os.MkdirAll(terraformDir, 0755); err != nil {
		return fmt.Errorf("failed to create terraform directory: %w", err)
	}

	// Generate main.tf
	if err := generateMainTF(cfg); err != nil {
		return fmt.Errorf("failed to generate main.tf: %w", err)
	}

	// Generate variables.tf
	if err := generateVariablesTF(cfg); err != nil {
		return fmt.Errorf("failed to generate variables.tf: %w", err)
	}

	// Generate outputs.tf
	if err := generateOutputsTF(cfg); err != nil {
		return fmt.Errorf("failed to generate outputs.tf: %w", err)
	}

	// Generate terraform.tfvars
	if err := generateTFVars(cfg); err != nil {
		return fmt.Errorf("failed to generate terraform.tfvars: %w", err)
	}

	return nil
}

// Init runs terraform init
func Init() error {
	logger.Debug("Running terraform init...")

	cmd := exec.Command("terraform", "init")
	cmd.Dir = terraformDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform init failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Plan runs terraform plan and returns the output
func Plan() (string, error) {
	logger.Debug("Running terraform plan...")

	cmd := exec.Command("terraform", "plan", "-out=tfplan")
	cmd.Dir = terraformDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("terraform plan failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// Apply runs terraform apply
func Apply() (string, error) {
	logger.Debug("Running terraform apply...")

	cmd := exec.Command("terraform", "apply", "-auto-approve", "tfplan")
	cmd.Dir = terraformDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("terraform apply failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// Destroy runs terraform destroy
func Destroy() (string, error) {
	logger.Debug("Running terraform destroy...")

	cmd := exec.Command("terraform", "destroy", "-auto-approve")
	cmd.Dir = terraformDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("terraform destroy failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// GetOutputs retrieves terraform outputs
func GetOutputs() (map[string]interface{}, error) {
	logger.Debug("Getting terraform outputs...")

	cmd := exec.Command("terraform", "output", "-json")
	cmd.Dir = terraformDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("terraform output failed: %w", err)
	}

	var outputs map[string]interface{}
	if err := json.Unmarshal(output, &outputs); err != nil {
		return nil, fmt.Errorf("failed to parse terraform outputs: %w", err)
	}

	// Extract values from terraform output format
	result := make(map[string]interface{})
	for key, value := range outputs {
		if outputMap, ok := value.(map[string]interface{}); ok {
			if val, exists := outputMap["value"]; exists {
				result[key] = val
			}
		}
	}

	return result, nil
}

func generateMainTF(cfg *config.DeploymentConfig) error {
	var content string

	switch cfg.CloudProvider {
	case config.AWS:
		content = generateAWSMainTF(cfg)
	case config.GCP:
		content = generateGCPMainTF(cfg)
	case config.Azure:
		content = generateAzureMainTF(cfg)
	default:
		return fmt.Errorf("unsupported cloud provider: %s", cfg.CloudProvider)
	}

	return writeFile("main.tf", content)
}

func generateVariablesTF(cfg *config.DeploymentConfig) error {
	content := `# Variables for deployment configuration

variable "project_name" {
  description = "Name of the project"
  type        = string
}

variable "region" {
  description = "Deployment region"
  type        = string
}

variable "app_type" {
  description = "Type of application"
  type        = string
}

variable "enable_database" {
  description = "Whether to enable database"
  type        = bool
  default     = false
}

variable "enable_storage" {
  description = "Whether to enable storage"
  type        = bool
  default     = false
}

variable "enable_load_balancer" {
  description = "Whether to enable load balancer"
  type        = bool
  default     = false
}
`
	return writeFile("variables.tf", content)
}

func generateOutputsTF(cfg *config.DeploymentConfig) error {
	var content string

	switch cfg.CloudProvider {
	case config.AWS:
		content = generateAWSOutputsTF(cfg)
	case config.GCP:
		content = generateGCPOutputsTF(cfg)
	case config.Azure:
		content = generateAzureOutputsTF(cfg)
	default:
		return fmt.Errorf("unsupported cloud provider: %s", cfg.CloudProvider)
	}

	return writeFile("outputs.tf", content)
}

func generateTFVars(cfg *config.DeploymentConfig) error {
	content := fmt.Sprintf(`project_name = "%s"
region = "%s"
app_type = "%s"
enable_database = %t
enable_storage = %t
enable_load_balancer = %t
`, cfg.ProjectName, cfg.Region, cfg.AppType, cfg.Services.Database, cfg.Services.Storage, cfg.Services.LoadBalancer)

	return writeFile("terraform.tfvars", content)
}

func writeFile(filename, content string) error {
	path := filepath.Join(terraformDir, filename)
	return os.WriteFile(path, []byte(content), 0644)
}
