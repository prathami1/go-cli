package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AppType represents the type of application to deploy
type AppType string

const (
	StaticSite AppType = "static-site"
	NodeJS     AppType = "nodejs"
	Flask      AppType = "flask"
	Docker     AppType = "docker"
)

// CloudProvider represents the target cloud provider
type CloudProvider string

const (
	AWS   CloudProvider = "aws"
	GCP   CloudProvider = "gcp"
	Azure CloudProvider = "azure"
)

// DeploymentConfig represents the configuration for a deployment
type DeploymentConfig struct {
	ProjectName    string        `json:"project_name"`
	AppType        AppType       `json:"app_type"`
	CloudProvider  CloudProvider `json:"cloud_provider"`
	Region         string        `json:"region"`
	ImageName      string        `json:"image_name,omitempty"`
	Services       Services      `json:"services"`
	CreatedAt      string        `json:"created_at"`
	LastDeployment string        `json:"last_deployment,omitempty"`
}

// Services represents optional services to include in the deployment
type Services struct {
	Database     bool `json:"database"`
	Storage      bool `json:"storage"`
	LoadBalancer bool `json:"load_balancer"`
}

// Config holds the global configuration
type Config struct {
	ConfigFile string
	Workdir    string
}

var GlobalConfig *Config

// Initialize sets up the global configuration
func Initialize() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	GlobalConfig = &Config{
		ConfigFile: filepath.Join(home, ".clouddeploy.json"),
		Workdir:    ".",
	}
}

// LoadConfig loads the deployment configuration from file
func LoadConfig() (*DeploymentConfig, error) {
	configPath := ".clouddeploy.json"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no configuration file found. Run 'clouddeploy init' first")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config DeploymentConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the deployment configuration to file
func SaveConfig(config *DeploymentConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := ".clouddeploy.json"
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ValidateAppType checks if the app type is valid
func ValidateAppType(appType string) bool {
	validTypes := []AppType{StaticSite, NodeJS, Flask, Docker}
	for _, t := range validTypes {
		if AppType(appType) == t {
			return true
		}
	}
	return false
}

// ValidateCloudProvider checks if the cloud provider is valid
func ValidateCloudProvider(provider string) bool {
	validProviders := []CloudProvider{AWS, GCP, Azure}
	for _, p := range validProviders {
		if CloudProvider(provider) == p {
			return true
		}
	}
	return false
}

// GetAppTypes returns all available app types
func GetAppTypes() []string {
	return []string{string(StaticSite), string(NodeJS), string(Flask), string(Docker)}
}

// GetCloudProviders returns all available cloud providers
func GetCloudProviders() []string {
	return []string{string(AWS), string(GCP), string(Azure)}
}
