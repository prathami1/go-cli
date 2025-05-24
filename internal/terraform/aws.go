package terraform

import (
	"github.com/prathami1/go-cli/internal/config"
)

// generateAWSMainTF generates the main Terraform configuration for AWS
func generateAWSMainTF(cfg *config.DeploymentConfig) string {
	content := `# Terraform configuration for AWS deployment
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.region
}

# Data sources
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_caller_identity" "current" {}

# VPC
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "${var.project_name}-vpc"
    Project     = var.project_name
    Environment = "production"
  }
}

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name    = "${var.project_name}-igw"
    Project = var.project_name
  }
}

# Public Subnets
resource "aws_subnet" "public" {
  count = 2

  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.${count.index + 1}.0/24"
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name    = "${var.project_name}-public-subnet-${count.index + 1}"
    Project = var.project_name
    Type    = "Public"
  }
}

# Route Table
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name    = "${var.project_name}-public-rt"
    Project = var.project_name
  }
}

# Route Table Association
resource "aws_route_table_association" "public" {
  count = length(aws_subnet.public)

  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# Security Group
resource "aws_security_group" "app" {
  name_prefix = "${var.project_name}-sg"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name    = "${var.project_name}-sg"
    Project = var.project_name
  }
}
`

	// Add app-specific resources based on app type
	switch cfg.AppType {
	case config.StaticSite:
		content += generateAWSStaticSiteResources()
	case config.NodeJS, config.Flask:
		content += generateAWSComputeResources()
	case config.Docker:
		content += generateAWSContainerResources()
	}

	// Add optional services
	if cfg.Services.Database {
		content += generateAWSDatabaseResources()
	}

	if cfg.Services.Storage {
		content += generateAWSStorageResources()
	}

	if cfg.Services.LoadBalancer {
		content += generateAWSLoadBalancerResources()
	}

	return content
}

// generateAWSOutputsTF generates outputs for AWS deployment
func generateAWSOutputsTF(cfg *config.DeploymentConfig) string {
	content := `# Outputs for AWS deployment

output "vpc_id" {
  description = "ID of the VPC"
  value       = aws_vpc.main.id
}

output "region" {
  description = "AWS region"
  value       = var.region
}

output "project_name" {
  description = "Project name"
  value       = var.project_name
}
`

	// Add app-specific outputs
	switch cfg.AppType {
	case config.StaticSite:
		content += `
output "website_url" {
  description = "Website URL"
  value       = "https://${aws_s3_bucket_website_configuration.main.website_endpoint}"
}

output "s3_bucket_name" {
  description = "S3 bucket name for static site"
  value       = aws_s3_bucket.static_site.bucket
}
`
	case config.NodeJS, config.Flask:
		content += `
output "instance_public_ip" {
  description = "Public IP address of the EC2 instance"
  value       = aws_instance.app.public_ip
}

output "instance_public_dns" {
  description = "Public DNS name of the EC2 instance"
  value       = aws_instance.app.public_dns
}

output "application_url" {
  description = "Application URL"
  value       = "http://${aws_instance.app.public_ip}"
}
`
	case config.Docker:
		content += `
output "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  value       = aws_ecs_cluster.main.name
}

output "load_balancer_dns" {
  description = "Load balancer DNS name"
  value       = var.enable_load_balancer ? aws_lb.main[0].dns_name : null
}
`
	}

	// Add optional service outputs
	if cfg.Services.Database {
		content += `
output "database_endpoint" {
  description = "Database endpoint"
  value       = aws_db_instance.main[0].endpoint
  sensitive   = true
}

output "database_name" {
  description = "Database name"
  value       = aws_db_instance.main[0].db_name
}
`
	}

	if cfg.Services.Storage {
		content += `
output "storage_bucket_name" {
  description = "Storage bucket name"
  value       = aws_s3_bucket.storage[0].bucket
}
`
	}

	return content
}

func generateAWSStaticSiteResources() string {
	return `
# S3 Bucket for static site
resource "aws_s3_bucket" "static_site" {
  bucket = "${var.project_name}-static-site-${random_id.bucket_suffix.hex}"

  tags = {
    Name    = "${var.project_name}-static-site"
    Project = var.project_name
  }
}

resource "random_id" "bucket_suffix" {
  byte_length = 4
}

resource "aws_s3_bucket_website_configuration" "main" {
  bucket = aws_s3_bucket.static_site.id

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "error.html"
  }
}

resource "aws_s3_bucket_public_access_block" "main" {
  bucket = aws_s3_bucket.static_site.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_policy" "main" {
  bucket = aws_s3_bucket.static_site.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "PublicReadGetObject"
        Effect    = "Allow"
        Principal = "*"
        Action    = "s3:GetObject"
        Resource  = "${aws_s3_bucket.static_site.arn}/*"
      },
    ]
  })

  depends_on = [aws_s3_bucket_public_access_block.main]
}
`
}

func generateAWSComputeResources() string {
	return `
# Get latest Amazon Linux 2 AMI
data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }
}

# EC2 Instance
resource "aws_instance" "app" {
  ami                    = data.aws_ami.amazon_linux.id
  instance_type          = "t3.micro"
  subnet_id              = aws_subnet.public[0].id
  vpc_security_group_ids = [aws_security_group.app.id]

  user_data = base64encode(templatefile("${path.module}/user_data.sh", {
    app_type = var.app_type
  }))

  tags = {
    Name    = "${var.project_name}-app"
    Project = var.project_name
  }
}

# User data script
resource "local_file" "user_data" {
  filename = "${path.module}/user_data.sh"
  content = var.app_type == "nodejs" ? local.nodejs_user_data : local.flask_user_data
}

locals {
  nodejs_user_data = <<-EOF
#!/bin/bash
yum update -y
yum install -y nodejs npm git
mkdir -p /opt/app
cd /opt/app
echo '{"name": "${var.project_name}", "version": "1.0.0", "scripts": {"start": "node server.js"}, "dependencies": {"express": "^4.18.0"}}' > package.json
echo 'const express = require("express"); const app = express(); app.get("/", (req, res) => res.send("Hello from ${var.project_name} on AWS!")); app.listen(80, () => console.log("Server running on port 80"));' > server.js
npm install
node server.js
EOF

  flask_user_data = <<-EOF
#!/bin/bash
yum update -y
yum install -y python3 python3-pip git
mkdir -p /opt/app
cd /opt/app
echo 'from flask import Flask; app = Flask(__name__); @app.route("/"); def hello(): return "Hello from ${var.project_name} on AWS!"; if __name__ == "__main__": app.run(host="0.0.0.0", port=80)' > app.py
pip3 install flask
python3 app.py
EOF
}
`
}

func generateAWSContainerResources() string {
	return `
# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = "${var.project_name}-cluster"

  tags = {
    Name    = "${var.project_name}-cluster"
    Project = var.project_name
  }
}

# ECS Task Definition
resource "aws_ecs_task_definition" "app" {
  family                   = "${var.project_name}-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = aws_iam_role.ecs_execution.arn

  container_definitions = jsonencode([
    {
      name  = var.project_name
      image = "nginx:alpine"
      portMappings = [
        {
          containerPort = 80
          protocol      = "tcp"
        }
      ]
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.app.name
          "awslogs-region"        = var.region
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])

  tags = {
    Name    = "${var.project_name}-task"
    Project = var.project_name
  }
}

# ECS Service
resource "aws_ecs_service" "app" {
  name            = "${var.project_name}-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.public[*].id
    security_groups  = [aws_security_group.app.id]
    assign_public_ip = true
  }

  tags = {
    Name    = "${var.project_name}-service"
    Project = var.project_name
  }
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "app" {
  name              = "/ecs/${var.project_name}"
  retention_in_days = 7

  tags = {
    Name    = "${var.project_name}-logs"
    Project = var.project_name
  }
}

# ECS Execution Role
resource "aws_iam_role" "ecs_execution" {
  name = "${var.project_name}-ecs-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name    = "${var.project_name}-ecs-execution-role"
    Project = var.project_name
  }
}

resource "aws_iam_role_policy_attachment" "ecs_execution" {
  role       = aws_iam_role.ecs_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}
`
}

func generateAWSDatabaseResources() string {
	return `
# Database Subnet Group
resource "aws_db_subnet_group" "main" {
  count = var.enable_database ? 1 : 0

  name       = "${var.project_name}-db-subnet-group"
  subnet_ids = aws_subnet.public[*].id

  tags = {
    Name    = "${var.project_name}-db-subnet-group"
    Project = var.project_name
  }
}

# Database Security Group
resource "aws_security_group" "database" {
  count = var.enable_database ? 1 : 0

  name_prefix = "${var.project_name}-db-sg"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 3306
    to_port         = 3306
    protocol        = "tcp"
    security_groups = [aws_security_group.app.id]
  }

  tags = {
    Name    = "${var.project_name}-db-sg"
    Project = var.project_name
  }
}

# RDS Instance
resource "aws_db_instance" "main" {
  count = var.enable_database ? 1 : 0

  identifier     = "${var.project_name}-db"
  engine         = "mysql"
  engine_version = "8.0"
  instance_class = "db.t3.micro"
  
  allocated_storage     = 20
  max_allocated_storage = 100
  storage_type          = "gp2"
  storage_encrypted     = true

  db_name  = replace(var.project_name, "-", "_")
  username = "admin"
  password = random_password.db_password[0].result

  vpc_security_group_ids = [aws_security_group.database[0].id]
  db_subnet_group_name   = aws_db_subnet_group.main[0].name

  backup_retention_period = 7
  backup_window          = "07:00-09:00"
  maintenance_window     = "sun:09:00-sun:10:00"

  skip_final_snapshot = true
  deletion_protection = false

  tags = {
    Name    = "${var.project_name}-db"
    Project = var.project_name
  }
}

resource "random_password" "db_password" {
  count = var.enable_database ? 1 : 0

  length  = 16
  special = true
}
`
}

func generateAWSStorageResources() string {
	return `
# Additional S3 Bucket for storage
resource "aws_s3_bucket" "storage" {
  count = var.enable_storage ? 1 : 0

  bucket = "${var.project_name}-storage-${random_id.storage_suffix[0].hex}"

  tags = {
    Name    = "${var.project_name}-storage"
    Project = var.project_name
  }
}

resource "random_id" "storage_suffix" {
  count = var.enable_storage ? 1 : 0

  byte_length = 4
}

resource "aws_s3_bucket_versioning" "storage" {
  count = var.enable_storage ? 1 : 0

  bucket = aws_s3_bucket.storage[0].id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "storage" {
  count = var.enable_storage ? 1 : 0

  bucket = aws_s3_bucket.storage[0].id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}
`
}

func generateAWSLoadBalancerResources() string {
	return `
# Application Load Balancer
resource "aws_lb" "main" {
  count = var.enable_load_balancer ? 1 : 0

  name               = "${var.project_name}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.app.id]
  subnets            = aws_subnet.public[*].id

  enable_deletion_protection = false

  tags = {
    Name    = "${var.project_name}-alb"
    Project = var.project_name
  }
}

# Target Group
resource "aws_lb_target_group" "main" {
  count = var.enable_load_balancer ? 1 : 0

  name     = "${var.project_name}-tg"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.main.id

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 5
    interval            = 30
    path                = "/"
    matcher             = "200"
  }

  tags = {
    Name    = "${var.project_name}-tg"
    Project = var.project_name
  }
}

# ALB Listener
resource "aws_lb_listener" "main" {
  count = var.enable_load_balancer ? 1 : 0

  load_balancer_arn = aws_lb.main[0].arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.main[0].arn
  }
}
`
}
