# CloudDeploy CLI

A powerful Go CLI tool built with Cobra that helps Bloomberg employees deploy applications to AWS, Azure, or GCP using Terraform under the hood.

## Features

- 🚀 **Multi-cloud Support**: Deploy to AWS, Azure, or GCP
- 🔧 **Multiple App Types**: Support for static sites, Node.js, Flask, and Docker applications
- 🔐 **Authentication Check**: Automatically verifies cloud provider CLI authentication
- 📄 **Terraform Generation**: Generates appropriate Terraform configurations based on your requirements
- 🎯 **Interactive Prompts**: User-friendly prompts for configuration
- 📊 **Deployment Outputs**: Shows important URLs, credentials, and endpoints after deployment
- 🗑️ **Easy Cleanup**: Simple infrastructure destruction with safety prompts

## Prerequisites

Before using CloudDeploy, ensure you have the following installed:

- **Go 1.21+** for building from source
- **Terraform** for infrastructure provisioning
- **Cloud Provider CLIs** (depending on your target):
  - AWS CLI (`aws`) - [Installation Guide](https://aws.amazon.com/cli/)
  - Google Cloud CLI (`gcloud`) - [Installation Guide](https://cloud.google.com/sdk/docs/install)
  - Azure CLI (`az`) - [Installation Guide](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli)

## Installation

### From Source

```bash
git clone <repository-url>
cd go-cli
go build -o clouddeploy .
```

### Using Go Install

```bash
go install github.com/prathami1/go-cli@latest
```

## Quick Start

1. **Initialize a new project**:
   ```bash
   ./clouddeploy init
   ```
   This will prompt you for:
   - Project name
   - Application type (static-site, nodejs, flask, docker)
   - Cloud provider (aws, gcp, azure)
   - Region
   - Optional services (database, storage, load balancer)

2. **Deploy your application**:
   ```bash
   ./clouddeploy deploy
   ```

3. **Destroy infrastructure when done**:
   ```bash
   ./clouddeploy destroy
   ```

## Commands

### `clouddeploy init`

Initialize a new deployment configuration. This command:
- Prompts for project details and preferences
- Checks cloud provider authentication
- Creates a `.clouddeploy.json` configuration file
- Validates all inputs

**Example**:
```bash
./clouddeploy init
```

### `clouddeploy deploy`

Deploy your application to the configured cloud provider. This command:
- Loads configuration from `.clouddeploy.json`
- Verifies cloud provider authentication
- Generates Terraform configuration files
- Runs `terraform init`, `plan`, and `apply`
- Displays deployment outputs (URLs, credentials, etc.)

**Flags**:
- `-y, --auto-approve`: Skip interactive approval of plan
- `-p, --plan-only`: Only run terraform plan, don't apply

**Examples**:
```bash
./clouddeploy deploy                    # Interactive deployment
./clouddeploy deploy --auto-approve     # Skip confirmation
./clouddeploy deploy --plan-only        # Only show plan
```

### `clouddeploy destroy`

Destroy the deployed infrastructure. This command:
- Loads configuration from `.clouddeploy.json`
- Runs `terraform destroy` to tear down all resources
- Optionally cleans up generated Terraform files

**Flags**:
- `-y, --auto-approve`: Skip interactive approval
- `-c, --cleanup`: Remove generated Terraform files after destroy

**Examples**:
```bash
./clouddeploy destroy                   # Interactive destruction
./clouddeploy destroy --auto-approve    # Skip confirmation
./clouddeploy destroy --cleanup         # Also remove Terraform files
```

## Supported Application Types

### Static Site
- Deploys to S3 (AWS), Cloud Storage (GCP), or Blob Storage (Azure)
- Configures static website hosting
- Sets up appropriate permissions and policies

### Node.js
- Provisions compute instances (EC2, Compute Engine, Virtual Machine)
- Installs Node.js and npm
- Deploys a simple Express.js application
- Configures security groups and networking

### Flask
- Provisions compute instances with Python environment
- Installs Flask and dependencies
- Deploys a simple Flask application
- Configures appropriate networking and security

### Docker
- Uses container services (ECS, Cloud Run, Container Instances)
- Deploys containerized applications
- Configures load balancing and networking
- Sets up logging and monitoring

## Optional Services

### Database
- **AWS**: RDS MySQL instance
- **GCP**: Cloud SQL instance
- **Azure**: Azure Database for MySQL
- Includes automated backups and security configurations

### Storage
- **AWS**: Additional S3 bucket with versioning and encryption
- **GCP**: Cloud Storage bucket
- **Azure**: Blob Storage container
- Configured with appropriate access controls

### Load Balancer
- **AWS**: Application Load Balancer (ALB)
- **GCP**: HTTP(S) Load Balancer
- **Azure**: Application Gateway
- Includes health checks and SSL termination

## Configuration File

The `.clouddeploy.json` file stores your deployment configuration:

```json
{
  "project_name": "my-app",
  "app_type": "nodejs",
  "cloud_provider": "aws",
  "region": "us-east-1",
  "services": {
    "database": true,
    "storage": false,
    "load_balancer": true
  },
  "created_at": "2024-01-15T10:30:00Z",
  "last_deployment": "2024-01-15T11:45:00Z"
}
```

## Project Structure

```
go-cli/
├── cmd/                    # Cobra commands
│   ├── root.go            # Root command and global configuration
│   ├── init.go            # Initialize command
│   ├── deploy.go          # Deploy command
│   └── destroy.go         # Destroy command
├── internal/              # Internal packages
│   ├── config/            # Configuration management
│   ├── logger/            # Structured logging
│   ├── providers/         # Cloud provider authentication
│   ├── terraform/         # Terraform operations and templates
│   │   ├── terraform.go   # Core Terraform operations
│   │   ├── aws.go         # AWS-specific templates
│   │   ├── gcp.go         # GCP-specific templates
│   │   └── azure.go       # Azure-specific templates
│   └── utils/             # Shared utilities and prompts
├── main.go                # Application entry point
├── go.mod                 # Go module definition
└── README.md              # This file
```

## Authentication

CloudDeploy checks for authentication with your selected cloud provider:

### AWS
- Checks `aws sts get-caller-identity`
- Requires `aws configure` or `aws sso login`

### Google Cloud
- Checks `gcloud auth list`
- Requires `gcloud auth login`

### Azure
- Checks `az account show`
- Requires `az login`

## Generated Files

When you run `clouddeploy deploy`, the following files are generated in `.terraform-generated/`:

- `main.tf` - Main Terraform configuration
- `variables.tf` - Variable definitions
- `outputs.tf` - Output definitions
- `terraform.tfvars` - Variable values
- `.terraform/` - Terraform state and modules (after `terraform init`)

## Global Flags

- `--config string`: Config file (default is `$HOME/.clouddeploy.yaml`)
- `-v, --verbose`: Enable verbose output for debugging
- `-h, --help`: Show help information

## Examples

### Deploy a Node.js app to AWS with database and load balancer

```bash
# Initialize project
./clouddeploy init
# Select: nodejs, aws, us-east-1, yes to database, no to storage, yes to load balancer

# Deploy
./clouddeploy deploy

# Check outputs for application URL and database credentials
```

### Deploy a static site to GCP

```bash
# Initialize project
./clouddeploy init
# Select: static-site, gcp, us-central1, no to all optional services

# Deploy
./clouddeploy deploy --auto-approve
```

## Troubleshooting

### Authentication Issues
- Ensure you're logged into the correct cloud provider CLI
- Check that your credentials have sufficient permissions
- Verify the CLI tools are installed and in your PATH

### Terraform Errors
- Check that Terraform is installed and accessible
- Ensure you have the required permissions for resource creation
- Review the generated Terraform files in `.terraform-generated/`

### Build Issues
- Ensure you have Go 1.21+ installed
- Run `go mod tidy` to resolve dependencies
- Check that all required packages are available

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues and questions:
- Check the troubleshooting section above
- Review the command help: `./clouddeploy [command] --help`
- Enable verbose logging: `./clouddeploy -v [command]` 