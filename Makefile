# CloudDeploy CLI Makefile

# Variables
BINARY_NAME=clouddeploy
MAIN_PACKAGE=.
BUILD_DIR=./bin
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
BUILD_USER=$(shell whoami)
GIT_COMMIT=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")

# Build flags for production
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.buildUser=$(BUILD_USER) -X main.gitCommit=$(GIT_COMMIT)"

# Default target
.PHONY: all
all: clean build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for current platform (quick build)
.PHONY: build-local
build-local:
	@echo "Building $(BINARY_NAME) for local development..."
	go build -o $(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Binary built: ./$(BINARY_NAME)"

# Build for multiple platforms
.PHONY: build-all
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	
	@echo "All binaries built in $(BUILD_DIR)/"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detector
.PHONY: test-race
test-race:
	@echo "Running tests with race detector..."
	go test -race ./...

# Run integration tests
.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	go test -tags=integration -v ./...

# Run all tests
.PHONY: test-all
test-all: test test-race test-integration

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@rm -rf .terraform-generated
	@rm -f .clouddeploy.json

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		go vet ./...; \
	fi

# Security scanning
.PHONY: security
security:
	@echo "Running security scans..."
	go vet ./...
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not found. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

# Run benchmarks
.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Production build with security flags
.PHONY: build-production
build-production:
	@echo "Building $(BINARY_NAME) for production..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -buildmode=pie $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Production binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application (for development)
.PHONY: run
run: build-local
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Install the binary to GOPATH/bin
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) $(MAIN_PACKAGE)

# Development setup
.PHONY: dev-setup
dev-setup: deps
	@echo "Setting up development environment..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@echo "Development environment ready!"

# Install security and development tools
.PHONY: install-tools
install-tools:
	@echo "Installing development and security tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Tools installed successfully"

# Pre-commit checks
.PHONY: pre-commit
pre-commit: fmt lint security test-all
	@echo "Pre-commit checks completed successfully"

# Release preparation
.PHONY: release-check
release-check: pre-commit build-production
	@echo "Release readiness check completed"

# Help
.PHONY: help
help:
	@echo "CloudDeploy CLI Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build              Build the binary for current platform"
	@echo "  build-local        Quick build for local development"
	@echo "  build-production   Build with production security flags"
	@echo "  build-all          Build for multiple platforms"
	@echo "  test               Run tests"
	@echo "  test-coverage      Run tests with coverage report"
	@echo "  test-race          Run tests with race detector"
	@echo "  test-integration   Run integration tests"
	@echo "  test-all           Run all tests (unit, race, integration)"
	@echo "  benchmark          Run benchmarks"
	@echo "  clean              Clean build artifacts"
	@echo "  deps               Install dependencies"
	@echo "  fmt                Format code"
	@echo "  lint               Lint code"
	@echo "  security           Run security scans"
	@echo "  run                Build and run the application"
	@echo "  install            Install binary to GOPATH/bin"
	@echo "  dev-setup          Set up development environment"
	@echo "  install-tools      Install development and security tools"
	@echo "  pre-commit         Run all pre-commit checks"
	@echo "  release-check      Check if ready for release"
	@echo "  help               Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build              # Build for current platform"
	@echo "  make test               # Run tests"
	@echo "  make build-all          # Build for all platforms"
	@echo "  make VERSION=v1.0.0 build  # Build with version" 