# AWS Compute Resources

{{if eq .AppType "static-site"}}
# S3 bucket for static site hosting
resource "aws_s3_bucket" "static_site" {
  bucket = "{{.ProjectName}}-static-site-${random_id.suffix.hex}"
}

resource "aws_s3_bucket_website_configuration" "static_site" {
  bucket = aws_s3_bucket.static_site.id

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "error.html"
  }
}

resource "aws_s3_bucket_public_access_block" "static_site" {
  bucket = aws_s3_bucket.static_site.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_policy" "static_site" {
  bucket = aws_s3_bucket.static_site.id
  depends_on = [aws_s3_bucket_public_access_block.static_site]

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
}

{{else if eq .AppType "docker"}}
# ECS Cluster for Docker applications
resource "aws_ecs_cluster" "main" {
  name = "{{.ProjectName}}-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

# ECS Task Definition
resource "aws_ecs_task_definition" "app" {
  family                   = "{{.ProjectName}}-app"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_execution_role.arn

  container_definitions = jsonencode([
    {
      name  = "{{.ProjectName}}-container"
      image = "nginx:latest"  # Default image, should be replaced with actual app image
      portMappings = [
        {
          containerPort = 80
          hostPort      = 80
          protocol      = "tcp"
        }
      ]
      essential = true
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.app.name
          awslogs-region        = "{{.Region}}"
          awslogs-stream-prefix = "ecs"
        }
      }
    }
  ])
}

# ECS Service
resource "aws_ecs_service" "app" {
  name            = "{{.ProjectName}}-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.public[*].id
    security_groups  = [aws_security_group.app.id]
    assign_public_ip = true
  }

  {{if .EnableLoadBalancer}}
  load_balancer {
    target_group_arn = aws_lb_target_group.app[0].arn
    container_name   = "{{.ProjectName}}-container"
    container_port   = 80
  }

  depends_on = [aws_lb_listener.app]
  {{end}}
}

# IAM Role for ECS Execution
resource "aws_iam_role" "ecs_execution_role" {
  name = "{{.ProjectName}}-ecs-execution-role"

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

resource "aws_iam_role_policy_attachment" "ecs_execution_role_policy" {
  role       = aws_iam_role.ecs_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "app" {
  name              = "/ecs/{{.ProjectName}}"
  retention_in_days = 7
}

{{else}}
# EC2 Instance for Node.js/Flask applications
data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# Key Pair for SSH access
resource "aws_key_pair" "app" {
  key_name   = "{{.ProjectName}}-key"
  public_key = tls_private_key.app.public_key_openssh
}

resource "tls_private_key" "app" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

# EC2 Instance
resource "aws_instance" "app" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = "t3.micro"
  key_name               = aws_key_pair.app.key_name
  vpc_security_group_ids = [aws_security_group.app.id]
  subnet_id              = aws_subnet.public[0].id

  user_data = base64encode(templatefile("${path.module}/user_data.sh", {
    app_type     = "{{.AppType}}"
    project_name = "{{.ProjectName}}"
  }))

  root_block_device {
    volume_type = "gp3"
    volume_size = 20
    encrypted   = true
  }

  tags = {
    Name = "{{.ProjectName}}-app-server"
  }
}

# User data script for application setup
resource "local_file" "user_data" {
  content = templatefile("${path.module}/scripts/setup_{{.AppType}}.sh", {
    project_name = "{{.ProjectName}}"
  })
  filename = "${path.module}/user_data.sh"
}

{{if eq .AppType "nodejs"}}
# User data script for Node.js
resource "local_file" "setup_nodejs" {
  content = <<-EOF
#!/bin/bash
set -e

# Update system
apt-get update
apt-get upgrade -y

# Install Node.js
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
apt-get install -y nodejs

# Install PM2 for process management
npm install -g pm2

# Create app directory
mkdir -p /opt/{{.ProjectName}}
cd /opt/{{.ProjectName}}

# Create a simple Express.js app
cat > package.json << 'EOL'
{
  "name": "{{.ProjectName}}",
  "version": "1.0.0",
  "description": "CloudDeploy generated Node.js app",
  "main": "app.js",
  "scripts": {
    "start": "node app.js"
  },
  "dependencies": {
    "express": "^4.18.0"
  }
}
EOL

cat > app.js << 'EOL'
const express = require('express');
const app = express();
const port = 80;

app.get('/', (req, res) => {
  res.send(`
    <h1>Welcome to {{.ProjectName}}</h1>
    <p>Your Node.js application is running successfully!</p>
    <p>Deployed with CloudDeploy CLI</p>
  `);
});

app.listen(port, () => {
  console.log(`{{.ProjectName}} listening at http://localhost:${port}`);
});
EOL

# Install dependencies
npm install

# Start the application with PM2
pm2 start app.js --name "{{.ProjectName}}"
pm2 startup
pm2 save

# Configure nginx as reverse proxy
apt-get install -y nginx
cat > /etc/nginx/sites-available/{{.ProjectName}} << 'EOL'
server {
    listen 80;
    server_name _;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
EOL

ln -s /etc/nginx/sites-available/{{.ProjectName}} /etc/nginx/sites-enabled/
rm /etc/nginx/sites-enabled/default
systemctl restart nginx
systemctl enable nginx

echo "Node.js application setup complete!"
EOF
  filename = "${path.module}/scripts/setup_nodejs.sh"
}

{{else if eq .AppType "flask"}}
# User data script for Flask
resource "local_file" "setup_flask" {
  content = <<-EOF
#!/bin/bash
set -e

# Update system
apt-get update
apt-get upgrade -y

# Install Python and pip
apt-get install -y python3 python3-pip python3-venv nginx

# Create app directory
mkdir -p /opt/{{.ProjectName}}
cd /opt/{{.ProjectName}}

# Create virtual environment
python3 -m venv venv
source venv/bin/activate

# Create requirements.txt
cat > requirements.txt << 'EOL'
Flask==2.3.3
gunicorn==21.2.0
EOL

# Create Flask app
cat > app.py << 'EOL'
from flask import Flask

app = Flask(__name__)

@app.route('/')
def hello():
    return '''
    <h1>Welcome to {{.ProjectName}}</h1>
    <p>Your Flask application is running successfully!</p>
    <p>Deployed with CloudDeploy CLI</p>
    '''

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
EOL

# Install dependencies
pip install -r requirements.txt

# Create systemd service
cat > /etc/systemd/system/{{.ProjectName}}.service << 'EOL'
[Unit]
Description={{.ProjectName}} Flask App
After=network.target

[Service]
User=root
WorkingDirectory=/opt/{{.ProjectName}}
Environment=PATH=/opt/{{.ProjectName}}/venv/bin
ExecStart=/opt/{{.ProjectName}}/venv/bin/gunicorn --bind 0.0.0.0:5000 app:app
Restart=always

[Install]
WantedBy=multi-user.target
EOL

# Start the service
systemctl daemon-reload
systemctl start {{.ProjectName}}
systemctl enable {{.ProjectName}}

# Configure nginx
cat > /etc/nginx/sites-available/{{.ProjectName}} << 'EOL'
server {
    listen 80;
    server_name _;

    location / {
        proxy_pass http://localhost:5000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOL

ln -s /etc/nginx/sites-available/{{.ProjectName}} /etc/nginx/sites-enabled/
rm /etc/nginx/sites-enabled/default
systemctl restart nginx
systemctl enable nginx

echo "Flask application setup complete!"
EOF
  filename = "${path.module}/scripts/setup_flask.sh"
}
{{end}}

{{end}} 