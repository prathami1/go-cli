package providers

import (
	"fmt"
	"os"
	"time"

	"github.com/prathami1/go-cli/internal/utils"
)

// DemoProgressBars demonstrates the various progress indicators available
func DemoProgressBars() {
	fmt.Println("🎯 CloudDeploy CLI Progress Indicators Demo")
	fmt.Println("============================================")

	// Demo 1: Download Progress Bar
	fmt.Println("\n📥 Demo 1: Download Progress Bar")
	downloadBar := utils.NewDownloadProgressBar(1024*1024*10, "Demo file") // 10MB demo
	for i := 0; i < 100; i++ {
		downloadBar.Add(1024 * 100) // Add 100KB each iteration
		time.Sleep(50 * time.Millisecond)
	}
	downloadBar.Finish()

	// Demo 2: Installation Spinner
	fmt.Println("\n🔧 Demo 2: Installation Spinner")
	spinner := utils.NewInstallSpinner("Installing demo package...")
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
	}
	spinner.Finish()
	fmt.Println("✅ Demo package installed!")

	// Demo 3: Multi-step Installation
	fmt.Println("\n🚀 Demo 3: Multi-step Installation Process")
	steps := []utils.InstallationStep{
		{Name: "Download", Description: "Downloading package"},
		{Name: "Extract", Description: "Extracting files"},
		{Name: "Configure", Description: "Configuring package"},
		{Name: "Install", Description: "Installing package"},
	}

	installer := utils.NewMultiStepInstaller(steps)

	for i := range steps {
		installer.StartStep(i)

		// Simulate different durations for each step
		duration := time.Duration(2+i) * time.Second
		start := time.Now()

		for time.Since(start) < duration {
			time.Sleep(200 * time.Millisecond)
		}

		installer.FinishStep()
	}

	installer.Finish()

	fmt.Println("\n🎉 Demo completed! These progress indicators will be shown during real CLI installations.")
}

// DemoDownloadProgress demonstrates downloading with progress
func DemoDownloadProgress() {
	fmt.Println("📥 Testing download progress with a real file...")

	// Create a temporary file for demo
	tempFile, err := os.CreateTemp("", "demo-download-*.txt")
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		return
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Demo downloading a small file from a reliable source
	url := "https://httpbin.org/bytes/1048576" // 1MB of random data

	fmt.Printf("Downloading from: %s\n", url)
	err = downloadFileWithProgress(url, tempFile.Name(), "Test file (1MB)")
	if err != nil {
		fmt.Printf("Download failed: %v\n", err)
		return
	}

	fmt.Println("✅ Download completed successfully!")

	// Check file size
	info, err := os.Stat(tempFile.Name())
	if err == nil {
		fmt.Printf("Downloaded file size: %d bytes\n", info.Size())
	}
}
