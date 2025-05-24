# S3 Storage Configuration
{{if .EnableStorage}}

# S3 bucket for application storage
resource "aws_s3_bucket" "storage" {
  bucket = "{{.ProjectName}}-storage-${random_id.suffix.hex}"
}

# Bucket versioning
resource "aws_s3_bucket_versioning" "storage" {
  bucket = aws_s3_bucket.storage.id
  versioning_configuration {
    status = "Enabled"
  }
}

# Bucket encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "storage" {
  bucket = aws_s3_bucket.storage.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
    bucket_key_enabled = true
  }
}

# Block public access
resource "aws_s3_bucket_public_access_block" "storage" {
  bucket = aws_s3_bucket.storage.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Bucket lifecycle configuration
resource "aws_s3_bucket_lifecycle_configuration" "storage" {
  bucket = aws_s3_bucket.storage.id

  rule {
    id     = "transition_to_ia"
    status = "Enabled"

    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }

    transition {
      days          = 90
      storage_class = "GLACIER"
    }

    expiration {
      days = 365
    }
  }

  rule {
    id     = "delete_multipart_uploads"
    status = "Enabled"

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}

# IAM policy for application access to storage bucket
resource "aws_iam_policy" "s3_access" {
  name        = "{{.ProjectName}}-s3-access"
  description = "Policy for {{.ProjectName}} to access S3 storage bucket"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.storage.arn,
          "${aws_s3_bucket.storage.arn}/*"
        ]
      }
    ]
  })
}

{{if eq .AppType "nodejs" "flask"}}
# IAM role for EC2 instance to access S3
resource "aws_iam_role" "ec2_s3_role" {
  name = "{{.ProjectName}}-ec2-s3-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ec2_s3_access" {
  role       = aws_iam_role.ec2_s3_role.name
  policy_arn = aws_iam_policy.s3_access.arn
}

resource "aws_iam_instance_profile" "ec2_profile" {
  name = "{{.ProjectName}}-ec2-profile"
  role = aws_iam_role.ec2_s3_role.name
}

{{else if eq .AppType "docker"}}
# IAM role for ECS task to access S3
resource "aws_iam_role" "ecs_task_role" {
  name = "{{.ProjectName}}-ecs-task-role"

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
}

resource "aws_iam_role_policy_attachment" "ecs_s3_access" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = aws_iam_policy.s3_access.arn
}
{{end}}

# Store S3 bucket name in Parameter Store
resource "aws_ssm_parameter" "storage_bucket_name" {
  name  = "/{{.ProjectName}}/storage/bucket_name"
  type  = "String"
  value = aws_s3_bucket.storage.bucket

  tags = {
    Name = "{{.ProjectName}}-storage-bucket-name"
  }
}

resource "aws_ssm_parameter" "storage_bucket_arn" {
  name  = "/{{.ProjectName}}/storage/bucket_arn"
  type  = "String"
  value = aws_s3_bucket.storage.arn

  tags = {
    Name = "{{.ProjectName}}-storage-bucket-arn"
  }
}

{{end}} 