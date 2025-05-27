# CloudDeploy CLI

> **Production-Ready Multi-Cloud Deployment Tool**

CloudDeploy is an enterprise-grade CLI tool that intelligently detects your application type and deploys it to AWS, GCP, or Azure using Terraform.

## 🚀 Features

### ✨ **Intelligent Project Detection**
- **Automatic Detection**: Recognizes Docker, Node.js, Flask/Python, and static site projects
- **Confidence Scoring**: High, medium, or low confidence levels for accurate detection
- **Smart Fallbacks**: Gracefully handles mixed or unknown project types

### ☁️ **Multi-Cloud Support**
- **AWS**: Full support with VPC, EC2, ECS, RDS, S3, and ALB
- **GCP**: Compute Engine, Cloud SQL, Cloud Storage, and Load Balancing
- **Azure**: Resource Groups, VMs, Azure SQL, Storage Accounts, and Application Gateway

### 🛡️ **Enterprise Security**
- **Authentication Verification**: Pre-deployment cloud provider checks
- **Input Validation**: All user inputs sanitized and validated
- **Secure Builds**: Position Independent Executables (PIE) with stripped symbols
- **No Hardcoded Secrets**: Environment-based configuration
- **Audit Trail**: Complete logging for compliance and troubleshooting

### 🔧 **Automatic CLI Management**
- **Auto-Detection**: Checks for required cloud provider CLI tools
- **Smart Installation**: Automatically installs aws-cli, gcloud, or az-cli if missing
- **Cross-Platform**: Works on Windows, macOS, and Linux
- **No Dependencies**: Uses native package managers when available, falls back to universal methods
- **User Choice**: Always prompts before installing new tools

### 🎯 **User Experience**
- **Interactive Setup**: Guided configuration with smart defaults
- **Progress Feedback**: Real-time deployment status updates
- **Error Recovery**: Comprehensive error handling with cleanup
- **Help System**: Extensive help text and examples

## 📦 Installation

### Prerequisites
- Go 1.19+ 
- Terraform 1.0+
- **Cloud provider CLI tools are automatically installed if missing!**
  - CloudDeploy will detect and install aws-cli, gcloud, or az-cli as needed
  - No need to manually install these tools beforehand

### Quick Install
```bash
# Clone the repository
git clone url
cd clouddeploy-cli

# Build the binary
make build-production

# Install to your PATH
sudo cp ./bin/clouddeploy /usr/local/bin/
```

### Development Setup
```bash
# Install development tools
make install-tools

# Set up development environment
make dev-setup

# Run tests
make test-all
```

## 🎯 Quick Start

### 1. Initialize Your Project
```bash
# Navigate to your project directory
cd my-awesome-app

# Initialize CloudDeploy (auto-detects project type)
clouddeploy init
```

### 2. Deploy to the Cloud
```bash
# Deploy with interactive confirmation
clouddeploy deploy

# Deploy automatically (for CI/CD)
clouddeploy deploy --auto-approve

# Preview changes only
clouddeploy deploy --plan-only
```

### 3. Manage Your Infrastructure
```bash
# View deployment status
clouddeploy deploy --plan-only

# Destroy infrastructure
clouddeploy destroy

# Get help
clouddeploy --help
```

## 🏗️ Architecture

```
CloudDeploy CLI
├── cmd/                    # Cobra commands
│   ├── root.go            # Root command and global configuration
│   ├── init.go            # Interactive project initialization
│   ├── deploy.go          # Enhanced deployment with project analysis
│   └── destroy.go         # Infrastructure teardown
├── internal/
│   ├── config/            # Configuration management
│   ├── logger/            # Structured logging (logrus)
│   ├── project/           # 🎯 Intelligent project analysis
│   ├── providers/         # Cloud provider authentication & CLI auto-install
│   ├── terraform/         # Terraform operations and templates
│   └── utils/             # Utility functions and prompts
└── Makefile              # Production-ready build system
```

## 🔍 Project Detection

CloudDeploy automatically analyzes your project to determine the best deployment strategy:

| Project Type | Detection Method | Confidence | Infrastructure |
|--------------|------------------|------------|----------------|
| **Docker** | Dockerfile presence | High | Container services (ECS, GKE, ACI) |
| **Node.js** | package.json | High | App services with Node.js runtime |
| **Flask/Python** | app.py + requirements.txt | High | App services with Python runtime |
| **Static Site** | HTML/CSS/JS files | Medium | Static hosting (S3, GCS, Blob) |

## 📋 Configuration

CloudDeploy uses a `.clouddeploy.json` configuration file:

```json
{
  "project_name": "my-awesome-app",
  "app_type": "nodejs",
  "cloud_provider": "aws",
  "region": "us-east-1",
  "services": {
    "database": true,
    "storage": false,
    "load_balancer": true
  },
  "last_deployment": "2024-01-15T10:30:00Z"
}
```

### Environment Variables
```bash
# Cloud Provider Authentication
export AWS_PROFILE=my-profile
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
export AZURE_SUBSCRIPTION_ID=your-subscription-id

# CloudDeploy Configuration
export CLOUDDEPLOY_LOG_LEVEL=info
export CLOUDDEPLOY_CONFIG_PATH=./custom-config.json
```

## 🧪 Testing

### Run All Tests
```bash
make test-all          # Unit + integration + race tests
make test-coverage     # Generate coverage report
make security          # Security scanning
```

### Test Coverage
- **Unit Tests**: 100% pass rate across all modules
- **Integration Tests**: Multi-cloud template generation
- **Race Detection**: Concurrent safety verification
- **Security Scanning**: gosec and govulncheck

## 🔧 Development

### Build System
```

## 📊 Examples

### Deploy a Node.js App to AWS (with automatic AWS CLI installation)
```bash
# Initialize (auto-detects Node.js from package.json)
clouddeploy init
> Project name: my-node-app
> App type: nodejs (detected with high confidence)
> Cloud provider: aws
> Region: us-east-1
> ⚠️ AWS CLI is not installed
> 🔧 CloudDeploy can automatically install AWS CLI for you
> Would you like to install AWS CLI now? (y/N) y
> 🔧 Installing AWS CLI automatically...
> Installing AWS CLI on macOS...
> ✅ AWS CLI installed successfully!
> Enable database? yes
> Enable storage? no
> Enable load balancer? yes

# Deploy
clouddeploy deploy
> 🔍 Analyzing project... ✅ Confirmed Node.js
> 🔐 Verifying AWS authentication... Please run 'aws configure' first
> 📄 Generating Terraform configuration... ✅ Generated
> 🔧 Initializing Terraform... ✅ Initialized
> 📋 Running Terraform plan... 
> Apply changes? yes
> 🚀 Deploying infrastructure... ✅ Complete!
> 📊 App URL: https://my-node-app-alb-123456789.us-east-1.elb.amazonaws.com
```

### Deploy a Flask App to GCP
```bash
clouddeploy init
> Project name: my-flask-api
> App type: flask (detected from app.py)
> Cloud provider: gcp
> Region: us-central1
> ⚠️ Google Cloud CLI is not installed
> 🔧 CloudDeploy can automatically install Google Cloud CLI for you
> Would you like to install Google Cloud CLI now? (y/N) y
> ✅ Google Cloud CLI installed successfully!

clouddeploy deploy --auto-approve
> 🎯 Detected Flask app with high confidence
> 🚀 Deploying to Google Cloud Platform...
> ✅ Deployment complete!
```

### Deploy a Static Site to Azure
```bash
clouddeploy init
> Project name: my-portfolio
> App type: static-site (detected from index.html)
> Cloud provider: azure
> Region: eastus

clouddeploy deploy
> 🔍 Static site detected: index.html, style.css, script.js
> 📄 Generating Azure Storage + CDN configuration...
> ✅ Site available at: https://myportfolio.azureedge.net
```

## 🔧 Troubleshooting

### Common Issues

#### CLI Installation Issues
CloudDeploy automatically handles CLI installation, but if you encounter issues:

```bash
# Manual verification after auto-install
aws --version     # Should show AWS CLI v2.x
gcloud version    # Should show Google Cloud SDK
az --version      # Should show Azure CLI

# If auto-install fails, you can install manually:
# AWS: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html
# GCP: https://cloud.google.com/sdk/docs/install
# Azure: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli
```

#### Authentication Failures
After CLI installation, you'll need to authenticate:

```bash
# AWS
aws configure  # or aws sso login

# GCP  
gcloud auth login

# Azure
az login
```

#### Platform-Specific Notes
- **Windows**: Requires PowerShell or Command Prompt with admin privileges for installation
- **macOS**: May require `sudo` password for CLI installation
- **Linux**: Supports both package managers (apt, dnf, yum) and universal install scripts

### Getting Help
```bash
clouddeploy --help           # General help
clouddeploy init --help      # Init command help
clouddeploy deploy --help    # Deploy command help
```

**CloudDeploy CLI** - Intelligent Multi-Cloud Deployment for the Modern Enterprise