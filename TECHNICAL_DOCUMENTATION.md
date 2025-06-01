# CloudDeploy CLI - Technical Documentation

> **How CloudDeploy Works Under the Hood** 🔧

This document provides a comprehensive technical overview of CloudDeploy's internal architecture, algorithms, and workflows. It's designed for developers who want to understand how the auto-detection, Terraform generation, deployment, and authentication mechanisms work.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Project Auto-Detection](#project-auto-detection)
3. [Terraform Configuration Generation](#terraform-configuration-generation)
4. [Cloud Provider Authentication](#cloud-provider-authentication)
5. [CLI Auto-Installation](#cli-auto-installation)
6. [Deployment Workflow](#deployment-workflow)
7. [Configuration Management](#configuration-management)
8. [User Interface and Progress Indication](#user-interface-and-progress-indication)
9. [Build System and Distribution](#build-system-and-distribution)
10. [Security Considerations](#security-considerations)

## Architecture Overview

CloudDeploy follows a modular, layered architecture that separates concerns and provides clean interfaces between components:

```
┌─────────────────────────────────────────────────────────────────┐
│                        Command Layer (cmd/)                     │
├─────────────────────────────────────────────────────────────────┤
│  start.go    │  deploy.go   │  destroy.go   │  root.go         │
│  (One-cmd)   │  (Deploy)    │  (Cleanup)    │  (Global)        │
└─────────────────────────────────────────────────────────────────┘
                                   │
┌─────────────────────────────────────────────────────────────────┐
│                      Business Logic (internal/)                 │
├─────────────────────────────────────────────────────────────────┤
│ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │
│ │   project/  │ │ terraform/  │ │ providers/  │ │   config/   │ │
│ │ (Analysis)  │ │(Generation) │ │  (Auth)     │ │(Management) │ │
│ └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ │
│ ┌─────────────┐ ┌─────────────┐                                 │
│ │   utils/    │ │   logger/   │                                 │
│ │ (UI/Utils)  │ │ (Logging)   │                                 │
│ └─────────────┘ └─────────────┘                                 │
└─────────────────────────────────────────────────────────────────┘
                                   │
┌─────────────────────────────────────────────────────────────────┐
│                      External Systems                           │
├─────────────────────────────────────────────────────────────────┤
│  Terraform CLI  │  AWS CLI     │  GCP CLI     │  Azure CLI      │
│  File System    │  HTTP APIs   │  Cloud APIs  │  Package Mgrs   │
└─────────────────────────────────────────────────────────────────┘
```

### Key Design Principles

1. **Separation of Concerns**: Each package has a single responsibility
2. **Dependency Injection**: External dependencies are injected, not hardcoded
3. **Error Propagation**: Errors bubble up with context for better debugging
4. **Immutable Configuration**: Configuration is created once and passed down
5. **Progressive Enhancement**: Basic functionality works, advanced features enhance UX

## Project Auto-Detection

The auto-detection system is one of CloudDeploy's most sophisticated features. It analyzes project directories to determine the application type with confidence scoring.

### Detection Algorithm

Located in `internal/project/analyzer.go`, the detection system works as follows:

```go
// Core detection flow
func AnalyzeProject(projectPath string) (*DetectionResult, error) {
    files := getProjectFiles(projectPath)  // Step 1: File Discovery
    result := analyzeFiles(files)          // Step 2: Pattern Analysis
    return result, nil                     // Step 3: Confidence Scoring
}
```

#### Step 1: File Discovery (`getProjectFiles`)

The system performs an intelligent directory walk that:

```go
// Key behaviors:
- Skips hidden directories (.git, .terraform)
- Ignores build artifacts (node_modules, dist, build)
- Includes important dotfiles (.dockerignore, .env)
- Uses filepath.Walk for efficient traversal
- Returns relative paths for analysis
```

**Ignored Directories:**
- `node_modules`, `vendor`, `__pycache__`
- `.git`, `.terraform`, `build`, `dist`, `target`
- `venv`, `env`, `.venv`

**Important Dotfiles Included:**
- `.dockerignore`, `.gitignore`, `.env`, `.env.example`

#### Step 2: Pattern Analysis (`analyzeFiles`)

The analysis engine uses a **priority-based detection system**:

```go
// Detection priority (highest to lowest):
1. Docker        → Dockerfile presence
2. Node.js       → package.json presence  
3. Python/Flask  → Python indicators + Flask-specific files
4. Static Site   → HTML/CSS/JS files
5. Fallback      → Default to static site
```

**Docker Detection** (Highest Priority):
```go
func hasDockerfile(files []string) bool {
    dockerFiles := []string{"Dockerfile", "dockerfile", "Dockerfile.prod", "Dockerfile.dev"}
    // Returns true if any Docker-related file exists
}
```

**Node.js Detection**:
```go
func hasPackageJson(files []string) bool {
    // Looks for package.json in project root
    // Handles mixed projects (Node.js + Python)
}
```

**Flask/Python Detection**:
```go
func hasFlaskIndicators(files []string, projectPath string) bool {
    // Checks for: app.py, application.py, wsgi.py
    // Scans requirements.txt for "flask" dependency
    // Confidence based on number of indicators
}
```

**Static Site Detection**:
```go
func hasStaticSiteIndicators(files []string) bool {
    // Looks for: index.html, .css, .js files
    // Medium confidence due to ubiquity
}
```

#### Step 3: Confidence Scoring

The system assigns confidence levels based on detection strength:

```go
type DetectionResult struct {
    AppType    config.AppType `json:"app_type"`
    Confidence string         `json:"confidence"`    // "high", "medium", "low"
    Indicators []string       `json:"indicators"`    // Evidence files
    Path       string         `json:"path"`          // Analyzed directory
}
```

**Confidence Levels:**
- **High**: Definitive indicators (Dockerfile, package.json + no conflicts)
- **Medium**: Some indicators but potential ambiguity
- **Low**: Fallback detection or conflicting signals

**Smart Conflict Resolution:**
```go
// Example: Node.js project with Python files
if hasPackageJson(files) && hasPythonFiles(files) {
    // Lower confidence due to mixed signals
    return &DetectionResult{
        AppType:    config.NodeJS,
        Confidence: "medium",
        Indicators: []string{"package.json", "Python files detected alongside Node.js"},
    }
}
```

### Detection Results Usage

The detection results influence the entire deployment pipeline:

1. **Interactive Prompts**: High confidence → suggest detected type
2. **Template Selection**: AppType determines Terraform template directory
3. **Infrastructure**: Different cloud resources per application type
4. **Validation**: Re-analysis during deployment to catch changes

## Terraform Configuration Generation

CloudDeploy's Terraform generation system is template-based and highly modular, supporting multiple cloud providers with consistent patterns.

### Template-Based Architecture

Located in `internal/terraform/`, the system uses Go's `text/template` with a sophisticated directory structure:

```
assets/templates/
├── aws/
│   ├── provider.tf.tpl      # AWS provider configuration
│   ├── compute.tf.tpl       # EC2, ECS, ELB resources
│   ├── database.tf.tpl      # RDS resources
│   ├── storage.tf.tpl       # S3 resources
│   └── loadbalancer.tf.tpl  # ALB/ELB resources
├── gcp/
│   ├── provider.tf.tpl      # GCP provider configuration
│   ├── compute.tf.tpl       # Compute Engine, Cloud Run
│   ├── database.tf.tpl      # Cloud SQL resources
│   ├── storage.tf.tpl       # Cloud Storage resources
│   └── loadbalancer.tf.tpl  # Load Balancer resources
└── azure/
    ├── provider.tf.tpl      # Azure provider configuration
    ├── compute.tf.tpl       # VM, Container Instances
    ├── database.tf.tpl      # Azure SQL resources
    ├── storage.tf.tpl       # Storage Account resources
    └── loadbalancer.tf.tpl  # Application Gateway resources
```

### Generation Process

The generation workflow in `internal/terraform/terraform.go`:

```go
func GenerateConfigInDir(cfg *config.DeploymentConfig, targetDir string) error {
    // Step 1: Prepare template data
    templateData := &TemplateData{
        ProjectName:        cfg.ProjectName,
        AppType:            string(cfg.AppType),
        CloudProvider:      string(cfg.CloudProvider),
        Region:             cfg.Region,
        Environment:        "production",
        EnableDatabase:     cfg.Services.Database,
        EnableStorage:      cfg.Services.Storage,
        EnableLoadBalancer: cfg.Services.LoadBalancer,
    }
    
    // Step 2: Select templates based on configuration
    templates := []string{"provider.tf.tpl", "compute.tf.tpl"}
    if cfg.Services.Database {
        templates = append(templates, "database.tf.tpl")
    }
    // ... conditional template loading
    
    // Step 3: Render each template
    for _, templateName := range templates {
        rendered, err := RenderTemplate(templatePath, templateData)
        outputPath := strings.TrimSuffix(templateName, ".tpl")
        err = os.WriteFile(filepath.Join(targetDir, outputPath), []byte(rendered), 0644)
    }
    
    // Step 4: Generate supporting files
    generateVariablesTFInDir(cfg, targetDir)
    generateOutputsTFInDir(cfg, targetDir) 
    generateTFVarsInDir(cfg, targetDir)
}
```

### Template Engine Features

The template system uses Go's `text/template` with custom functions:

```go
func templateFuncs() template.FuncMap {
    return template.FuncMap{
        "replace": strings.ReplaceAll,
        "lower":   strings.ToLower,
        "upper":   strings.ToUpper,
        "title":   strings.Title,
    }
}
```

**Template Examples:**

AWS Compute Template (`assets/templates/aws/compute.tf.tpl`):
```hcl
{{if eq .AppType "docker"}}
# ECS Cluster for Docker applications
resource "aws_ecs_cluster" "main" {
  name = "{{.ProjectName}}-cluster"
  
  tags = {
    Project = "{{.ProjectName}}"
    AppType = "{{.AppType}}"
  }
}
{{else}}
# EC2 Instance for Node.js/Flask applications
resource "aws_instance" "app" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.micro"
  subnet_id     = aws_subnet.public[0].id
  
  tags = {
    Name    = "{{.ProjectName}}-instance"
    Project = "{{.ProjectName}}"
  }
}
{{end}}
```

### Application-Specific Infrastructure

Different application types generate different infrastructure:

**Static Sites:**
- **AWS**: S3 bucket with website hosting + CloudFront
- **GCP**: Cloud Storage with public access
- **Azure**: Storage Account with static website hosting

**Node.js/Flask Applications:**
- **AWS**: EC2 instance with security groups + optional ALB
- **GCP**: Compute Engine instance with firewall rules
- **Azure**: Virtual Machine with Network Security Group

**Docker Applications:**
- **AWS**: ECS cluster with Fargate + ALB
- **GCP**: Cloud Run service with public access
- **Azure**: Container Instances with public IP

### Output Generation

The system generates comprehensive outputs for each cloud provider:

```go
func generateAWSOutputs(cfg *config.DeploymentConfig) string {
    switch cfg.AppType {
    case config.StaticSite:
        return `
output "website_url" {
  description = "Website URL"
  value       = "https://${aws_s3_bucket_website_configuration.main.website_endpoint}"
}`
    case config.NodeJS, config.Flask:
        return `
output "application_url" {
  description = "Application URL"  
  value       = "http://${aws_instance.app.public_ip}"
}`
    }
}
```

### Working Directory Management

CloudDeploy uses a persistent working directory (`.clouddeploy-tf/`) to maintain Terraform state:

```go
const TerraformWorkingDir = ".clouddeploy-tf"

// Benefits:
- Persistent state across runs
- Plan files preserved for apply
- State file location consistent
- Easy cleanup with destroy command
```

## Cloud Provider Authentication

CloudDeploy's authentication system is designed for enterprise environments with automatic CLI installation and Bloomberg-specific optimizations.

### Authentication Architecture

The authentication system in `internal/providers/providers.go` provides:

1. **Automatic CLI Detection & Installation**
2. **Environment-Aware Authentication** (Bloomberg vs Standard)
3. **Progressive Fallback Strategies**
4. **Interactive Authentication Flows**

### Authentication Flow

```go
func CheckAuthentication(provider config.CloudProvider) error {
    switch provider {
    case config.AWS:
        return checkAndAutoLoginAWS()
    case config.GCP:
        return checkAndAutoLoginGCP()
    case config.Azure:
        return checkAndAutoLoginAzure()
    }
}
```

#### AWS Authentication

```go
func checkAndAutoLoginAWS() error {
    // Step 1: Ensure CLI is installed
    if err := checkAndPromptInstall(config.AWS, "aws"); err != nil {
        return fmt.Errorf("AWS CLI installation failed: %w", err)
    }
    
    // Step 2: Check current authentication
    cmd := exec.Command("aws", "sts", "get-caller-identity")
    output, err := cmd.Output()
    if err != nil {
        // Step 3: Trigger automatic login
        return autoLoginAWS()
    }
    
    // Step 4: Verify authentication success
    if strings.Contains(string(output), "UserId") {
        return nil // Already authenticated
    }
}
```

### Enterprise Authentication (Bloomberg)

CloudDeploy includes sophisticated Bloomberg environment detection:

```go
func isBloombergEnvironment() bool {
    hostname, _ := os.Hostname()
    username := os.Getenv("USER")
    domain := os.Getenv("USERDOMAIN")
    
    return strings.Contains(strings.ToLower(hostname), "bloomberg") ||
           strings.Contains(strings.ToLower(username), "corp") ||
           strings.Contains(strings.ToLower(domain), "bloomberg") ||
           os.Getenv("BLOOMBERG_ENV") != ""
}
```

**Bloomberg-Specific Authentication:**

**Azure (Primary Bloomberg Provider):**
```go
func authenticateBloombergAzure() error {
    // Uses device code flow for enterprise SSO compatibility
    deviceCmd := exec.Command("az", "login", "--use-device-code", "--tenant", "common")
    deviceCmd.Stdout = os.Stdout
    deviceCmd.Stderr = os.Stderr
    deviceCmd.Stdin = os.Stdin
    
    if err := deviceCmd.Run(); err != nil {
        // Fallback to standard method
        return authenticateStandardAzure()
    }
}
```

**AWS (Enterprise SSO):**
```go
func authenticateBloombergAWS() error {
    // Try SSO first for Bloomberg
    ssoCmd := exec.Command("aws", "sso", "login")
    ssoCmd.Stdout = os.Stdout
    ssoCmd.Stderr = os.Stderr
    ssoCmd.Stdin = os.Stdin
    
    if err := ssoCmd.Run(); err != nil {
        // Fallback to standard AWS configure
        return authenticateStandardAWS()
    }
}
```

### Authentication Verification

Each provider has specific verification commands:

- **AWS**: `aws sts get-caller-identity`
- **GCP**: `gcloud auth list --filter=status:ACTIVE --format=value(account)`
- **Azure**: `az account show`

## CLI Auto-Installation

CloudDeploy automatically installs missing cloud provider CLI tools with beautiful progress indicators and platform-specific optimizations.

### Installation Architecture

Located in `internal/providers/installer.go`, the installation system supports:

- **Cross-platform compatibility** (Windows, macOS, Linux)
- **Multiple installation methods** per platform
- **Progressive fallback strategies**
- **Beautiful progress indicators**
- **Security-conscious downloads**

### Installation Detection

```go
func checkAndPromptInstall(provider config.CloudProvider, cmdName string) error {
    if commandExists(cmdName) {
        return nil // CLI already exists
    }
    
    // Prompt user for installation permission
    if !promptYesNo(fmt.Sprintf("Would you like to install %s CLI now? (y/N)", provider)) {
        return fmt.Errorf("%s CLI is required but not installed", provider)
    }
    
    return installCLI(provider)
}
```

### Platform-Specific Installation

#### AWS CLI Installation

**Windows:**
```go
func installAWSCLIWindows() error {
    // Multi-step installer with progress tracking
    steps := []utils.InstallationStep{
        {Name: "Download", Description: "Downloading AWS CLI installer"},
        {Name: "Install", Description: "Installing AWS CLI"},
    }
    installer := utils.NewMultiStepInstaller(steps)
    
    // Step 1: Download MSI
    installer.StartStep(0)
    msiURL := "https://awscli.amazonaws.com/AWSCLIV2.msi"
    downloadFileWithProgress(msiURL, msiPath, "AWS CLI installer")
    installer.FinishStep()
    
    // Step 2: Silent installation
    installer.StartStep(1)
    cmd := exec.Command("msiexec.exe", "/i", msiPath, "/quiet", "/norestart")
    cmd.Run()
    installer.FinishStep()
}
```

**macOS:**
```go
func installAWSCLIMacOS() error {
    // Uses PKG installer with sudo
    pkgURL := "https://awscli.amazonaws.com/AWSCLIV2.pkg"
    cmd := exec.Command("sudo", "installer", "-pkg", pkgPath, "-target", "/")
    cmd.Stdout = os.Stdout  // Interactive sudo prompt
    cmd.Stderr = os.Stderr
    return cmd.Run()
}
```

**Linux:**
```go
func installAWSCLILinux() error {
    // Architecture detection
    arch := "x86_64"
    if runtime.GOARCH == "arm64" {
        arch = "aarch64"
    }
    
    // Universal ZIP installation
    zipURL := fmt.Sprintf("https://awscli.amazonaws.com/awscli-exe-linux-%s.zip", arch)
    // Download → Extract → Install with sudo
}
```

#### Google Cloud CLI Installation

**Package Manager Priority:**
1. **Debian/Ubuntu**: APT with official Google repository
2. **RedHat/CentOS/Fedora**: DNF/YUM with official Google repository  
3. **Universal**: Install script fallback

**Debian Installation:**
```go
func installGoogleCloudCLIDebian() error {
    commands := [][]string{
        {"apt-get", "update"},
        {"apt-get", "install", "-y", "apt-transport-https", "ca-certificates", "gnupg", "curl"},
        {"bash", "-c", "curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg"},
        {"bash", "-c", `echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list`},
        {"apt-get", "update"},
        {"apt-get", "install", "-y", "google-cloud-cli"},
    }
    // Execute with multi-step progress tracking
}
```

#### Azure CLI Installation

**Installation Priority:**
1. **Windows**: winget → MSI fallback
2. **macOS**: Homebrew → install script fallback
3. **Linux**: Package manager → install script fallback

**macOS with Homebrew:**
```go
func installAzureCLIMacOS() error {
    // Try Homebrew first
    if commandExists("brew") {
        spinner := utils.NewInstallSpinner("Installing Azure CLI via Homebrew...")
        cmd := exec.Command("brew", "install", "azure-cli")
        if err := cmd.Run(); err == nil {
            spinner.Finish()
            return nil
        }
    }
    
    // Fallback to install script
    cmd := exec.Command("bash", "-c", "curl -L https://aka.ms/InstallAzureCli | bash")
    return cmd.Run()
}
```

### Progress Indication System

CloudDeploy provides sophisticated progress feedback during installations:

#### Download Progress Bars

```go
func downloadFileWithProgress(url, filepath, description string) error {
    // Get file size for progress calculation
    resp, err := http.Head(url)
    fileSize := resp.ContentLength
    
    // Create progress bar with download-specific features
    bar := utils.NewDownloadProgressBar(fileSize, description)
    progressWriter := utils.NewProgressWriter(out, bar)
    
    // Stream download with real-time progress
    _, err = io.Copy(progressWriter, resp.Body)
    return err
}
```

#### Multi-Step Installation Progress

```go
type MultiStepInstaller struct {
    steps       []InstallationStep
    currentStep int
    spinner     *progressbar.ProgressBar
}

func (msi *MultiStepInstaller) StartStep(stepIndex int) {
    step := msi.steps[stepIndex]
    description := fmt.Sprintf("[%d/%d] %s", stepIndex+1, len(msi.steps), step.Description)
    msi.spinner = NewInstallSpinner(description)
}

func (msi *MultiStepInstaller) FinishStep() {
    msi.spinner.Finish()
    step := msi.steps[msi.currentStep]
    fmt.Printf("✅ %s completed!\n", step.Name)
}
```

**Example Output:**
```bash
🔧 [1/2] Downloading AWS CLI installer 100% [██████████████████████] (5.2 MB/s) [3s]
✅ Download completed!
🔧 [2/2] Installing AWS CLI  [spinning dots]
✅ Install completed!  
🎉 Installation completed successfully!
```

## Deployment Workflow

The deployment workflow orchestrates the entire process from configuration to live infrastructure.

### Unified Start Command

The `start` command in `cmd/start.go` implements the one-command workflow:

```go
func runStart(cmd *cobra.Command) error {
    // Phase 1: Project Analysis & Configuration
    analysis, err := project.AnalyzeProject(currentDir)
    projectName := getProjectName(analysis)
    appType := confirmOrSelectAppType(analysis)
    cloudProvider := selectCloudProvider()
    region := selectRegion(cloudProvider)
    services := configureServices()
    
    // Phase 2: Authentication
    if err := providers.CheckAuthentication(providerType); err != nil {
        return fmt.Errorf("authentication failed: %w", err)
    }
    
    // Phase 3: Configuration Persistence
    cfg := createConfig(projectName, appType, cloudProvider, region, services)
    if err := config.SaveConfig(cfg); err != nil {
        return err
    }
    
    // Phase 4: Conditional Deployment
    configOnly, _ := cmd.Flags().GetBool("config-only")
    if configOnly {
        return nil // Stop here
    }
    
    // Phase 5: Automatic Deployment
    autoApprove, _ := cmd.Flags().GetBool("auto-approve")
    return runDeploymentFromStart(cfg, autoApprove)
}
```

### Deployment Execution

The deployment process in `runDeploymentFromStart`:

```go
func runDeploymentFromStart(cfg *config.DeploymentConfig, autoApprove bool) error {
    // Step 1: Re-verify authentication
    if err := providers.CheckAuthentication(cfg.CloudProvider); err != nil {
        return fmt.Errorf("authentication failed: %w", err)
    }
    
    // Step 2: Re-analyze project (detect changes)
    detectionResult, err := project.AnalyzeProject(".")
    if detectionResult.AppType != cfg.AppType {
        // Handle app type conflicts with user input
        if detectionResult.IsHighConfidence() && !autoApprove {
            useDetected, _ := utils.PromptYesNo("Use detected app type instead?")
            if useDetected {
                cfg.AppType = detectionResult.AppType
                config.SaveConfig(cfg) // Update configuration
            }
        }
    }
    
    // Step 3: Generate Terraform Configuration
    terraformDir := terraform.TerraformWorkingDir
    os.MkdirAll(terraformDir, 0755)
    
    if err := terraform.GenerateConfigInDir(cfg, terraformDir); err != nil {
        return fmt.Errorf("failed to generate Terraform config: %w", err)
    }
    
    // Step 4: Initialize Terraform
    if err := terraform.InitInDir(terraformDir); err != nil {
        return fmt.Errorf("terraform init failed: %w", err)
    }
    
    // Step 5: Plan Infrastructure Changes
    planOutput, err := terraform.PlanInDir(terraformDir)
    if err != nil {
        return fmt.Errorf("terraform plan failed: %w", err)
    }
    fmt.Println(planOutput) // Show plan to user
    
    // Step 6: User Confirmation (unless auto-approved)
    if !autoApprove {
        approve, _ := utils.PromptYesNo("Do you want to apply these changes?")
        if !approve {
            return nil // User cancelled
        }
    }
    
    // Step 7: Apply Infrastructure Changes
    applyOutput, err := terraform.ApplyInDir(terraformDir)
    if err != nil {
        return fmt.Errorf("terraform apply failed: %w", err)
    }
    fmt.Println(applyOutput)
    
    // Step 8: Update Configuration & Display Results
    cfg.LastDeployment = time.Now().Format(time.RFC3339)
    config.SaveConfig(cfg)
    
    outputs, _ := terraform.GetOutputsInDir(terraformDir)
    displayDeploymentOutputs(cfg, outputs)
    
    return nil
}
```

### Terraform CLI Integration

CloudDeploy executes Terraform commands directly via the CLI:

```go
func InitInDir(dir string) error {
    cmd := exec.Command("terraform", "init")
    cmd.Dir = dir
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("terraform init failed: %w\nOutput: %s", err, string(output))
    }
    return nil
}

func PlanInDir(dir string) (string, error) {
    cmd := exec.Command("terraform", "plan", "-out=tfplan")
    cmd.Dir = dir
    output, err := cmd.CombinedOutput()
    return string(output), err
}

func ApplyInDir(dir string) (string, error) {
    cmd := exec.Command("terraform", "apply", "-auto-approve", "tfplan")
    cmd.Dir = dir
    output, err := cmd.CombinedOutput()
    return string(output), err
}
```

### Output Processing

CloudDeploy extracts and formats Terraform outputs:

```go
func GetOutputsInDir(dir string) (map[string]interface{}, error) {
    cmd := exec.Command("terraform", "output", "-json")
    cmd.Dir = dir
    output, err := cmd.Output()
    
    var outputs map[string]interface{}
    json.Unmarshal(output, &outputs)
    
    // Extract values from Terraform's nested format
    result := make(map[string]interface{})
    for key, value := range outputs {
        if outputMap, ok := value.(map[string]interface{}); ok {
            if val, exists := outputMap["value"]; exists {
                result[key] = val // Extract actual value
            }
        }
    }
    return result, nil
}
```

## Configuration Management

CloudDeploy uses a JSON-based configuration system that persists deployment settings.

### Configuration Structure

```go
type DeploymentConfig struct {
    ProjectName    string        `json:"project_name"`
    AppType        AppType       `json:"app_type"`        // nodejs, flask, docker, static-site
    CloudProvider  CloudProvider `json:"cloud_provider"`  // aws, gcp, azure
    Region         string        `json:"region"`
    ImageName      string        `json:"image_name,omitempty"`
    Services       Services      `json:"services"`
    CreatedAt      string        `json:"created_at"`
    LastDeployment string        `json:"last_deployment,omitempty"`
}

type Services struct {
    Database     bool `json:"database"`
    Storage      bool `json:"storage"`
    LoadBalancer bool `json:"load_balancer"`
}
```

### Configuration Persistence

The configuration is stored as `.clouddeploy.json` in the project directory:

```go
func SaveConfig(cfg *DeploymentConfig) error {
    cfg.CreatedAt = time.Now().Format(time.RFC3339)
    
    data, err := json.MarshalIndent(cfg, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(".clouddeploy.json", data, 0644)
}

func LoadConfig() (*DeploymentConfig, error) {
    data, err := os.ReadFile(".clouddeploy.json")
    if err != nil {
        return nil, err
    }
    
    var cfg DeploymentConfig
    err = json.Unmarshal(data, &cfg)
    return &cfg, err
}
```

### Configuration Validation

The system validates configuration consistency:

```go
func ValidateProjectName(input string) error {
    input = strings.TrimSpace(input)
    if len(input) == 0 {
        return fmt.Errorf("project name cannot be empty")
    }
    if len(input) < 3 || len(input) > 50 {
        return fmt.Errorf("project name must be 3-50 characters")
    }
    
    // Validate characters (alphanumeric, hyphens, underscores only)
    for _, r := range input {
        if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
             (r >= '0' && r <= '9') || r == '-' || r == '_') {
            return fmt.Errorf("project name can only contain letters, numbers, hyphens, and underscores")
        }
    }
    return nil
}
```

## User Interface and Progress Indication

CloudDeploy provides a rich, interactive user experience with beautiful progress indicators and clear feedback.

### Interactive Prompts

The UI system in `internal/utils/prompt.go` uses the `promptui` library:

```go
func PromptSelect(label string, items []string) (string, error) {
    prompt := promptui.Select{
        Label: label,
        Items: items,
    }
    _, result, err := prompt.Run()
    return result, err
}

func PromptYesNo(label string) (bool, error) {
    prompt := promptui.Select{
        Label: label,
        Items: []string{"Yes", "No"},
    }
    _, result, err := prompt.Run()
    return result == "Yes", nil
}
```

### Progress Indicators

The progress system in `internal/utils/progress.go` provides multiple types of feedback:

#### Download Progress Bars

```go
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
```

#### Installation Spinners

```go
func NewInstallSpinner(description string) *progressbar.ProgressBar {
    return progressbar.NewOptions(-1,
        progressbar.OptionSetDescription(fmt.Sprintf("🔧 %s", description)),
        progressbar.OptionSetWidth(50),
        progressbar.OptionSpinnerType(11), // Modern dots spinner
        progressbar.OptionEnableColorCodes(true),
        progressbar.OptionShowElapsedTimeOnFinish(),
    )
}
```

#### Multi-Step Progress

```go
type MultiStepInstaller struct {
    steps       []InstallationStep
    currentStep int
    spinner     *progressbar.ProgressBar
}

// Provides coordinated progress across multiple installation steps
// with clear visual feedback and completion notifications
```

### Logging System

CloudDeploy uses structured logging with `logrus`:

```go
// internal/logger/logger.go
func Init() {
    logrus.SetFormatter(&logrus.TextFormatter{
        FullTimestamp: true,
        DisableColors: false,
    })
    logrus.SetLevel(logrus.InfoLevel)
}

func Info(args ...interface{}) {
    logrus.Info(args...)
}

func Infof(format string, args ...interface{}) {
    logrus.Infof(format, args...)
}
```

**Log Levels:**
- **Debug**: Internal operations, API calls
- **Info**: User-visible progress and status
- **Warn**: Non-fatal issues and fallbacks
- **Error**: Failures requiring user attention
- **Fatal**: Unrecoverable errors (exits process)

## Build System and Distribution

CloudDeploy uses a sophisticated Makefile-based build system that supports multiple targets and deployment scenarios.

### Build Targets

The `Makefile` provides comprehensive build options:

```make
# Development builds (with debug symbols)
build-dev: deps
	CGO_ENABLED=0 go build \
		-ldflags="$(LDFLAGS_DEV)" \
		-o $(BUILD_DIR)/$(BINARY_NAME) .

# Production builds (optimized, stripped)
build-production: deps  
	CGO_ENABLED=0 go build \
		-ldflags="$(LDFLAGS_PROD) -s -w" \
		-a -installsuffix cgo \
		-o $(BUILD_DIR)/$(BINARY_NAME) .

# Multi-platform builds
build-all: deps
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS_PROD) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS_PROD) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS_PROD) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS_PROD) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .
```

### Build Configuration

```make
# Build flags
LDFLAGS_BASE := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)
LDFLAGS_DEV := $(LDFLAGS_BASE)
LDFLAGS_PROD := $(LDFLAGS_BASE) -s -w -extldflags "-static"

# Security flags
CGO_ENABLED=0          # Disable CGO for static builds
-a                     # Force rebuilding of packages
-installsuffix cgo     # Add suffix to distinguish from CGO builds  
-s -w                  # Strip symbol table and debug info
-extldflags "-static"  # Static linking
```

### Quality Assurance

```make
# Comprehensive testing
test-all: test test-race test-coverage
	@echo "$(GREEN)✅ All tests passed$(NC)"

# Security scanning
security:
	gosec ./...                                    # Security vulnerabilities
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...  # Dependency vulnerabilities

# Code quality
lint:
	golangci-lint run --config .golangci.yml

# Formatting
format:
	gofmt -s -w .
	goimports -w .
```

### Dependency Management

```make
# Dependency operations
deps:
	go mod download
	go mod verify

deps-update:
	go get -u all
	go mod tidy

tidy:
	go mod tidy
	go mod verify
```

## Security Considerations

CloudDeploy implements multiple security layers to protect user credentials and infrastructure.

### Authentication Security

1. **No Credential Storage**: CloudDeploy never stores cloud provider credentials
2. **CLI Delegation**: Authentication is delegated to official cloud CLI tools
3. **Session Reuse**: Leverages existing authenticated sessions when available
4. **Automatic Expiration**: Respects cloud provider session timeouts

### Input Validation

```go
// Project name validation prevents injection attacks
func ValidateProjectName(input string) error {
    // Strict alphanumeric + hyphen/underscore validation
    for _, r := range input {
        if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
             (r >= '0' && r <= '9') || r == '-' || r == '_') {
            return fmt.Errorf("project name can only contain letters, numbers, hyphens, and underscores")
        }
    }
}
```

### File System Security

1. **Temporary File Cleanup**: All downloaded files are cleaned up automatically
2. **Permission Control**: Generated files use restrictive permissions (0644, 0755)
3. **Directory Traversal Prevention**: Path validation prevents directory traversal attacks

```go
// Zip extraction with path validation
func extractZip(src, dest string) error {
    for _, file := range reader.File {
        path := filepath.Join(dest, file.Name)
        
        // Prevent ZipSlip vulnerability
        if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
            return fmt.Errorf("invalid file path: %s", file.Name)
        }
    }
}
```

### Build Security

1. **Static Builds**: CGO disabled to prevent dynamic library vulnerabilities
2. **Symbol Stripping**: Production builds strip debug symbols and symbol tables
3. **Dependency Scanning**: Automated vulnerability scanning with `govulncheck`
4. **Code Analysis**: Security scanning with `gosec`

### Network Security

1. **HTTPS Only**: All downloads use HTTPS with certificate validation
2. **Official Sources**: CLI tools downloaded only from official vendor URLs
3. **Checksum Validation**: Where available, downloads are checksum-verified

### Terraform State Security

1. **Local State Management**: State files remain local unless user configures remote backend
2. **Sensitive Output Handling**: Terraform outputs marked sensitive are handled appropriately
3. **Workspace Isolation**: Each project uses its own Terraform workspace

---

## Summary

CloudDeploy is a sophisticated infrastructure deployment tool that combines intelligent project detection, template-based Terraform generation, enterprise-grade authentication, and beautiful user experience. Its modular architecture enables extension to new cloud providers and application types while maintaining consistency and reliability.

The system's key strengths include:

- **Intelligence**: Sophisticated project auto-detection with confidence scoring
- **Automation**: One-command workflow from detection to deployment
- **Enterprise**: Bloomberg-optimized authentication with automatic CLI installation
- **Security**: Multi-layered security with no credential storage
- **UX**: Beautiful progress indicators and interactive prompts
- **Reliability**: Comprehensive error handling and fallback strategies

The codebase demonstrates modern Go development practices with clean architecture, comprehensive testing, and production-ready build systems. 