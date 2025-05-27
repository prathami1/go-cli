package utils

import (
	"fmt"
	"io"
	"time"

	"github.com/schollz/progressbar/v3"
)

// ProgressConfig contains configuration for progress indicators
type ProgressConfig struct {
	Description string
	ShowBytes   bool
	ShowSpeed   bool
	Theme       *progressbar.Theme
}

// NewDownloadProgressBar creates a progress bar optimized for file downloads
func NewDownloadProgressBar(maxBytes int64, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions64(maxBytes,
		progressbar.OptionSetDescription(fmt.Sprintf("📥 %s", description)),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowElapsedTimeOnFinish(),
	)
}

// NewInstallSpinner creates a spinner for installation processes
func NewInstallSpinner(description string) *progressbar.ProgressBar {
	return progressbar.NewOptions(-1,
		progressbar.OptionSetDescription(fmt.Sprintf("🔧 %s", description)),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSpinnerType(11), // Modern dots spinner
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowElapsedTimeOnFinish(),
	)
}

// NewGenericProgressBar creates a customizable progress bar for general use
func NewGenericProgressBar(max int64, config ProgressConfig) *progressbar.ProgressBar {
	options := []progressbar.Option{
		progressbar.OptionSetDescription(config.Description),
		progressbar.OptionSetWidth(50),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowElapsedTimeOnFinish(),
	}

	if config.ShowBytes {
		options = append(options, progressbar.OptionShowBytes(true))
	}

	if config.ShowSpeed {
		options = append(options, progressbar.OptionShowIts())
	}

	if config.Theme != nil {
		options = append(options, progressbar.OptionSetTheme(*config.Theme))
	} else {
		// Default modern theme
		options = append(options, progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	}

	return progressbar.NewOptions64(max, options...)
}

// ProgressWriter wraps an io.Writer with a progress bar
type ProgressWriter struct {
	io.Writer
	ProgressBar *progressbar.ProgressBar
}

// NewProgressWriter creates a new ProgressWriter
func NewProgressWriter(writer io.Writer, progressBar *progressbar.ProgressBar) *ProgressWriter {
	return &ProgressWriter{
		Writer:      writer,
		ProgressBar: progressBar,
	}
}

// Write implements io.Writer and updates the progress bar
func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.Writer.Write(p)
	if err != nil {
		return n, err
	}

	// Update progress bar with bytes written
	pw.ProgressBar.Add(n)
	return n, nil
}

// InstallationStep represents a single step in an installation process
type InstallationStep struct {
	Name        string
	Description string
	Duration    time.Duration
}

// MultiStepInstaller manages progress across multiple installation steps
type MultiStepInstaller struct {
	steps       []InstallationStep
	currentStep int
	spinner     *progressbar.ProgressBar
}

// NewMultiStepInstaller creates a new multi-step installation progress manager
func NewMultiStepInstaller(steps []InstallationStep) *MultiStepInstaller {
	return &MultiStepInstaller{
		steps:       steps,
		currentStep: 0,
	}
}

// StartStep begins a new installation step with a spinner
func (msi *MultiStepInstaller) StartStep(stepIndex int) {
	if stepIndex >= len(msi.steps) {
		return
	}

	step := msi.steps[stepIndex]
	msi.currentStep = stepIndex

	description := fmt.Sprintf("[%d/%d] %s", stepIndex+1, len(msi.steps), step.Description)
	msi.spinner = NewInstallSpinner(description)
}

// UpdateStep updates the current step description
func (msi *MultiStepInstaller) UpdateStep(description string) {
	if msi.spinner != nil {
		stepDesc := fmt.Sprintf("[%d/%d] %s", msi.currentStep+1, len(msi.steps), description)
		msi.spinner.Describe(fmt.Sprintf("🔧 %s", stepDesc))
	}
}

// FinishStep completes the current step
func (msi *MultiStepInstaller) FinishStep() {
	if msi.spinner != nil {
		msi.spinner.Finish()

		// Print completion message
		step := msi.steps[msi.currentStep]
		fmt.Printf("✅ %s completed!\n", step.Name)
	}
}

// Finish completes the entire installation process
func (msi *MultiStepInstaller) Finish() {
	if msi.spinner != nil {
		msi.spinner.Finish()
	}
	fmt.Println("🎉 Installation completed successfully!")
}
