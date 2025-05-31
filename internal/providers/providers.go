package providers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/prathami1/go-cli/internal/config"
	"github.com/prathami1/go-cli/internal/logger"
)

// isBloombergEnvironment detects if we're running in a Bloomberg environment
func isBloombergEnvironment() bool {
	// Check for Bloomberg-specific environment variables or domain
	hostname, _ := os.Hostname()
	username := os.Getenv("USER")
	domain := os.Getenv("USERDOMAIN")

	// Common indicators of Bloomberg environment
	return strings.Contains(strings.ToLower(hostname), "bloomberg") ||
		strings.Contains(strings.ToLower(username), "corp") ||
		strings.Contains(strings.ToLower(domain), "bloomberg") ||
		os.Getenv("BLOOMBERG_ENV") != ""
}

// CheckAuthentication verifies if the user is authenticated with the specified cloud provider
// If not authenticated, it automatically triggers the login process
func CheckAuthentication(provider config.CloudProvider) error {
	logger.Debugf("Checking authentication for %s", provider)

	switch provider {
	case config.AWS:
		return checkAndAutoLoginAWS()
	case config.GCP:
		return checkAndAutoLoginGCP()
	case config.Azure:
		return checkAndAutoLoginAzure()
	default:
		return fmt.Errorf("unsupported cloud provider: %s", provider)
	}
}

// GetRegions returns available regions for the specified cloud provider
func GetRegions(provider config.CloudProvider) ([]string, error) {
	logger.Debugf("Getting regions for %s", provider)

	switch provider {
	case config.AWS:
		return getAWSRegions(), nil
	case config.GCP:
		return getGCPRegions(), nil
	case config.Azure:
		return getAzureRegions(), nil
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", provider)
	}
}

// AWS authentication check with auto-login
func checkAndAutoLoginAWS() error {
	// Check if AWS CLI is installed, if not install it
	if err := checkAndPromptInstall(config.AWS, "aws"); err != nil {
		return fmt.Errorf("AWS CLI installation failed: %w", err)
	}

	// Check if user is authenticated by trying to get caller identity
	cmd := exec.Command("aws", "sts", "get-caller-identity")
	output, err := cmd.Output()
	if err != nil {
		// Not authenticated - trigger automatic login
		logger.Info("🔐 Not authenticated with AWS. Initiating automatic login...")
		return autoLoginAWS()
	}

	if strings.Contains(string(output), "UserId") {
		logger.Debug("AWS authentication verified")
		return nil
	}

	return fmt.Errorf("AWS authentication verification failed")
}

// GCP authentication check with auto-login
func checkAndAutoLoginGCP() error {
	// Check if gcloud CLI is installed, if not install it
	if err := checkAndPromptInstall(config.GCP, "gcloud"); err != nil {
		return fmt.Errorf("Google Cloud CLI installation failed: %w", err)
	}

	// Check if user is authenticated
	cmd := exec.Command("gcloud", "auth", "list", "--filter=status:ACTIVE", "--format=value(account)")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check GCP authentication status")
	}

	if strings.TrimSpace(string(output)) == "" {
		// Not authenticated - trigger automatic login
		logger.Info("🔐 Not authenticated with GCP. Initiating automatic login...")
		return autoLoginGCP()
	}

	logger.Debug("GCP authentication verified")
	return nil
}

// Azure authentication check with auto-login
func checkAndAutoLoginAzure() error {
	// Check if Azure CLI is installed, if not install it
	if err := checkAndPromptInstall(config.Azure, "az"); err != nil {
		return fmt.Errorf("Azure CLI installation failed: %w", err)
	}

	// Check if user is authenticated
	cmd := exec.Command("az", "account", "show")
	_, err := cmd.Output()
	if err != nil {
		// Not authenticated - trigger automatic login
		logger.Info("🔐 Not authenticated with Azure. Initiating automatic login...")
		return autoLoginAzure()
	}

	logger.Debug("Azure authentication verified")
	return nil
}

// autoLoginAWS automatically handles AWS authentication with Bloomberg SSO support
func autoLoginAWS() error {
	logger.Info("🚀 Starting AWS authentication flow...")

	if isBloombergEnvironment() {
		logger.Info("💼 Detected Bloomberg environment. Using enterprise SSO authentication...")
		return authenticateBloombergAWS()
	} else {
		logger.Info("🏢 Using standard AWS authentication...")
		return authenticateStandardAWS()
	}
}

// authenticateBloombergAWS handles Bloomberg-specific AWS authentication
func authenticateBloombergAWS() error {
	logger.Info("🔑 Starting Bloomberg AWS SSO authentication...")
	logger.Info("📱 Please have your B-Unit or B-Unit phone app ready for 2FA")

	// For Bloomberg, try SSO first
	logger.Info("🔐 Attempting AWS SSO login...")
	ssoCmd := exec.Command("aws", "sso", "login")
	ssoCmd.Stdout = os.Stdout
	ssoCmd.Stderr = os.Stderr
	ssoCmd.Stdin = os.Stdin

	if err := ssoCmd.Run(); err != nil {
		logger.Warn("AWS SSO failed, trying alternative authentication...")
		return authenticateStandardAWS()
	}

	logger.Info("✅ Bloomberg AWS SSO authentication successful!")
	return nil
}

// authenticateStandardAWS handles standard AWS authentication
func authenticateStandardAWS() error {
	logger.Info("📝 Using AWS configure for authentication...")
	configCmd := exec.Command("aws", "configure")
	configCmd.Stdout = os.Stdout
	configCmd.Stderr = os.Stderr
	configCmd.Stdin = os.Stdin

	if err := configCmd.Run(); err != nil {
		return fmt.Errorf("AWS authentication failed: %w", err)
	}

	logger.Info("✅ AWS authentication successful!")
	return nil
}

// autoLoginGCP automatically handles GCP authentication with Bloomberg SSO support
func autoLoginGCP() error {
	logger.Info("🚀 Starting GCP authentication flow...")

	if isBloombergEnvironment() {
		logger.Info("💼 Detected Bloomberg environment. Using enterprise authentication...")
		return authenticateBloombergGCP()
	} else {
		logger.Info("🏢 Using standard GCP authentication...")
		return authenticateStandardGCP()
	}
}

// authenticateBloombergGCP handles Bloomberg-specific GCP authentication
func authenticateBloombergGCP() error {
	logger.Info("🔑 Starting Bloomberg GCP authentication...")
	logger.Info("📱 Please have your B-Unit or B-Unit phone app ready for 2FA")

	// Use gcloud auth login with enterprise-friendly options
	loginCmd := exec.Command("gcloud", "auth", "login", "--enable-gdrive-access", "--brief")
	loginCmd.Stdout = os.Stdout
	loginCmd.Stderr = os.Stderr
	loginCmd.Stdin = os.Stdin

	if err := loginCmd.Run(); err != nil {
		return fmt.Errorf("Bloomberg GCP authentication failed. Please ensure you have access to GCP with your CORP credentials: %w", err)
	}

	logger.Info("✅ Bloomberg GCP authentication successful!")
	return nil
}

// authenticateStandardGCP handles standard GCP authentication
func authenticateStandardGCP() error {
	loginCmd := exec.Command("gcloud", "auth", "login")
	loginCmd.Stdout = os.Stdout
	loginCmd.Stderr = os.Stderr
	loginCmd.Stdin = os.Stdin

	if err := loginCmd.Run(); err != nil {
		return fmt.Errorf("GCP authentication failed: %w", err)
	}

	logger.Info("✅ GCP authentication successful!")
	return nil
}

// autoLoginAzure automatically handles Azure authentication with Bloomberg SSO support
func autoLoginAzure() error {
	logger.Info("🚀 Starting Azure authentication flow...")

	if isBloombergEnvironment() {
		logger.Info("💼 Detected Bloomberg environment. Using enterprise SSO authentication...")
		return authenticateBloombergAzure()
	} else {
		logger.Info("🏢 Using standard Azure authentication...")
		return authenticateStandardAzure()
	}
}

// authenticateBloombergAzure handles Bloomberg-specific Azure authentication
func authenticateBloombergAzure() error {
	logger.Info("🔑 Starting Bloomberg Azure SSO authentication...")
	logger.Info("📱 Please have your B-Unit or B-Unit phone app ready for 2FA")

	// For Bloomberg, try device code flow first as it works best with enterprise SSO
	logger.Info("🖥️  Using device code flow for enterprise compatibility...")

	deviceCmd := exec.Command("az", "login", "--use-device-code", "--tenant", "common")
	deviceCmd.Stdout = os.Stdout
	deviceCmd.Stderr = os.Stderr
	deviceCmd.Stdin = os.Stdin

	if err := deviceCmd.Run(); err != nil {
		logger.Warn("Device code authentication failed, trying alternative methods...")

		// Try with specific tenant if device code fails
		logger.Info("🏢 Attempting login with enterprise tenant...")
		tenantCmd := exec.Command("az", "login", "--allow-no-subscriptions")
		tenantCmd.Stdout = os.Stdout
		tenantCmd.Stderr = os.Stderr
		tenantCmd.Stdin = os.Stdin

		if err := tenantCmd.Run(); err != nil {
			return fmt.Errorf("Bloomberg Azure authentication failed. Please ensure you have access to Azure with your CORP credentials: %w", err)
		}
	}

	// Verify authentication worked
	verifyCmd := exec.Command("az", "account", "show")
	if _, err := verifyCmd.Output(); err != nil {
		return fmt.Errorf("Azure authentication verification failed")
	}

	logger.Info("✅ Bloomberg Azure authentication successful!")
	return nil
}

// authenticateStandardAzure handles standard Azure authentication
func authenticateStandardAzure() error {
	// Standard browser-based authentication
	loginCmd := exec.Command("az", "login")
	loginCmd.Stdout = os.Stdout
	loginCmd.Stderr = os.Stderr
	loginCmd.Stdin = os.Stdin

	if err := loginCmd.Run(); err != nil {
		return fmt.Errorf("Azure authentication failed: %w", err)
	}

	logger.Info("✅ Azure authentication successful!")
	return nil
}

// commandExists checks if a command is available in the system PATH
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// getAWSRegions returns common AWS regions
func getAWSRegions() []string {
	return []string{
		"us-east-1",      // N. Virginia
		"us-east-2",      // Ohio
		"us-west-1",      // N. California
		"us-west-2",      // Oregon
		"eu-west-1",      // Ireland
		"eu-west-2",      // London
		"eu-west-3",      // Paris
		"eu-central-1",   // Frankfurt
		"ap-southeast-1", // Singapore
		"ap-southeast-2", // Sydney
		"ap-northeast-1", // Tokyo
		"ap-northeast-2", // Seoul
		"ap-south-1",     // Mumbai
		"ca-central-1",   // Canada
		"sa-east-1",      // São Paulo
	}
}

// getGCPRegions returns common GCP regions
func getGCPRegions() []string {
	return []string{
		"us-central1",          // Iowa
		"us-east1",             // South Carolina
		"us-east4",             // Northern Virginia
		"us-west1",             // Oregon
		"us-west2",             // Los Angeles
		"us-west3",             // Salt Lake City
		"us-west4",             // Las Vegas
		"europe-west1",         // Belgium
		"europe-west2",         // London
		"europe-west3",         // Frankfurt
		"europe-west4",         // Netherlands
		"europe-west6",         // Zurich
		"asia-east1",           // Taiwan
		"asia-northeast1",      // Tokyo
		"asia-northeast2",      // Osaka
		"asia-northeast3",      // Seoul
		"asia-south1",          // Mumbai
		"asia-southeast1",      // Singapore
		"asia-southeast2",      // Jakarta
		"australia-southeast1", // Sydney
	}
}

// getAzureRegions returns common Azure regions
func getAzureRegions() []string {
	return []string{
		"eastus",             // East US
		"eastus2",            // East US 2
		"westus",             // West US
		"westus2",            // West US 2
		"westus3",            // West US 3
		"centralus",          // Central US
		"northcentralus",     // North Central US
		"southcentralus",     // South Central US
		"westcentralus",      // West Central US
		"canadacentral",      // Canada Central
		"canadaeast",         // Canada East
		"brazilsouth",        // Brazil South
		"northeurope",        // North Europe
		"westeurope",         // West Europe
		"ukwest",             // UK West
		"uksouth",            // UK South
		"francecentral",      // France Central
		"germanywestcentral", // Germany West Central
		"norwayeast",         // Norway East
		"switzerlandnorth",   // Switzerland North
		"eastasia",           // East Asia
		"southeastasia",      // Southeast Asia
		"japaneast",          // Japan East
		"japanwest",          // Japan West
		"australiaeast",      // Australia East
		"australiasoutheast", // Australia Southeast
		"koreacentral",       // Korea Central
		"koreasouth",         // Korea South
		"southindia",         // South India
		"westindia",          // West India
		"centralindia",       // Central India
	}
}
