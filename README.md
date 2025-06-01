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
- **Beautiful Progress Indicators**: Real-time download progress bars and installation spinners
- **Multi-Step Progress**: Visual feedback for complex installation processes

### 🎯 **User Experience**
- **Interactive Setup**: Guided configuration with smart defaults
- **Progress Feedback**: Real-time deployment status updates with beautiful progress bars
- **Download Progress**: Shows download speed, file size, and estimated time remaining
- **Installation Spinners**: Elegant spinners for background operations
- **Step-by-Step Visual**: Multi-step installation progress with clear completion indicators
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

# Beautiful installation progress with real-time feedback:
📥 [1/2] Downloading AWS CLI installer 100% [██████████████████████] (5.2 MB/s) [3s]
✅ Download completed!
🔧 [2/2] Installing AWS CLI  [0s]
✅ Install completed!
🎉 Installation completed successfully!

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

# Multi-step installation with progress tracking:
🔧 [1/5] Updating package lists  [0s]
✅ Update completed!
🔧 [2/5] Installing dependencies  [0s]  
✅ Dependencies completed!
🔧 [3/5] Adding Google Cloud signing key  [0s]
✅ Key completed!
🔧 [4/5] Adding Google Cloud repository  [0s]
✅ Repository completed!
🔧 [5/5] Installing Google Cloud CLI  [0s]
✅ Install completed!
🎉 Installation completed successfully!

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

---

# 🔐 Bloomberg Enterprise Authentication

CloudDeploy includes **automatic authentication** specifically optimized for Bloomberg's enterprise environment! This eliminates the need to manually run authentication commands.

## 🎯 What's New for Bloomberg Employees

Instead of seeing error messages like:
```
ERRO Authentication check failed: not authenticated with Azure. Please run 'az login'
```

CloudDeploy now **automatically initiates the login process** for you with Bloomberg-optimized settings.

## 🏢 Bloomberg Environment Detection

CloudDeploy automatically detects if you're in a Bloomberg environment by checking:
- Hostname containing "bloomberg"
- Username containing "corp" 
- Domain containing "bloomberg"
- `BLOOMBERG_ENV` environment variable

## ☁️ Cloud Provider Support for Bloomberg

### Azure (Recommended for Bloomberg)
When you need to authenticate with Azure, CloudDeploy will:

1. **Detect Bloomberg environment** 
2. **Use device code flow** (better for enterprise SSO)
3. **Prompt for B-Unit readiness**
4. **Launch**: `az login --use-device-code --tenant common`
5. **Fallback** to: `az login --allow-no-subscriptions` if needed

**Example flow:**
```
🔐 Not authenticated with Azure. Initiating automatic login...
🚀 Starting Azure authentication flow...
💼 Detected Bloomberg environment. Using enterprise SSO authentication...
🔑 Starting Bloomberg Azure SSO authentication...
📱 Please have your B-Unit or B-Unit phone app ready for 2FA
🖥️  Using device code flow for enterprise compatibility...
```

### AWS (Enterprise SSO)
For AWS authentication, CloudDeploy will:

1. **Try AWS SSO first**: `aws sso login`
2. **Fallback to standard**: `aws configure` if SSO fails

### Google Cloud Platform
For GCP authentication, CloudDeploy will:

1. **Use enterprise-friendly options**: `gcloud auth login --enable-gdrive-access --brief`
2. **Integrate with CORP credentials**

## 📱 What Bloomberg Employees Need

Make sure you have:

- **CORP credentials** (same username/password as your PC login)
- **B-Unit device** or **B-Unit phone app** for 2FA authentication
- **Access to the cloud provider** through Bloomberg's enterprise setup

## 🔄 Authentication Flow Comparison

### Before (Manual Process)
```bash
clouddeploy init
> ERRO Authentication check failed: not authenticated with Azure. Please run 'az login'
> INFO Please authenticate with your cloud provider and try again.
> FATA Initialization failed: not authenticated with Azure. Please run 'az login'

# You had to manually run:
az login
# Then run clouddeploy again
clouddeploy init
```

### After (Automatic Process)
```bash
clouddeploy init
> 🔐 Not authenticated with Azure. Initiating automatic login...
> 🚀 Starting Azure authentication flow...
> 💼 Detected Bloomberg environment. Using enterprise SSO authentication...
> 🔑 Starting Bloomberg Azure SSO authentication...
> 📱 Please have your B-Unit or B-Unit phone app ready for 2FA
> 🖥️  Using device code flow for enterprise compatibility...
> [Authentication flow opens automatically]
> ✅ Bloomberg Azure authentication successful!
> [Continues with clouddeploy init...]
```

---

# 🔧 Troubleshooting

## Common Issues

### CLI Installation Issues
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

### Authentication Issues

#### For Bloomberg Employees
CloudDeploy now **automatically handles authentication** for you! 🎉

When you run `clouddeploy init` or `clouddeploy deploy`, the tool will:
1. **Detect your Bloomberg environment** automatically
2. **Launch the appropriate SSO flow** for your cloud provider
3. **Guide you through the process** with Bloomberg-specific instructions

#### Authentication Failures
If automatic authentication fails, you can still authenticate manually:

```bash
# Azure
az login

# AWS  
aws sso login
# or
aws configure

# GCP
gcloud auth login
```

#### Environment Variables for Bloomberg

You can force Bloomberg environment detection by setting:
```bash
export BLOOMBERG_ENV=true
```

### Platform-Specific Notes
- **Windows**: Requires PowerShell or Command Prompt with admin privileges for installation
- **macOS**: May require `sudo` password for CLI installation
- **Linux**: Supports both package managers (apt, dnf, yum) and universal install scripts

### Getting Help
```bash
clouddeploy --help           # General help
clouddeploy init --help      # Init command help
clouddeploy deploy --help    # Deploy command help
```

---

# 📋 Changelog

## [v2.0.0] - 2024-12-XX - Bloomberg Enterprise Authentication

### 🎉 Major New Features

#### **Automatic Authentication for Bloomberg Employees**
- **No more manual authentication steps!** CloudDeploy now automatically handles authentication for Bloomberg employees
- **Smart environment detection** - Automatically detects Bloomberg corporate environment
- **Enterprise SSO optimization** - Uses the best authentication methods for Bloomberg's SSO setup

### ✨ New Features

#### **Bloomberg Environment Detection**
- Automatically detects Bloomberg environment by checking:
  - Hostname containing "bloomberg"
  - Username containing "corp"
  - Domain containing "bloomberg" 
  - `BLOOMBERG_ENV` environment variable

#### **Azure Authentication (Optimized for Bloomberg)**
- **Device code flow** for better enterprise SSO compatibility
- **Automatic tenant handling** with fallback options
- **B-Unit integration prompts** to prepare users for 2FA
- Commands used: `az login --use-device-code --tenant common`

#### **AWS Authentication (Enterprise SSO)**
- **AWS SSO first** - Tries `aws sso login` before standard methods
- **Automatic fallback** to `aws configure` if SSO fails
- **Bloomberg enterprise setup aware**

#### **GCP Authentication (Enterprise-friendly)**
- **Enhanced login options** with `--enable-gdrive-access --brief`
- **CORP credential integration**
- **Enterprise authentication flow**

### 🔧 Improvements

#### **Better Error Messages**
- Instead of: `"Please run 'az login'"`
- Now shows: `"The authentication process was attempted automatically. If you're a Bloomberg employee, make sure your CORP credentials and B-Unit are ready"`

#### **User Experience**
- **Proactive 2FA prompts** - Warns users to have B-Unit ready
- **Real-time feedback** during authentication process
- **Clear Bloomberg-specific guidance**
- **Automatic retry mechanisms** with fallback options

#### **Non-Bloomberg Compatibility**
- **Standard authentication** still works for non-Bloomberg users
- **Automatic detection** - No configuration needed
- **Manual override** available with environment variables

### 📝 Updated Files
- `internal/providers/providers.go` - Complete authentication overhaul
- `cmd/init.go` - Updated error messaging
- `cmd/deploy.go` - Updated error messaging  
- `README.md` - Enhanced troubleshooting section

### 🧪 Testing
- Verified build compatibility
- Maintained backward compatibility for non-Bloomberg users

### 💡 Usage Examples

#### Before (Manual)
```bash
clouddeploy init
# ERROR: not authenticated with Azure. Please run 'az login'
az login
clouddeploy init
```

#### After (Automatic)
```bash
clouddeploy init
# 🔐 Not authenticated with Azure. Initiating automatic login...
# 💼 Detected Bloomberg environment...
# [Authentication happens automatically]
# ✅ Bloomberg Azure authentication successful!
```

### 🔗 Related
- Based on Bloomberg NEXI documentation for enterprise SSO
- Integrates with Bloomberg's CORP credential system
- Supports B-Unit and B-Unit phone app for 2FA 

---

**CloudDeploy CLI** - Intelligent Multi-Cloud Deployment for the Modern Enterprise