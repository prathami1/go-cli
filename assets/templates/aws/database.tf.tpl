# RDS Database Configuration
{{if .EnableDatabase}}

# Private subnets for database
resource "aws_subnet" "private" {
  count = 2

  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.${count.index + 10}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = "{{.ProjectName}}-private-subnet-${count.index + 1}"
    Type = "Private"
  }
}

# Database subnet group
resource "aws_db_subnet_group" "main" {
  name       = "{{.ProjectName}}-db-subnet-group"
  subnet_ids = aws_subnet.private[*].id

  tags = {
    Name = "{{.ProjectName}}-db-subnet-group"
  }
}

# Security group for database
resource "aws_security_group" "database" {
  name_prefix = "{{.ProjectName}}-db-sg"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 3306
    to_port         = 3306
    protocol        = "tcp"
    security_groups = [aws_security_group.app.id]
    description     = "MySQL access from application"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "{{.ProjectName}}-db-sg"
  }
}

# Random password for database
resource "random_password" "db_password" {
  length  = 16
  special = true
}

# RDS instance
resource "aws_db_instance" "main" {
  identifier     = "{{.ProjectName}}-db"
  engine         = "mysql"
  engine_version = "8.0"
  instance_class = "db.t3.micro"
  
  allocated_storage     = 20
  max_allocated_storage = 100
  storage_type          = "gp2"
  storage_encrypted     = true

  db_name  = "{{.ProjectName | replace "-" "_"}}"
  username = "admin"
  password = random_password.db_password.result

  vpc_security_group_ids = [aws_security_group.database.id]
  db_subnet_group_name   = aws_db_subnet_group.main.name

  backup_retention_period = 7
  backup_window          = "07:00-09:00"
  maintenance_window     = "sun:09:00-sun:10:00"

  skip_final_snapshot = true
  deletion_protection = false

  # Performance Insights
  performance_insights_enabled = true
  monitoring_interval         = 60
  monitoring_role_arn        = aws_iam_role.rds_monitoring.arn

  tags = {
    Name = "{{.ProjectName}}-database"
  }
}

# IAM role for RDS monitoring
resource "aws_iam_role" "rds_monitoring" {
  name = "{{.ProjectName}}-rds-monitoring-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "monitoring.rds.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "rds_monitoring" {
  role       = aws_iam_role.rds_monitoring.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
}

# Store database credentials in Systems Manager Parameter Store
resource "aws_ssm_parameter" "db_endpoint" {
  name  = "/{{.ProjectName}}/database/endpoint"
  type  = "String"
  value = aws_db_instance.main.endpoint

  tags = {
    Name = "{{.ProjectName}}-db-endpoint"
  }
}

resource "aws_ssm_parameter" "db_username" {
  name  = "/{{.ProjectName}}/database/username"
  type  = "String"
  value = aws_db_instance.main.username

  tags = {
    Name = "{{.ProjectName}}-db-username"
  }
}

resource "aws_ssm_parameter" "db_password" {
  name  = "/{{.ProjectName}}/database/password"
  type  = "SecureString"
  value = random_password.db_password.result

  tags = {
    Name = "{{.ProjectName}}-db-password"
  }
}

resource "aws_ssm_parameter" "db_name" {
  name  = "/{{.ProjectName}}/database/name"
  type  = "String"
  value = aws_db_instance.main.db_name

  tags = {
    Name = "{{.ProjectName}}-db-name"
  }
}

{{end}} 