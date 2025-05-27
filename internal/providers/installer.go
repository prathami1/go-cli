package providers

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/prathami1/go-cli/internal/config"
	"github.com/prathami1/go-cli/internal/logger"
)

// installCLI automatically installs the required CLI tool for the specified cloud provider
func installCLI(provider config.CloudProvider) error {
	logger.Infof("🔧 Installing %s CLI automatically...", provider)

	switch provider {
	case config.AWS:
		return installAWSCLI()
	case config.GCP:
		return installGoogleCloudCLI()
	case config.Azure:
		return installAzureCLI()
	default:
		return fmt.Errorf("unsupported cloud provider: %s", provider)
	}
}

// installAWSCLI installs AWS CLI based on the operating system
func installAWSCLI() error {
	switch runtime.GOOS {
	case "windows":
		return installAWSCLIWindows()
	case "darwin":
		return installAWSCLIMacOS()
	case "linux":
		return installAWSCLILinux()
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// installAWSCLIWindows installs AWS CLI on Windows using MSI
func installAWSCLIWindows() error {
	logger.Info("Installing AWS CLI on Windows...")

	// Download the MSI installer
	msiURL := "https://awscli.amazonaws.com/AWSCLIV2.msi"
	msiPath := filepath.Join(os.TempDir(), "AWSCLIV2.msi")

	if err := downloadFile(msiURL, msiPath); err != nil {
		return fmt.Errorf("failed to download AWS CLI installer: %w", err)
	}
	defer os.Remove(msiPath)

	// Install using msiexec with quiet mode
	cmd := exec.Command("msiexec.exe", "/i", msiPath, "/quiet", "/norestart")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install AWS CLI: %w", err)
	}

	logger.Info("✅ AWS CLI installed successfully!")
	return nil
}

// installAWSCLIMacOS installs AWS CLI on macOS using PKG installer
func installAWSCLIMacOS() error {
	logger.Info("Installing AWS CLI on macOS...")

	// Download the PKG installer
	pkgURL := "https://awscli.amazonaws.com/AWSCLIV2.pkg"
	pkgPath := filepath.Join(os.TempDir(), "AWSCLIV2.pkg")

	if err := downloadFile(pkgURL, pkgPath); err != nil {
		return fmt.Errorf("failed to download AWS CLI installer: %w", err)
	}
	defer os.Remove(pkgPath)

	// Install using installer command
	cmd := exec.Command("sudo", "installer", "-pkg", pkgPath, "-target", "/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install AWS CLI: %w", err)
	}

	logger.Info("✅ AWS CLI installed successfully!")
	return nil
}

// installAWSCLILinux installs AWS CLI on Linux using the universal method
func installAWSCLILinux() error {
	logger.Info("Installing AWS CLI on Linux...")

	// Determine architecture
	arch := "x86_64"
	if runtime.GOARCH == "arm64" {
		arch = "aarch64"
	} else if runtime.GOARCH == "arm" {
		arch = "aarch64" // AWS CLI uses aarch64 for both arm and arm64
	}

	// Download and extract AWS CLI
	zipURL := fmt.Sprintf("https://awscli.amazonaws.com/awscli-exe-linux-%s.zip", arch)
	zipPath := filepath.Join(os.TempDir(), "awscliv2.zip")
	extractDir := filepath.Join(os.TempDir(), "aws-cli-install")

	logger.Infof("Downloading AWS CLI from %s...", zipURL)
	if err := downloadFile(zipURL, zipPath); err != nil {
		return fmt.Errorf("failed to download AWS CLI: %w", err)
	}
	defer os.Remove(zipPath)
	defer os.RemoveAll(extractDir)

	logger.Info("Extracting AWS CLI...")
	// Extract the zip file
	if err := extractZip(zipPath, extractDir); err != nil {
		return fmt.Errorf("failed to extract AWS CLI: %w", err)
	}

	logger.Info("Installing AWS CLI...")
	// Install AWS CLI
	installScript := filepath.Join(extractDir, "aws", "install")

	// Make the install script executable
	if err := os.Chmod(installScript, 0755); err != nil {
		return fmt.Errorf("failed to make install script executable: %w", err)
	}

	cmd := exec.Command("sudo", installScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install AWS CLI: %w", err)
	}

	logger.Info("✅ AWS CLI installed successfully!")
	return nil
}

// installGoogleCloudCLI installs Google Cloud CLI based on the operating system
func installGoogleCloudCLI() error {
	switch runtime.GOOS {
	case "windows":
		return installGoogleCloudCLIWindows()
	case "darwin":
		return installGoogleCloudCLIMacOS()
	case "linux":
		return installGoogleCloudCLILinux()
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// installGoogleCloudCLIWindows installs Google Cloud CLI on Windows
func installGoogleCloudCLIWindows() error {
	logger.Info("Installing Google Cloud CLI on Windows...")

	// Download the installer
	installerURL := "https://dl.google.com/dl/cloudsdk/channels/rapid/GoogleCloudSDKInstaller.exe"
	installerPath := filepath.Join(os.TempDir(), "GoogleCloudSDKInstaller.exe")

	if err := downloadFile(installerURL, installerPath); err != nil {
		return fmt.Errorf("failed to download Google Cloud CLI installer: %w", err)
	}
	defer os.Remove(installerPath)

	// Install silently
	cmd := exec.Command(installerPath, "/S")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Google Cloud CLI: %w", err)
	}

	logger.Info("✅ Google Cloud CLI installed successfully!")
	return nil
}

// installGoogleCloudCLIMacOS installs Google Cloud CLI on macOS
func installGoogleCloudCLIMacOS() error {
	logger.Info("Installing Google Cloud CLI on macOS...")

	// Use the install script method for macOS
	cmd := exec.Command("bash", "-c", "curl https://sdk.cloud.google.com | bash -s -- --disable-prompts")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Google Cloud CLI: %w", err)
	}

	logger.Info("✅ Google Cloud CLI installed successfully!")
	return nil
}

// installGoogleCloudCLILinux installs Google Cloud CLI on Linux
func installGoogleCloudCLILinux() error {
	logger.Info("Installing Google Cloud CLI on Linux...")

	// Check if we can use package manager first
	if commandExists("apt-get") {
		return installGoogleCloudCLIDebian()
	} else if commandExists("dnf") || commandExists("yum") {
		return installGoogleCloudCLIRedHat()
	}

	// Fall back to universal install script
	cmd := exec.Command("bash", "-c", "curl https://sdk.cloud.google.com | bash -s -- --disable-prompts")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Google Cloud CLI: %w", err)
	}

	logger.Info("✅ Google Cloud CLI installed successfully!")
	return nil
}

// installGoogleCloudCLIDebian installs Google Cloud CLI on Debian/Ubuntu
func installGoogleCloudCLIDebian() error {
	logger.Info("Installing Google Cloud CLI using apt package manager...")

	commands := [][]string{
		{"apt-get", "update"},
		{"apt-get", "install", "-y", "apt-transport-https", "ca-certificates", "gnupg", "curl"},
		{"bash", "-c", "curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg"},
		{"bash", "-c", `echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list`},
		{"apt-get", "update"},
		{"apt-get", "install", "-y", "google-cloud-cli"},
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command("sudo", cmdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run command %v: %w", cmdArgs, err)
		}
	}

	return nil
}

// installGoogleCloudCLIRedHat installs Google Cloud CLI on RedHat/CentOS/Fedora
func installGoogleCloudCLIRedHat() error {
	logger.Info("Installing Google Cloud CLI using dnf/yum package manager...")

	packageManager := "dnf"
	if !commandExists("dnf") {
		packageManager = "yum"
	}

	commands := [][]string{
		{"bash", "-c", `sudo tee -a /etc/yum.repos.d/google-cloud-sdk.repo << EOM
[google-cloud-cli]
name=Google Cloud CLI
baseurl=https://packages.cloud.google.com/yum/repos/cloud-sdk-el9-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=0
gpgkey=https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
EOM`},
		{packageManager, "install", "-y", "google-cloud-cli"},
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command("sudo", cmdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run command %v: %w", cmdArgs, err)
		}
	}

	return nil
}

// installAzureCLI installs Azure CLI based on the operating system
func installAzureCLI() error {
	switch runtime.GOOS {
	case "windows":
		return installAzureCLIWindows()
	case "darwin":
		return installAzureCLIMacOS()
	case "linux":
		return installAzureCLILinux()
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// installAzureCLIWindows installs Azure CLI on Windows
func installAzureCLIWindows() error {
	logger.Info("Installing Azure CLI on Windows...")

	// Try using winget first, fall back to MSI
	if commandExists("winget") {
		cmd := exec.Command("winget", "install", "--exact", "--id", "Microsoft.AzureCLI", "--silent")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			logger.Info("✅ Azure CLI installed successfully!")
			return nil
		}
		logger.Debug("winget failed, falling back to MSI installer")
	}

	// Fall back to MSI installer
	msiURL := "https://aka.ms/installazurecliwindows"
	msiPath := filepath.Join(os.TempDir(), "AzureCLI.msi")

	if err := downloadFile(msiURL, msiPath); err != nil {
		return fmt.Errorf("failed to download Azure CLI installer: %w", err)
	}
	defer os.Remove(msiPath)

	cmd := exec.Command("msiexec.exe", "/i", msiPath, "/quiet", "/norestart")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Azure CLI: %w", err)
	}

	logger.Info("✅ Azure CLI installed successfully!")
	return nil
}

// installAzureCLIMacOS installs Azure CLI on macOS
func installAzureCLIMacOS() error {
	logger.Info("Installing Azure CLI on macOS...")

	// Try using Homebrew first
	if commandExists("brew") {
		cmd := exec.Command("brew", "install", "azure-cli")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err == nil {
			logger.Info("✅ Azure CLI installed successfully!")
			return nil
		}
		logger.Debug("Homebrew failed, falling back to install script")
	}

	// Fall back to install script
	cmd := exec.Command("bash", "-c", "curl -L https://aka.ms/InstallAzureCli | bash")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Azure CLI: %w", err)
	}

	logger.Info("✅ Azure CLI installed successfully!")
	return nil
}

// installAzureCLILinux installs Azure CLI on Linux
func installAzureCLILinux() error {
	logger.Info("Installing Azure CLI on Linux...")

	// Check if we can use package manager first
	if commandExists("apt-get") {
		return installAzureCLIDebian()
	} else if commandExists("dnf") || commandExists("yum") {
		return installAzureCLIRedHat()
	}

	// Fall back to universal install script
	cmd := exec.Command("bash", "-c", "curl -L https://aka.ms/InstallAzureCLIDeb | sudo bash")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Azure CLI: %w", err)
	}

	logger.Info("✅ Azure CLI installed successfully!")
	return nil
}

// installAzureCLIDebian installs Azure CLI on Debian/Ubuntu
func installAzureCLIDebian() error {
	logger.Info("Installing Azure CLI using apt package manager...")

	commands := [][]string{
		{"apt-get", "update"},
		{"apt-get", "install", "-y", "apt-transport-https", "ca-certificates", "curl", "gnupg", "lsb-release"},
		{"bash", "-c", "sudo mkdir -p /etc/apt/keyrings"},
		{"bash", "-c", "curl -sLS https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor | sudo tee /etc/apt/keyrings/microsoft.gpg > /dev/null"},
		{"bash", "-c", "sudo chmod go+r /etc/apt/keyrings/microsoft.gpg"},
		{"bash", "-c", `AZ_DIST=$(lsb_release -cs) && echo "Types: deb
URIs: https://packages.microsoft.com/repos/azure-cli/
Suites: ${AZ_DIST}
Components: main
Architectures: $(dpkg --print-architecture)
Signed-by: /etc/apt/keyrings/microsoft.gpg" | sudo tee /etc/apt/sources.list.d/azure-cli.sources`},
		{"apt-get", "update"},
		{"apt-get", "install", "-y", "azure-cli"},
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command("sudo", cmdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run command %v: %w", cmdArgs, err)
		}
	}

	return nil
}

// installAzureCLIRedHat installs Azure CLI on RedHat/CentOS/Fedora
func installAzureCLIRedHat() error {
	logger.Info("Installing Azure CLI using dnf/yum package manager...")

	packageManager := "dnf"
	if !commandExists("dnf") {
		packageManager = "yum"
	}

	commands := [][]string{
		{"rpm", "--import", "https://packages.microsoft.com/keys/microsoft.asc"},
		{packageManager, "install", "-y", "https://packages.microsoft.com/config/rhel/9.0/packages-microsoft-prod.rpm"},
		{packageManager, "install", "-y", "azure-cli"},
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command("sudo", cmdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run command %v: %w", cmdArgs, err)
		}
	}

	return nil
}

// downloadFile downloads a file from URL to local path
func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractZip extracts a zip file to destination directory using Go's built-in zip library
func extractZip(src, dest string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Create destination directory
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	// Extract files
	for _, file := range reader.File {
		path := filepath.Join(dest, file.Name)

		// Check for ZipSlip vulnerability
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.FileInfo().Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		_, err = io.Copy(targetFile, fileReader)
		if err != nil {
			return err
		}
	}

	return nil
}

// checkAndPromptInstall checks if CLI tool exists and prompts for installation
func checkAndPromptInstall(provider config.CloudProvider, cmdName string) error {
	if commandExists(cmdName) {
		return nil // CLI already exists
	}

	logger.Warnf("%s CLI is not installed", provider)
	logger.Infof("CloudDeploy can automatically install %s CLI for you", provider)

	// Prompt user for installation
	if !promptYesNo(fmt.Sprintf("Would you like to install %s CLI now? (y/N)", provider)) {
		return fmt.Errorf("%s CLI is required but not installed", provider)
	}

	return installCLI(provider)
}

// promptYesNo prompts the user for a yes/no answer
func promptYesNo(message string) bool {
	fmt.Printf("%s ", message)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		logger.Errorf("Failed to read user input: %v", err)
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
