package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/prathami1/go-cli/internal/config"
	"github.com/prathami1/go-cli/internal/logger"
)

// DetectionResult contains the analysis results
type DetectionResult struct {
	AppType    config.AppType `json:"app_type"`
	Confidence string         `json:"confidence"` // "high", "medium", "low"
	Indicators []string       `json:"indicators"` // Files that influenced the decision
	Path       string         `json:"path"`       // Directory analyzed
}

// AnalyzeProject analyzes the current directory to detect the project type
func AnalyzeProject(projectPath string) (*DetectionResult, error) {
	if projectPath == "" {
		projectPath = "."
	}

	logger.Debugf("Analyzing project at: %s", projectPath)

	// Get list of files in the project directory
	files, err := getProjectFiles(projectPath)
	if err != nil {
		return nil, err
	}

	logger.Debugf("Found %d files to analyze", len(files))

	// Analyze files and detect project type
	result := analyzeFiles(files, projectPath)

	logger.Infof("Detected project type: %s (confidence: %s)", result.AppType, result.Confidence)
	if len(result.Indicators) > 0 {
		logger.Debugf("Detection indicators: %v", result.Indicators)
	}

	return result, nil
}

// getProjectFiles returns a list of files in the project directory
func getProjectFiles(projectPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and files (except .dockerignore, etc.)
		if strings.HasPrefix(info.Name(), ".") && !isImportantDotFile(info.Name()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip node_modules, build directories, etc.
		if info.IsDir() && isIgnoredDirectory(info.Name()) {
			return filepath.SkipDir
		}

		// Only include files (not directories)
		if !info.IsDir() {
			// Get relative path from project root
			relPath, err := filepath.Rel(projectPath, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}

		return nil
	})

	return files, err
}

// analyzeFiles examines the files and determines the project type
func analyzeFiles(files []string, projectPath string) *DetectionResult {
	indicators := make([]string, 0)

	// Check for Docker
	if hasDockerfile(files) {
		indicators = append(indicators, "Dockerfile")
		return &DetectionResult{
			AppType:    config.Docker,
			Confidence: "high",
			Indicators: indicators,
			Path:       projectPath,
		}
	}

	// Check for Node.js
	if hasPackageJson(files) {
		indicators = append(indicators, "package.json")

		// Check if it's also a potential Flask/Python project
		if hasPythonFiles(files) {
			indicators = append(indicators, "Python files detected alongside Node.js")
			return &DetectionResult{
				AppType:    config.NodeJS,
				Confidence: "medium", // Lower confidence due to mixed signals
				Indicators: indicators,
				Path:       projectPath,
			}
		}

		return &DetectionResult{
			AppType:    config.NodeJS,
			Confidence: "high",
			Indicators: indicators,
			Path:       projectPath,
		}
	}

	// Check for Python/Flask
	if hasPythonIndicators(files) {
		if hasFlaskIndicators(files, projectPath) {
			indicators = append(indicators, getFlaskIndicators(files)...)

			// Check if Flask was detected via requirements.txt content
			hasFlaskInReq := false
			for _, file := range files {
				if filepath.Base(file) == "requirements.txt" && hasFlaskInRequirements(file) {
					hasFlaskInReq = true
					break
				}
			}

			confidence := "high"
			if hasFlaskInReq && len(indicators) == 1 && indicators[0] == "requirements.txt" {
				// Only requirements.txt with Flask, still high confidence
				confidence = "high"
			}

			return &DetectionResult{
				AppType:    config.Flask,
				Confidence: confidence,
				Indicators: indicators,
				Path:       projectPath,
			}
		} else {
			indicators = append(indicators, getPythonIndicators(files)...)
			return &DetectionResult{
				AppType:    config.Flask,
				Confidence: "medium", // Assume Flask for Python projects
				Indicators: indicators,
				Path:       projectPath,
			}
		}
	}

	// Check for static site indicators
	if hasStaticSiteIndicators(files) {
		indicators = append(indicators, getStaticSiteIndicators(files)...)
		return &DetectionResult{
			AppType:    config.StaticSite,
			Confidence: "medium",
			Indicators: indicators,
			Path:       projectPath,
		}
	}

	// Default fallback to static site
	return &DetectionResult{
		AppType:    config.StaticSite,
		Confidence: "low",
		Indicators: []string{"No specific indicators found, defaulting to static site"},
		Path:       projectPath,
	}
}

// Docker detection
func hasDockerfile(files []string) bool {
	dockerFiles := []string{"Dockerfile", "dockerfile", "Dockerfile.prod", "Dockerfile.dev"}
	for _, file := range files {
		fileName := filepath.Base(file)
		for _, dockerFile := range dockerFiles {
			if fileName == dockerFile {
				return true
			}
		}
	}
	return false
}

// Node.js detection
func hasPackageJson(files []string) bool {
	for _, file := range files {
		if filepath.Base(file) == "package.json" {
			return true
		}
	}
	return false
}

// Python detection
func hasPythonFiles(files []string) bool {
	for _, file := range files {
		if strings.HasSuffix(file, ".py") {
			return true
		}
	}
	return false
}

func hasPythonIndicators(files []string) bool {
	pythonFiles := []string{"requirements.txt", "setup.py", "pyproject.toml", "Pipfile", "app.py", "main.py"}
	for _, file := range files {
		fileName := filepath.Base(file)
		for _, pythonFile := range pythonFiles {
			if fileName == pythonFile {
				return true
			}
		}
		// Also check for any .py files
		if strings.HasSuffix(file, ".py") {
			return true
		}
	}
	return false
}

func getPythonIndicators(files []string) []string {
	indicators := make([]string, 0)
	pythonFiles := []string{"requirements.txt", "setup.py", "pyproject.toml", "Pipfile", "app.py", "main.py"}

	for _, file := range files {
		fileName := filepath.Base(file)
		for _, pythonFile := range pythonFiles {
			if fileName == pythonFile {
				indicators = append(indicators, fileName)
			}
		}
		if strings.HasSuffix(file, ".py") && len(indicators) < 3 { // Limit to avoid too many .py files
			indicators = append(indicators, fileName)
		}
	}
	return indicators
}

// Flask-specific detection
func hasFlaskIndicators(files []string, projectPath string) bool {
	flaskFiles := []string{"app.py", "application.py", "wsgi.py"}
	for _, file := range files {
		fileName := filepath.Base(file)
		for _, flaskFile := range flaskFiles {
			if fileName == flaskFile {
				return true
			}
		}
	}

	// Check if requirements.txt contains Flask
	for _, file := range files {
		if filepath.Base(file) == "requirements.txt" {
			// Create full path for the file
			fullPath := filepath.Join(projectPath, file)
			return hasFlaskInRequirements(fullPath)
		}
	}

	return false
}

func getFlaskIndicators(files []string) []string {
	indicators := make([]string, 0)
	flaskFiles := []string{"app.py", "application.py", "wsgi.py", "requirements.txt"}

	for _, file := range files {
		fileName := filepath.Base(file)
		for _, flaskFile := range flaskFiles {
			if fileName == flaskFile {
				indicators = append(indicators, fileName)
			}
		}
	}
	return indicators
}

func hasFlaskInRequirements(filePath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	contentLower := strings.ToLower(string(content))
	return strings.Contains(contentLower, "flask")
}

// Static site detection
func hasStaticSiteIndicators(files []string) bool {
	staticFiles := []string{"index.html", "index.htm", "README.md", ".html", ".css", ".js"}

	for _, file := range files {
		fileName := filepath.Base(file)
		for _, staticFile := range staticFiles {
			if strings.HasSuffix(fileName, staticFile) {
				return true
			}
		}
	}
	return false
}

func getStaticSiteIndicators(files []string) []string {
	indicators := make([]string, 0)
	staticExtensions := []string{".html", ".htm", ".css", ".js", ".md"}

	for _, file := range files {
		fileName := filepath.Base(file)
		if fileName == "index.html" || fileName == "index.htm" {
			indicators = append(indicators, fileName)
		}

		for _, ext := range staticExtensions {
			if strings.HasSuffix(fileName, ext) && len(indicators) < 5 {
				indicators = append(indicators, fileName)
				break
			}
		}
	}
	return indicators
}

// Helper functions
func isImportantDotFile(name string) bool {
	importantDotFiles := []string{".dockerignore", ".gitignore", ".env", ".env.example"}
	for _, dotFile := range importantDotFiles {
		if name == dotFile {
			return true
		}
	}
	return false
}

func isIgnoredDirectory(name string) bool {
	ignoredDirs := []string{
		"node_modules", "vendor", "__pycache__", ".git",
		"build", "dist", "target", ".terraform",
		"venv", "env", ".venv", ".env",
	}
	for _, ignoredDir := range ignoredDirs {
		if name == ignoredDir {
			return true
		}
	}
	return false
}

// GetAppTypeString returns the string representation of the detected app type
func (r *DetectionResult) GetAppTypeString() string {
	return string(r.AppType)
}

// IsHighConfidence returns true if the detection confidence is high
func (r *DetectionResult) IsHighConfidence() bool {
	return r.Confidence == "high"
}

// String returns a human-readable description of the detection result
func (r *DetectionResult) String() string {
	return fmt.Sprintf("Detected %s project with %s confidence based on: %v",
		r.AppType, r.Confidence, r.Indicators)
}
