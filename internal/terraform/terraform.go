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

// GenerateConfigInDir generates Terraform configuration files in a specific directory
func GenerateConfigInDir(cfg *config.DeploymentConfig, targetDir string) error {
	logger.Debugf("Generating Terraform configuration in directory: %s", targetDir)

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Generate main.tf
	if err := generateMainTFInDir(cfg, targetDir); err != nil {
		return fmt.Errorf("failed to generate main.tf: %w", err)
	}

	// Generate variables.tf
	if err := generateVariablesTFInDir(cfg, targetDir); err != nil {
		return fmt.Errorf("failed to generate variables.tf: %w", err)
	}

	// Generate outputs.tf
	if err := generateOutputsTFInDir(cfg, targetDir); err != nil {
		return fmt.Errorf("failed to generate outputs.tf: %w", err)
	}

	// Generate terraform.tfvars
	if err := generateTFVarsInDir(cfg, targetDir); err != nil {
		return fmt.Errorf("failed to generate terraform.tfvars: %w", err)
	}

	logger.Debugf("Successfully generated Terraform configuration in: %s", targetDir)
	return nil
}

// Helper functions for generating files in specific directories
func generateMainTFInDir(cfg *config.DeploymentConfig, dir string) error {
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

	return writeFileInDir(dir, "main.tf", content)
}

func generateVariablesTFInDir(cfg *config.DeploymentConfig, dir string) error {
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
	return writeFileInDir(dir, "variables.tf", content)
}

func generateOutputsTFInDir(cfg *config.DeploymentConfig, dir string) error {
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

	return writeFileInDir(dir, "outputs.tf", content)
}

func generateTFVarsInDir(cfg *config.DeploymentConfig, dir string) error {
	content := fmt.Sprintf(`# Terraform variables
project_name = "%s"
region = "%s"
app_type = "%s"
enable_database = %t
enable_storage = %t
enable_load_balancer = %t
`, cfg.ProjectName, cfg.Region, cfg.AppType,
		cfg.Services.Database, cfg.Services.Storage, cfg.Services.LoadBalancer)

	return writeFileInDir(dir, "terraform.tfvars", content)
}

func writeFileInDir(dir, filename, content string) error {
	filePath := filepath.Join(dir, filename)
	return os.WriteFile(filePath, []byte(content), 0644)
}

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
	return InitInDir(terraformDir)
}

// InitInDir runs terraform init in a specific directory
func InitInDir(dir string) error {
	logger.Debugf("Running terraform init in directory: %s", dir)

	cmd := exec.Command("terraform", "init")
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform init failed: %w\nOutput: %s", err, string(output))
	}

	logger.Debugf("Terraform init completed successfully in: %s", dir)
	return nil
}

// Plan runs terraform plan and returns the output
func Plan() (string, error) {
	return PlanInDir(terraformDir)
}

// PlanInDir runs terraform plan in a specific directory and returns the output
func PlanInDir(dir string) (string, error) {
	logger.Debugf("Running terraform plan in directory: %s", dir)

	cmd := exec.Command("terraform", "plan", "-out=tfplan")
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("terraform plan failed: %w\nOutput: %s", err, string(output))
	}

	logger.Debugf("Terraform plan completed successfully in: %s", dir)
	return string(output), nil
}

// Apply runs terraform apply
func Apply() (string, error) {
	return ApplyInDir(terraformDir)
}

// ApplyInDir runs terraform apply in a specific directory
func ApplyInDir(dir string) (string, error) {
	logger.Debugf("Running terraform apply in directory: %s", dir)

	cmd := exec.Command("terraform", "apply", "-auto-approve", "tfplan")
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("terraform apply failed: %w\nOutput: %s", err, string(output))
	}

	logger.Debugf("Terraform apply completed successfully in: %s", dir)
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
	return GetOutputsInDir(terraformDir)
}

// GetOutputsInDir retrieves terraform outputs from a specific directory
func GetOutputsInDir(dir string) (map[string]interface{}, error) {
	logger.Debugf("Getting terraform outputs from directory: %s", dir)

	cmd := exec.Command("terraform", "output", "-json")
	cmd.Dir = dir

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

	logger.Debugf("Successfully retrieved %d outputs from: %s", len(result), dir)
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
