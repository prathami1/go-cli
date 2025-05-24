# CloudDeploy CLI Makefile

# Variables
BINARY_NAME=clouddeploy
MAIN_PACKAGE=.
BUILD_DIR=./bin
VERSION?=dev
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

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

# Help
.PHONY: help
help:
	@echo "CloudDeploy CLI Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build         Build the binary for current platform"
	@echo "  build-local   Quick build for local development"
	@echo "  build-all     Build for multiple platforms"
	@echo "  test          Run tests"
	@echo "  test-coverage Run tests with coverage report"
	@echo "  clean         Clean build artifacts"
	@echo "  deps          Install dependencies"
	@echo "  fmt           Format code"
	@echo "  lint          Lint code"
	@echo "  run           Build and run the application"
	@echo "  install       Install binary to GOPATH/bin"
	@echo "  dev-setup     Set up development environment"
	@echo "  help          Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build              # Build for current platform"
	@echo "  make test               # Run tests"
	@echo "  make build-all          # Build for all platforms"
	@echo "  make VERSION=v1.0.0 build  # Build with version" 