package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/prathami1/go-cli/internal/config"
	"github.com/prathami1/go-cli/internal/logger"
)

// TemplateData represents the data structure passed to templates
type TemplateData struct {
	ProjectName        string
	AppType            string
	CloudProvider      string
	Region             string
	Environment        string
	EnableDatabase     bool
	EnableStorage      bool
	EnableLoadBalancer bool
}

// RenderTemplate renders a Terraform template with the given data
func RenderTemplate(templatePath string, data interface{}) (string, error) {
	logger.Debugf("Rendering template: %s", templatePath)

	// Read template file
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	// Create template with custom functions
	tmpl, err := template.New(filepath.Base(templatePath)).
		Funcs(templateFuncs()).
		Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}

	// Execute template
	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templatePath, err)
	}

	return result.String(), nil
}

// GenerateConfigFromTemplates generates Terraform configuration using templates
func GenerateConfigFromTemplates(cfg *config.DeploymentConfig) error {
	logger.Debug("Generating Terraform configuration from templates...")

	// Create terraform directory if it doesn't exist
	if err := os.MkdirAll(terraformDir, 0755); err != nil {
		return fmt.Errorf("failed to create terraform directory: %w", err)
	}

	// Prepare template data
	templateData := &TemplateData{
		ProjectName:        cfg.ProjectName,
		AppType:            string(cfg.AppType),
		CloudProvider:      string(cfg.CloudProvider),
		Region:             cfg.Region,
		Environment:        "production", // Default environment
		EnableDatabase:     cfg.Services.Database,
		EnableStorage:      cfg.Services.Storage,
		EnableLoadBalancer: cfg.Services.LoadBalancer,
	}

	// Get template directory path
	templateDir := getTemplateDir(cfg.CloudProvider)
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		return fmt.Errorf("template directory not found: %s", templateDir)
	}

	// Render each template
	templates := []string{"provider.tf.tpl", "compute.tf.tpl"}

	// Add optional templates based on services
	if cfg.Services.Database {
		templates = append(templates, "database.tf.tpl")
	}
	if cfg.Services.Storage {
		templates = append(templates, "storage.tf.tpl")
	}
	if cfg.Services.LoadBalancer {
		templates = append(templates, "loadbalancer.tf.tpl")
	}

	// Render and write each template
	for _, templateName := range templates {
		templatePath := filepath.Join(templateDir, templateName)
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			logger.Debugf("Template not found, skipping: %s", templatePath)
			continue
		}

		// Render template
		rendered, err := RenderTemplate(templatePath, templateData)
		if err != nil {
			return fmt.Errorf("failed to render template %s: %w", templateName, err)
		}

		// Write rendered content to file
		outputFileName := strings.TrimSuffix(templateName, ".tpl")
		outputPath := filepath.Join(terraformDir, outputFileName)

		if err := os.WriteFile(outputPath, []byte(rendered), 0644); err != nil {
			return fmt.Errorf("failed to write rendered template to %s: %w", outputPath, err)
		}

		logger.Debugf("Generated: %s", outputPath)
	}

	// Generate variables.tf and outputs.tf
	if err := generateTemplateVariablesTF(cfg); err != nil {
		return fmt.Errorf("failed to generate variables.tf: %w", err)
	}

	if err := generateTemplateOutputsTF(cfg); err != nil {
		return fmt.Errorf("failed to generate outputs.tf: %w", err)
	}

	// Generate terraform.tfvars
	if err := generateTFVars(cfg); err != nil {
		return fmt.Errorf("failed to generate terraform.tfvars: %w", err)
	}

	logger.Debug("Terraform configuration generated successfully from templates")
	return nil
}

// getTemplateDir returns the template directory path for the given cloud provider
func getTemplateDir(provider config.CloudProvider) string {
	switch provider {
	case config.AWS:
		return "assets/templates/aws"
	case config.GCP:
		return "assets/templates/gcp"
	case config.Azure:
		return "assets/templates/azure"
	default:
		return ""
	}
}

// templateFuncs returns custom template functions
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"replace": strings.ReplaceAll,
		"lower":   strings.ToLower,
		"upper":   strings.ToUpper,
		"title":   strings.Title,
	}
}

// generateTemplateVariablesTF generates the variables.tf file for templates
func generateTemplateVariablesTF(cfg *config.DeploymentConfig) error {
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

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
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
	return writeTemplateFile("variables.tf", content)
}

// generateTemplateOutputsTF generates the outputs.tf file for templates
func generateTemplateOutputsTF(cfg *config.DeploymentConfig) error {
	var content strings.Builder

	content.WriteString("# Outputs for deployment\n\n")
	content.WriteString(`output "project_name" {
  description = "Project name"
  value       = var.project_name
}

output "region" {
  description = "Deployment region"
  value       = var.region
}

output "app_type" {
  description = "Application type"
  value       = var.app_type
}

output "environment" {
  description = "Environment"
  value       = var.environment
}
`)

	// Add cloud-provider specific outputs
	switch cfg.CloudProvider {
	case config.AWS:
		content.WriteString(generateAWSOutputs(cfg))
	case config.GCP:
		content.WriteString(generateGCPOutputs(cfg))
	case config.Azure:
		content.WriteString(generateAzureOutputs(cfg))
	}

	return writeTemplateFile("outputs.tf", content.String())
}

// writeTemplateFile writes content to a file in the terraform directory
func writeTemplateFile(filename, content string) error {
	path := filepath.Join(terraformDir, filename)
	return os.WriteFile(path, []byte(content), 0644)
}

func generateAWSOutputs(cfg *config.DeploymentConfig) string {
	var outputs strings.Builder

	switch cfg.AppType {
	case config.StaticSite:
		outputs.WriteString(`
output "website_url" {
  description = "Website URL"
  value       = "https://${aws_s3_bucket_website_configuration.main.website_endpoint}"
}

output "s3_bucket_name" {
  description = "S3 bucket name"
  value       = aws_s3_bucket.static_site.bucket
}
`)
	case config.NodeJS, config.Flask:
		outputs.WriteString(`
output "instance_public_ip" {
  description = "Public IP address of the EC2 instance"
  value       = aws_instance.app.public_ip
}

output "application_url" {
  description = "Application URL"
  value       = "http://${aws_instance.app.public_ip}"
}
`)
	case config.Docker:
		outputs.WriteString(`
output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = aws_ecs_cluster.main.name
}
`)
	}

	if cfg.Services.Database {
		outputs.WriteString(`
output "database_endpoint" {
  description = "Database endpoint"
  value       = aws_db_instance.main.endpoint
  sensitive   = true
}
`)
	}

	if cfg.Services.LoadBalancer {
		outputs.WriteString(`
output "load_balancer_dns" {
  description = "Load balancer DNS name"
  value       = aws_lb.main.dns_name
}
`)
	}

	return outputs.String()
}

func generateGCPOutputs(cfg *config.DeploymentConfig) string {
	var outputs strings.Builder

	switch cfg.AppType {
	case config.StaticSite:
		outputs.WriteString(`
output "website_url" {
  description = "Website URL"
  value       = "https://storage.googleapis.com/${google_storage_bucket.static_site.name}/index.html"
}
`)
	case config.NodeJS, config.Flask:
		outputs.WriteString(`
output "instance_external_ip" {
  description = "Instance external IP"
  value       = google_compute_instance.app.network_interface[0].access_config[0].nat_ip
}

output "application_url" {
  description = "Application URL"
  value       = "http://${google_compute_instance.app.network_interface[0].access_config[0].nat_ip}"
}
`)
	case config.Docker:
		outputs.WriteString(`
output "cloud_run_url" {
  description = "Cloud Run service URL"
  value       = google_cloud_run_service.app.status[0].url
}
`)
	}

	if cfg.Services.LoadBalancer {
		outputs.WriteString(`
output "load_balancer_ip" {
  description = "Load balancer IP address"
  value       = google_compute_global_address.app.address
}
`)
	}

	return outputs.String()
}

func generateAzureOutputs(cfg *config.DeploymentConfig) string {
	var outputs strings.Builder

	switch cfg.AppType {
	case config.StaticSite:
		outputs.WriteString(`
output "website_url" {
  description = "Website URL"
  value       = "https://${azurerm_storage_account.static_site.primary_web_endpoint}"
}
`)
	case config.NodeJS, config.Flask:
		outputs.WriteString(`
output "public_ip" {
  description = "Public IP address"
  value       = azurerm_public_ip.main.ip_address
}

output "application_url" {
  description = "Application URL"
  value       = "http://${azurerm_public_ip.main.ip_address}"
}
`)
	case config.Docker:
		outputs.WriteString(`
output "container_fqdn" {
  description = "Container FQDN"
  value       = azurerm_container_group.app.fqdn
}

output "application_url" {
  description = "Application URL"
  value       = "http://${azurerm_container_group.app.fqdn}"
}
`)
	}

	return outputs.String()
}
