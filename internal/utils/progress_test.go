package utils

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/schollz/progressbar/v3"
)

func TestNewDownloadProgressBar(t *testing.T) {
	// Test that progress bar is created successfully
	bar := NewDownloadProgressBar(1024, "test file")
	if bar == nil {
		t.Error("NewDownloadProgressBar returned nil")
	}

	// Test that description is set correctly
	if !strings.Contains(bar.String(), "📥 test file") {
		t.Error("Progress bar doesn't contain expected description")
	}
}

func TestNewInstallSpinner(t *testing.T) {
	// Test that spinner is created successfully
	spinner := NewInstallSpinner("test installation")
	if spinner == nil {
		t.Error("NewInstallSpinner returned nil")
	}

	// Test that description is set correctly
	if !strings.Contains(spinner.String(), "🔧 test installation") {
		t.Error("Spinner doesn't contain expected description")
	}
}

func TestNewGenericProgressBar(t *testing.T) {
	config := ProgressConfig{
		Description: "test progress",
		ShowBytes:   true,
		ShowSpeed:   true,
	}

	bar := NewGenericProgressBar(100, config)
	if bar == nil {
		t.Error("NewGenericProgressBar returned nil")
	}
}

func TestProgressWriter(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Create a progress bar that writes to our buffer
	bar := progressbar.NewOptions(100,
		progressbar.OptionSetWriter(&buf),
		progressbar.OptionSetDescription("test"),
		progressbar.OptionSetWidth(10),
	)

	// Create a progress writer
	pw := NewProgressWriter(&buf, bar)

	// Write some data
	testData := []byte("hello world")
	n, err := pw.Write(testData)

	if err != nil {
		t.Errorf("ProgressWriter.Write() error = %v", err)
	}

	if n != len(testData) {
		t.Errorf("ProgressWriter.Write() wrote %d bytes, expected %d", n, len(testData))
	}

	// Verify that the progress bar was updated
	if bar.State().CurrentNum != int64(len(testData)) {
		t.Errorf("Progress bar current value = %d, expected %d", bar.State().CurrentNum, len(testData))
	}
}

func TestMultiStepInstaller(t *testing.T) {
	steps := []InstallationStep{
		{Name: "Step 1", Description: "First step"},
		{Name: "Step 2", Description: "Second step"},
	}

	installer := NewMultiStepInstaller(steps)
	if installer == nil {
		t.Error("NewMultiStepInstaller returned nil")
	}

	if len(installer.steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(installer.steps))
	}

	// Test starting a step
	installer.StartStep(0)
	if installer.currentStep != 0 {
		t.Errorf("Expected currentStep to be 0, got %d", installer.currentStep)
	}

	// Test finishing a step
	installer.FinishStep()

	// Test updating step description
	installer.UpdateStep("Updated description")

	// Test finishing installation
	installer.Finish()
}

func TestInstallationStep(t *testing.T) {
	step := InstallationStep{
		Name:        "Test Step",
		Description: "Test Description",
		Duration:    5 * time.Second,
	}

	if step.Name != "Test Step" {
		t.Errorf("Expected name 'Test Step', got '%s'", step.Name)
	}

	if step.Description != "Test Description" {
		t.Errorf("Expected description 'Test Description', got '%s'", step.Description)
	}

	if step.Duration != 5*time.Second {
		t.Errorf("Expected duration 5s, got %v", step.Duration)
	}
}

func TestProgressConfig(t *testing.T) {
	config := ProgressConfig{
		Description: "Test Config",
		ShowBytes:   true,
		ShowSpeed:   false,
		Theme:       &progressbar.Theme{},
	}

	if config.Description != "Test Config" {
		t.Errorf("Expected description 'Test Config', got '%s'", config.Description)
	}

	if !config.ShowBytes {
		t.Error("Expected ShowBytes to be true")
	}

	if config.ShowSpeed {
		t.Error("Expected ShowSpeed to be false")
	}

	if config.Theme == nil {
		t.Error("Expected Theme to be set")
	}
}
