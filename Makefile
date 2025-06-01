# CloudDeploy CLI Makefile

BINARY_NAME=cdeploy
VERSION=v1.0.0
GO_VERSION=1.21
MAIN_PATH=main.go

# Build directories
BUILD_DIR=bin
DIST_DIR=dist

# Go build flags for production
PROD_LDFLAGS=-w -s -X main.Version=$(VERSION)
DEV_LDFLAGS=-X main.Version=$(VERSION)-dev

# Security flags
SECURITY_FLAGS=-buildmode=pie
CGO_ENABLED=0

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: all build build-production build-dev test test-all test-coverage \
        test-race test-integration clean deps install-tools security \
        release release-all dev-setup help lint format tidy

# Default target
all: build

# === BUILD TARGETS ===

## Build for development (with debug symbols)
build: build-dev

## Build for development (faster builds, debug symbols)
build-dev:
	@echo "$(BLUE)Building $(BINARY_NAME) for development...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) go build \
		-ldflags="$(DEV_LDFLAGS)" \
		-o $(BUILD_DIR)/$(BINARY_NAME) \
		$(MAIN_PATH)
	@echo "$(GREEN)✅ Development build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## Build for production (optimized, stripped)
build-production:
	@echo "$(BLUE)Building $(BINARY_NAME) for production...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) go build \
		$(SECURITY_FLAGS) \
		-ldflags="$(PROD_LDFLAGS)" \
		-trimpath \
		-o $(BUILD_DIR)/$(BINARY_NAME) \
		$(MAIN_PATH)
	@echo "$(GREEN)✅ Production build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"
	@echo "$(YELLOW)Binary size: $(shell du -h $(BUILD_DIR)/$(BINARY_NAME) | cut -f1)$(NC)"

## Build for all major platforms
build-all: clean-dist
	@echo "$(BLUE)Building $(BINARY_NAME) for all platforms...$(NC)"
	@mkdir -p $(DIST_DIR)
	
	# Linux AMD64
	@echo "$(YELLOW)Building for Linux AMD64...$(NC)"
	GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build \
		$(SECURITY_FLAGS) \
		-ldflags="$(PROD_LDFLAGS)" \
		-trimpath \
		-o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 \
		$(MAIN_PATH)
	
	# Linux ARM64
	@echo "$(YELLOW)Building for Linux ARM64...$(NC)"
	GOOS=linux GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) go build \
		$(SECURITY_FLAGS) \
		-ldflags="$(PROD_LDFLAGS)" \
		-trimpath \
		-o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 \
		$(MAIN_PATH)
	
	# macOS AMD64
	@echo "$(YELLOW)Building for macOS AMD64...$(NC)"
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build \
		$(SECURITY_FLAGS) \
		-ldflags="$(PROD_LDFLAGS)" \
		-trimpath \
		-o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 \
		$(MAIN_PATH)
	
	# macOS ARM64 (Apple Silicon)
	@echo "$(YELLOW)Building for macOS ARM64...$(NC)"
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) go build \
		$(SECURITY_FLAGS) \
		-ldflags="$(PROD_LDFLAGS)" \
		-trimpath \
		-o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 \
		$(MAIN_PATH)
	
	# Windows AMD64
	@echo "$(YELLOW)Building for Windows AMD64...$(NC)"
	GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build \
		-ldflags="$(PROD_LDFLAGS)" \
		-trimpath \
		-o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe \
		$(MAIN_PATH)
	
	@echo "$(GREEN)✅ Multi-platform build complete in $(DIST_DIR)/$(NC)"
	@ls -la $(DIST_DIR)/

# === CLEAN TARGETS ===

## Remove build artifacts and config files
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@rm -f .clouddeploy.json
	@rm -rf .clouddeploy-tf
	@echo "$(GREEN)✅ Clean complete$(NC)"

## Remove only distribution artifacts
clean-dist:
	@echo "$(YELLOW)Cleaning distribution artifacts...$(NC)"
	@rm -rf $(DIST_DIR)

# === DEPENDENCY MANAGEMENT ===

## Download and verify dependencies
deps:
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	go mod download
	go mod verify
	@echo "$(GREEN)✅ Dependencies ready$(NC)"

## Update dependencies
deps-update:
	@echo "$(BLUE)Updating dependencies...$(NC)"
	go get -u ./...
	go mod tidy
	@echo "$(GREEN)✅ Dependencies updated$(NC)"

## Tidy up go.mod and go.sum
tidy:
	@echo "$(BLUE)Tidying dependencies...$(NC)"
	go mod tidy
	@echo "$(GREEN)✅ Dependencies tidied$(NC)"

# === TESTING TARGETS ===

## Run unit tests
test:
	@echo "$(BLUE)Running unit tests...$(NC)"
	go test -v ./...
	@echo "$(GREEN)✅ Unit tests complete$(NC)"

## Run all tests (unit + integration + race detection)
test-all: test test-race test-integration

## Run tests with coverage report
test-coverage:
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	@mkdir -p coverage
	go test -v -coverprofile=coverage/coverage.out ./...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "$(GREEN)✅ Coverage report: coverage/coverage.html$(NC)"

## Run tests with race detection
test-race:
	@echo "$(BLUE)Running tests with race detection...$(NC)"
	go test -race -v ./...
	@echo "$(GREEN)✅ Race detection tests complete$(NC)"

## Run integration tests
test-integration:
	@echo "$(BLUE)Running integration tests...$(NC)"
	# Add integration test commands here
	@echo "$(YELLOW)⚠️  Integration tests not yet implemented$(NC)"

# === CODE QUALITY ===

## Run linter
lint:
	@echo "$(BLUE)Running linter...$(NC)"
	@which golangci-lint > /dev/null || (echo "$(RED)golangci-lint not found. Run 'make install-tools'$(NC)" && exit 1)
	golangci-lint run
	@echo "$(GREEN)✅ Linting complete$(NC)"

## Format code
format:
	@echo "$(BLUE)Formatting code...$(NC)"
	go fmt ./...
	@echo "$(GREEN)✅ Code formatted$(NC)"

# === SECURITY ===

## Run security scanning
security:
	@echo "$(BLUE)Running security scans...$(NC)"
	@which gosec > /dev/null || (echo "$(RED)gosec not found. Run 'make install-tools'$(NC)" && exit 1)
	gosec ./...
	@echo "$(YELLOW)Running vulnerability check...$(NC)"
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	@echo "$(GREEN)✅ Security scanning complete$(NC)"

# === DEVELOPMENT SETUP ===

## Install development tools
install-tools:
	@echo "$(BLUE)Installing development tools...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/govulncheck@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "$(GREEN)✅ Development tools installed$(NC)"

## Set up development environment
dev-setup: install-tools deps
	@echo "$(BLUE)Setting up development environment...$(NC)"
	@echo "$(GREEN)✅ Development environment ready$(NC)"
	@echo "$(YELLOW)Now run: make build-dev$(NC)"

# === INSTALLATION ===

## Install binary to system PATH (requires sudo)
install: build-production
	@echo "$(BLUE)Installing $(BINARY_NAME) to /usr/local/bin...$(NC)"
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)✅ $(BINARY_NAME) installed to /usr/local/bin$(NC)"

## Uninstall binary from system PATH (requires sudo)
uninstall:
	@echo "$(YELLOW)Removing $(BINARY_NAME) from /usr/local/bin...$(NC)"
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)✅ $(BINARY_NAME) uninstalled$(NC)"

# === HELP ===

## Show available targets and their descriptions
help:
	@echo "$(BLUE)CloudDeploy CLI Makefile$(NC)"
	@echo ""
	@echo "$(GREEN)Available targets:$(NC)"
	@echo ""
	@echo "$(YELLOW)Build:$(NC)"
	@echo "  build              Build for development (default)"
	@echo "  build-dev          Build with debug symbols"
	@echo "  build-production   Build optimized for production"
	@echo "  build-all          Build for all platforms"
	@echo ""
	@echo "$(YELLOW)Clean:$(NC)"
	@echo "  clean              Remove all build artifacts"
	@echo "  clean-dist         Remove distribution artifacts"
	@echo ""
	@echo "$(YELLOW)Dependencies:$(NC)"
	@echo "  deps               Download dependencies"
	@echo "  deps-update        Update dependencies"
	@echo "  tidy               Tidy go.mod and go.sum"
	@echo ""
	@echo "$(YELLOW)Testing:$(NC)"
	@echo "  test               Run unit tests"
	@echo "  test-all           Run all tests"
	@echo "  test-coverage      Generate coverage report"
	@echo "  test-race          Run with race detection"
	@echo ""
	@echo "$(YELLOW)Code Quality:$(NC)"
	@echo "  lint               Run linter"
	@echo "  format             Format code"
	@echo "  security           Run security scans"
	@echo ""
	@echo "$(YELLOW)Development:$(NC)"
	@echo "  install-tools      Install development tools"
	@echo "  dev-setup          Set up development environment"
	@echo ""
	@echo "$(YELLOW)Installation:$(NC)"
	@echo "  install            Install to /usr/local/bin"
	@echo "  uninstall          Remove from /usr/local/bin"
	@echo ""
	@echo "Examples:"
	@echo "  make build-production  # Build optimized binary"
	@echo "  make test-all          # Run comprehensive tests"
	@echo "  make install           # Install to system PATH" 