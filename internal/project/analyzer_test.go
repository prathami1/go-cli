package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/prathami1/go-cli/internal/config"
)

func TestAnalyzeProject(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		expectedType  config.AppType
		expectedConf  string
		expectedFiles []string // Files that should be in indicators
	}{
		{
			name:          "Docker project",
			files:         []string{"Dockerfile", "package.json", "src/main.js"},
			expectedType:  config.Docker,
			expectedConf:  "high",
			expectedFiles: []string{"Dockerfile"},
		},
		{
			name:          "Node.js project",
			files:         []string{"package.json", "src/index.js", "README.md"},
			expectedType:  config.NodeJS,
			expectedConf:  "high",
			expectedFiles: []string{"package.json"},
		},
		{
			name:          "Node.js with Python files",
			files:         []string{"package.json", "script.py", "src/index.js"},
			expectedType:  config.NodeJS,
			expectedConf:  "medium",
			expectedFiles: []string{"package.json"},
		},
		{
			name:          "Flask project with app.py",
			files:         []string{"app.py", "requirements.txt", "templates/index.html"},
			expectedType:  config.Flask,
			expectedConf:  "high",
			expectedFiles: []string{"app.py", "requirements.txt"},
		},
		{
			name:          "Python project with requirements.txt",
			files:         []string{"main.py", "requirements.txt", "utils.py"},
			expectedType:  config.Flask,
			expectedConf:  "medium",
			expectedFiles: []string{"requirements.txt", "main.py"},
		},
		{
			name:          "Flask in requirements.txt",
			files:         []string{"server.py", "requirements.txt"},
			expectedType:  config.Flask,
			expectedConf:  "high",
			expectedFiles: []string{"requirements.txt"},
		},
		{
			name:          "Static site with HTML",
			files:         []string{"index.html", "style.css", "script.js"},
			expectedType:  config.StaticSite,
			expectedConf:  "medium",
			expectedFiles: []string{"index.html"},
		},
		{
			name:         "Empty project defaults to static",
			files:        []string{},
			expectedType: config.StaticSite,
			expectedConf: "low",
		},
		{
			name:          "Mixed indicators prioritizes Docker",
			files:         []string{"Dockerfile", "package.json", "app.py", "index.html"},
			expectedType:  config.Docker,
			expectedConf:  "high",
			expectedFiles: []string{"Dockerfile"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory with test files
			tempDir := createTestProject(t, tt.files)
			defer os.RemoveAll(tempDir)

			// Create requirements.txt with Flask if needed for Flask test
			if tt.name == "Flask in requirements.txt" {
				reqPath := filepath.Join(tempDir, "requirements.txt")
				err := os.WriteFile(reqPath, []byte("Flask==2.0.0\nrequests==2.25.1"), 0644)
				if err != nil {
					t.Fatalf("Failed to create requirements.txt: %v", err)
				}
			}

			// Analyze the project
			result, err := AnalyzeProject(tempDir)
			if err != nil {
				t.Fatalf("AnalyzeProject failed: %v", err)
			}

			// Check app type
			if result.AppType != tt.expectedType {
				t.Errorf("Expected app type %s, got %s", tt.expectedType, result.AppType)
			}

			// Check confidence
			if result.Confidence != tt.expectedConf {
				t.Errorf("Expected confidence %s, got %s", tt.expectedConf, result.Confidence)
			}

			// Check that expected files are in indicators
			for _, expectedFile := range tt.expectedFiles {
				found := false
				for _, indicator := range result.Indicators {
					if indicator == expectedFile {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected indicator %s not found in %v", expectedFile, result.Indicators)
				}
			}
		})
	}
}

func TestHasDockerfile(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected bool
	}{
		{"Standard Dockerfile", []string{"Dockerfile", "src/main.go"}, true},
		{"Lowercase dockerfile", []string{"dockerfile", "README.md"}, true},
		{"Production Dockerfile", []string{"Dockerfile.prod", "config.yaml"}, true},
		{"No Dockerfile", []string{"package.json", "index.js"}, false},
		{"Dockerfile in subdirectory", []string{"deploy/Dockerfile", "src/main.py"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasDockerfile(tt.files)
			if result != tt.expected {
				t.Errorf("hasDockerfile() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHasPackageJson(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected bool
	}{
		{"Has package.json", []string{"package.json", "src/index.js"}, true},
		{"No package.json", []string{"requirements.txt", "app.py"}, false},
		{"package.json in subdirectory", []string{"frontend/package.json", "backend/app.py"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasPackageJson(tt.files)
			if result != tt.expected {
				t.Errorf("hasPackageJson() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHasPythonIndicators(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected bool
	}{
		{"Has requirements.txt", []string{"requirements.txt", "main.py"}, true},
		{"Has app.py", []string{"app.py", "templates/index.html"}, true},
		{"Has .py files", []string{"script.py", "utils.py"}, true},
		{"Has setup.py", []string{"setup.py", "README.md"}, true},
		{"No Python indicators", []string{"package.json", "index.js"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasPythonIndicators(tt.files)
			if result != tt.expected {
				t.Errorf("hasPythonIndicators() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHasFlaskIndicators(t *testing.T) {
	// Create a temporary directory for testing requirements.txt content
	tempDir := createTestProject(t, []string{"requirements.txt"})
	defer os.RemoveAll(tempDir)

	// Create requirements.txt with Flask
	reqPath := filepath.Join(tempDir, "requirements.txt")
	err := os.WriteFile(reqPath, []byte("Flask==2.0.0\nrequests==2.25.1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create requirements.txt: %v", err)
	}

	tests := []struct {
		name     string
		files    []string
		expected bool
	}{
		{"Has app.py", []string{"app.py", "templates/index.html"}, true},
		{"Has application.py", []string{"application.py", "static/style.css"}, true},
		{"Has wsgi.py", []string{"wsgi.py", "config.py"}, true},
		{"Has Flask in requirements", []string{"requirements.txt"}, true},
		{"No Flask indicators", []string{"script.py", "utils.py"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasFlaskIndicators(tt.files, tempDir)
			if result != tt.expected {
				t.Errorf("hasFlaskIndicators() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHasStaticSiteIndicators(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected bool
	}{
		{"Has index.html", []string{"index.html", "style.css"}, true},
		{"Has HTML files", []string{"about.html", "contact.html"}, true},
		{"Has CSS files", []string{"styles.css", "main.css"}, true},
		{"Has JS files", []string{"script.js", "app.js"}, true},
		{"Has README.md", []string{"README.md"}, true},
		{"No static indicators", []string{"main.py", "app.py"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasStaticSiteIndicators(tt.files)
			if result != tt.expected {
				t.Errorf("hasStaticSiteIndicators() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestDetectionResultMethods(t *testing.T) {
	result := &DetectionResult{
		AppType:    config.NodeJS,
		Confidence: "high",
		Indicators: []string{"package.json"},
		Path:       "/test/path",
	}

	// Test GetAppTypeString
	if result.GetAppTypeString() != "nodejs" {
		t.Errorf("GetAppTypeString() = %s, expected nodejs", result.GetAppTypeString())
	}

	// Test IsHighConfidence
	if !result.IsHighConfidence() {
		t.Error("IsHighConfidence() = false, expected true")
	}

	// Test with medium confidence
	result.Confidence = "medium"
	if result.IsHighConfidence() {
		t.Error("IsHighConfidence() = true, expected false for medium confidence")
	}

	// Test String method
	str := result.String()
	expectedSubstrings := []string{"nodejs", "medium", "package.json"}
	for _, substr := range expectedSubstrings {
		if !containsString(str, substr) {
			t.Errorf("String() output %q does not contain %q", str, substr)
		}
	}
}

func TestIsIgnoredDirectory(t *testing.T) {
	tests := []struct {
		name     string
		dir      string
		expected bool
	}{
		{"node_modules", "node_modules", true},
		{"__pycache__", "__pycache__", true},
		{".git", ".git", true},
		{"build", "build", true},
		{"src", "src", false},
		{"assets", "assets", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isIgnoredDirectory(tt.dir)
			if result != tt.expected {
				t.Errorf("isIgnoredDirectory(%s) = %v, expected %v", tt.dir, result, tt.expected)
			}
		})
	}
}

func TestIsImportantDotFile(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		expected bool
	}{
		{".dockerignore", ".dockerignore", true},
		{".gitignore", ".gitignore", true},
		{".env", ".env", true},
		{".hidden", ".hidden", false},
		{".DS_Store", ".DS_Store", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isImportantDotFile(tt.file)
			if result != tt.expected {
				t.Errorf("isImportantDotFile(%s) = %v, expected %v", tt.file, result, tt.expected)
			}
		})
	}
}

// Helper functions for tests

func createTestProject(t *testing.T, files []string) string {
	tempDir, err := os.MkdirTemp("", "test-project-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	for _, file := range files {
		fullPath := filepath.Join(tempDir, file)

		// Create directory if needed
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}

		// Create file
		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	return tempDir
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
