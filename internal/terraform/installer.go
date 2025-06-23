package terraform

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

	"github.com/prathami1/go-cli/internal/logger"
	"github.com/prathami1/go-cli/internal/utils"
)

// ensureTerraformInstalled checks if Terraform is installed and prompts for installation if not
func ensureTerraformInstalled() error {
	if commandExists("terraform") {
		return nil // Terraform already exists
	}

	logger.Warn("Terraform is not installed")
	logger.Info("CloudDeploy can automatically install Terraform for you")

	// Prompt user for installation
	if !promptYesNo("Would you like to install Terraform now? (y/N)") {
		return fmt.Errorf("Terraform is required but not installed. Please install Terraform manually from https://www.terraform.io/downloads")
	}

	return installTerraform()
}

// commandExists checks if a command is available in the system PATH
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
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

// installTerraform installs Terraform based on the operating system
func installTerraform() error {
	logger.Info("🔧 Installing Terraform automatically...")

	switch runtime.GOOS {
	case "windows":
		return installTerraformWindows()
	case "darwin":
		return installTerraformMacOS()
	case "linux":
		return installTerraformLinux()
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// installTerraformWindows installs Terraform on Windows
func installTerraformWindows() error {
	steps := []utils.InstallationStep{
		{Name: "Download", Description: "Downloading Terraform"},
		{Name: "Extract", Description: "Extracting Terraform"},
		{Name: "Install", Description: "Installing Terraform to PATH"},
	}
	installer := utils.NewMultiStepInstaller(steps)

	// Step 1: Download
	installer.StartStep(0)

	// Determine architecture
	arch := "amd64"
	if runtime.GOARCH == "arm64" {
		arch = "arm64"
	}

	version := "1.6.6" // Latest stable version
	zipURL := fmt.Sprintf("https://releases.hashicorp.com/terraform/%s/terraform_%s_windows_%s.zip", version, version, arch)
	zipPath := filepath.Join(os.TempDir(), "terraform.zip")

	if err := downloadFileWithProgress(zipURL, zipPath, "Terraform"); err != nil {
		return fmt.Errorf("failed to download Terraform: %w", err)
	}
	defer os.Remove(zipPath)
	installer.FinishStep()

	// Step 2: Extract
	installer.StartStep(1)
	installer.UpdateStep("Extracting Terraform binary...")

	// Create installation directory
	installDir := filepath.Join(os.Getenv("PROGRAMFILES"), "Terraform")
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory: %w", err)
	}

	if err := extractZip(zipPath, installDir); err != nil {
		return fmt.Errorf("failed to extract Terraform: %w", err)
	}
	installer.FinishStep()

	// Step 3: Add to PATH
	installer.StartStep(2)
	installer.UpdateStep("Adding Terraform to PATH...")

	// Add to PATH (requires admin privileges or user environment)
	logger.Info("Please add %s to your PATH environment variable", installDir)
	logger.Info("You may need to restart your terminal after installation")

	installer.FinishStep()
	installer.Finish()

	logger.Info("✅ Terraform installed successfully!")
	logger.Info("💡 You may need to restart your terminal to use terraform command")
	return nil
}

// installTerraformMacOS installs Terraform on macOS
func installTerraformMacOS() error {
	// Try using Homebrew first if available
	if commandExists("brew") {
		spinner := utils.NewInstallSpinner("Installing Terraform via Homebrew...")
		cmd := exec.Command("brew", "install", "terraform")
		if err := cmd.Run(); err == nil {
			spinner.Finish()
			logger.Info("✅ Terraform installed successfully via Homebrew!")
			return nil
		}
		spinner.Finish()
		logger.Debug("Homebrew failed, falling back to manual installation")
	}

	// Fall back to manual installation
	steps := []utils.InstallationStep{
		{Name: "Download", Description: "Downloading Terraform"},
		{Name: "Extract", Description: "Extracting Terraform"},
		{Name: "Install", Description: "Installing Terraform to /usr/local/bin"},
	}
	installer := utils.NewMultiStepInstaller(steps)

	// Step 1: Download
	installer.StartStep(0)

	// Determine architecture
	arch := "amd64"
	if runtime.GOARCH == "arm64" {
		arch = "arm64"
	}

	version := "1.6.6" // Latest stable version
	zipURL := fmt.Sprintf("https://releases.hashicorp.com/terraform/%s/terraform_%s_darwin_%s.zip", version, version, arch)
	zipPath := filepath.Join(os.TempDir(), "terraform.zip")

	if err := downloadFileWithProgress(zipURL, zipPath, "Terraform"); err != nil {
		return fmt.Errorf("failed to download Terraform: %w", err)
	}
	defer os.Remove(zipPath)
	installer.FinishStep()

	// Step 2: Extract
	installer.StartStep(1)
	installer.UpdateStep("Extracting Terraform binary...")

	extractDir := filepath.Join(os.TempDir(), "terraform-extract")
	if err := extractZip(zipPath, extractDir); err != nil {
		return fmt.Errorf("failed to extract Terraform: %w", err)
	}
	defer os.RemoveAll(extractDir)
	installer.FinishStep()

	// Step 3: Install to /usr/local/bin
	installer.StartStep(2)
	installer.UpdateStep("Installing to /usr/local/bin (requires sudo)...")

	terraformBinary := filepath.Join(extractDir, "terraform")
	installPath := "/usr/local/bin/terraform"

	// Copy with sudo
	cmd := exec.Command("sudo", "cp", terraformBinary, installPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Terraform: %w", err)
	}

	// Make executable
	cmd = exec.Command("sudo", "chmod", "+x", installPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to make Terraform executable: %w", err)
	}

	installer.FinishStep()
	installer.Finish()

	logger.Info("✅ Terraform installed successfully!")
	return nil
}

// installTerraformLinux installs Terraform on Linux
func installTerraformLinux() error {
	steps := []utils.InstallationStep{
		{Name: "Download", Description: "Downloading Terraform"},
		{Name: "Extract", Description: "Extracting Terraform"},
		{Name: "Install", Description: "Installing Terraform to /usr/local/bin"},
	}
	installer := utils.NewMultiStepInstaller(steps)

	// Step 1: Download
	installer.StartStep(0)

	// Determine architecture
	arch := "amd64"
	if runtime.GOARCH == "arm64" {
		arch = "arm64"
	} else if runtime.GOARCH == "arm" {
		arch = "arm"
	}

	version := "1.6.6" // Latest stable version
	zipURL := fmt.Sprintf("https://releases.hashicorp.com/terraform/%s/terraform_%s_linux_%s.zip", version, version, arch)
	zipPath := filepath.Join(os.TempDir(), "terraform.zip")

	if err := downloadFileWithProgress(zipURL, zipPath, "Terraform"); err != nil {
		return fmt.Errorf("failed to download Terraform: %w", err)
	}
	defer os.Remove(zipPath)
	installer.FinishStep()

	// Step 2: Extract
	installer.StartStep(1)
	installer.UpdateStep("Extracting Terraform binary...")

	extractDir := filepath.Join(os.TempDir(), "terraform-extract")
	if err := extractZip(zipPath, extractDir); err != nil {
		return fmt.Errorf("failed to extract Terraform: %w", err)
	}
	defer os.RemoveAll(extractDir)
	installer.FinishStep()

	// Step 3: Install to /usr/local/bin
	installer.StartStep(2)
	installer.UpdateStep("Installing to /usr/local/bin (requires sudo)...")

	terraformBinary := filepath.Join(extractDir, "terraform")
	installPath := "/usr/local/bin/terraform"

	// Copy with sudo
	cmd := exec.Command("sudo", "cp", terraformBinary, installPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Terraform: %w", err)
	}

	// Make executable
	cmd = exec.Command("sudo", "chmod", "+x", installPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to make Terraform executable: %w", err)
	}

	installer.FinishStep()
	installer.Finish()

	logger.Info("✅ Terraform installed successfully!")
	return nil
}

// downloadFileWithProgress downloads a file with a beautiful progress bar
func downloadFileWithProgress(url, filepath, description string) error {
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

	// Get the file size for progress tracking
	size := resp.ContentLength

	if size > 0 {
		// Use progress bar if size is known
		progressReader := &progressReader{
			Reader:      resp.Body,
			total:       size,
			description: description,
		}
		_, err = io.Copy(out, progressReader)
	} else {
		// Simple copy if size unknown
		_, err = io.Copy(out, resp.Body)
	}

	return err
}

// progressReader implements io.Reader with progress tracking
type progressReader struct {
	io.Reader
	total       int64
	read        int64
	description string
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.read += int64(n)

	if pr.total > 0 {
		percentage := float64(pr.read) / float64(pr.total) * 100
		fmt.Printf("\r📥 Downloading %s: %.1f%% (%d/%d bytes)",
			pr.description, percentage, pr.read, pr.total)
	}

	if err != nil {
		fmt.Println() // New line after progress
	}

	return n, err
}

// extractZip extracts a zip file to a destination directory
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
