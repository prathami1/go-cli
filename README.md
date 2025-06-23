# 🚀 CloudDeploy

**The simplest way to deploy cloud applications**

CloudDeploy is a powerful command-line tool that simplifies cloud deployment across AWS, Google Cloud Platform (GCP), and Microsoft Azure. With automatic CLI installation, streamlined authentication, and intelligent project analysis, you can go from code to cloud in minutes.

## ✨ Key Features

- 🔧 **Auto-Install Everything**: Automatically installs Terraform and cloud CLI tools
- 🔐 **Seamless Authentication**: Handles cloud provider authentication flows 
- 🎯 **Intelligent Analysis**: Analyzes your project structure and suggests optimal deployment configurations
- ☁️ **Multi-Cloud Support**: Deploy to AWS, GCP, or Azure with the same simple commands
- 📦 **Zero Configuration**: Works out of the box with sensible defaults
- 🚀 **Lightning Fast**: Get from idea to deployed application in under 5 minutes

## 📦 Installation

### Prerequisites
- Go 1.19+ 
- **All required tools are automatically installed if missing!**
  - **Terraform**: CloudDeploy will detect and install Terraform automatically if not found
  - **Cloud provider CLI tools**: aws-cli, gcloud, or az-cli are automatically installed as needed

### Quick Install

```bash
# Clone and build
git clone https://github.com/yourusername/clouddeploy
cd clouddeploy
go build -o bin/cdeploy .

# Add to PATH (optional)
export PATH=$PATH:$(pwd)/bin
```

### Or Download Binary

Download the latest release from the [releases page](https://github.com/yourusername/clouddeploy/releases).

## 🚀 Quick Start

### 1. Initialize Your Project

```bash
cdeploy start
```

This command will:
- 📊 **Analyze your project** structure automatically
- 🏗️ **Suggest deployment configuration** based on your code
- ☁️ **Help you choose** the best cloud provider for your needs
- 🔐 **Handle authentication** automatically
- 📝 **Generate Terraform files** ready for deployment

### 2. Deploy to the Cloud

```bash
cdeploy deploy
```

This command will:
- 🔍 **Validate your configuration** 
- 🏗️ **Plan your infrastructure** with Terraform
- 🚀 **Deploy your application** to your chosen cloud provider
- 📊 **Show deployment status** and URLs

### 3. Clean Up Resources

```bash
cdeploy destroy
```

## 💡 How It Works

1. **Project Analysis**: CloudDeploy scans your project to understand its structure and requirements
2. **Smart Configuration**: Based on your project type, it suggests the optimal cloud resources
3. **Auto-Authentication**: Handles the complex authentication flows for cloud providers
4. **Infrastructure as Code**: Generates Terraform configurations for reproducible deployments
5. **One-Click Deploy**: Applies the infrastructure and deploys your application

## 🛠️ Supported Project Types

| Project Type | AWS | GCP | Azure | Auto-Detection |
|-------------|-----|-----|-------|----------------|
| **Node.js/React** | ✅ | ✅ | ✅ | ✅ |
| **Python/Django** | ✅ | ✅ | ✅ | ✅ |
| **Go** | ✅ | ✅ | ✅ | ✅ |
| **Static Sites** | ✅ | ✅ | ✅ | ✅ |
| **Docker** | ✅ | ✅ | ✅ | ✅ |

## ☁️ Cloud Provider Support

### Amazon Web Services (AWS)
- **Compute**: EC2, ECS, Lambda
- **Storage**: S3
- **Database**: RDS (PostgreSQL, MySQL)
- **Load Balancer**: Application Load Balancer
- **Authentication**: AWS CLI with IAM

### Google Cloud Platform (GCP)
- **Compute**: Compute Engine, Cloud Run, Cloud Functions
- **Storage**: Cloud Storage
- **Database**: Cloud SQL (PostgreSQL, MySQL)
- **Load Balancer**: Cloud Load Balancing
- **Authentication**: gcloud CLI with service accounts

### Microsoft Azure
- **Compute**: Virtual Machines, Container Instances, Functions
- **Storage**: Blob Storage
- **Database**: Azure Database (PostgreSQL, MySQL)
- **Load Balancer**: Azure Load Balancer
- **Authentication**: Azure CLI with service principals

## 🔐 Authentication

CloudDeploy automatically handles authentication for all supported cloud providers:

### AWS
```bash
# CloudDeploy will automatically run:
aws configure
# or guide you through SSO setup
```

### Google Cloud Platform
```bash
# CloudDeploy will automatically run:
gcloud auth login
```

### Microsoft Azure
```bash
# CloudDeploy will automatically run:
az login
```

## 📖 Detailed Usage

### Project Structure Detection

CloudDeploy automatically detects your project type by examining:
- Package files (`package.json`, `requirements.txt`, `go.mod`)
- Framework markers (React, Django, Flask, etc.)
- Docker files
- Build configurations

### Configuration Options

When running `cdeploy start`, you can specify:
- **Cloud Provider**: AWS, GCP, or Azure
- **Region**: Choose from available regions
- **Services**: Database, storage, load balancer
- **Environment**: Development, staging, production

### Environment Variables

You can customize CloudDeploy behavior with environment variables:

```bash
# Force specific cloud provider
export CLOUDDEPLOY_PROVIDER=aws

# Set default region
export CLOUDDEPLOY_REGION=us-west-2

# Enable debug logging
export CLOUDDEPLOY_DEBUG=true
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

#### Authentication Issues

If automatic authentication fails, you can authenticate manually:

```bash
# AWS
aws configure

# GCP
gcloud auth login

# Azure
az login
```

#### Terraform Issues

If you encounter Terraform-related errors:

```bash
# Check Terraform installation
terraform --version

# If CloudDeploy's auto-install failed, install manually:
# https://learn.hashicorp.com/tutorials/terraform/install-cli
```

### Platform-Specific Notes
- **Windows**: Requires PowerShell or Command Prompt with admin privileges for installation
- **macOS**: May require `sudo` password for CLI installation
- **Linux**: Supports both package managers (apt, dnf, yum) and universal install scripts

### Getting Help
```bash
cdeploy --help           # General help
cdeploy start --help     # Start command help
cdeploy deploy --help    # Deploy command help
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/yourusername/clouddeploy
cd clouddeploy

# Install dependencies
go mod download

# Build
go build -o bin/cdeploy .

# Run tests
go test ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Terraform for infrastructure as code
- AWS, GCP, and Azure for cloud platforms
- The Go community for excellent tooling

---

**Made with ❤️ for developers who want to deploy fast and focus on building great applications.** 