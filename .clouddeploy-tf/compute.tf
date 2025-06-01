# AWS Compute Resources


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
  key_name   = "max-key"
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
    app_type     = "python"
    project_name = "max"
  }))

  root_block_device {
    volume_type = "gp3"
    volume_size = 20
    encrypted   = true
  }

  tags = {
    Name = "max-app-server"
  }
}

# User data script for application setup
resource "local_file" "user_data" {
  content = templatefile("${path.module}/scripts/setup_python.sh", {
    project_name = "max"
  })
  filename = "${path.module}/user_data.sh"
}



 