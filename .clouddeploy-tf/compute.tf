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
  key_name   = "test-key"
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
    app_type     = "flask"
    project_name = "test"
  }))

  root_block_device {
    volume_type = "gp3"
    volume_size = 20
    encrypted   = true
  }

  tags = {
    Name = "test-app-server"
  }
}

# User data script for application setup
resource "local_file" "user_data" {
  content = templatefile("${path.module}/scripts/setup_flask.sh", {
    project_name = "test"
  })
  filename = "${path.module}/user_data.sh"
}


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
mkdir -p /opt/test
cd /opt/test

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
    <h1>Welcome to test</h1>
    <p>Your Flask application is running successfully!</p>
    <p>Deployed with CloudDeploy CLI</p>
    '''

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
EOL

# Install dependencies
pip install -r requirements.txt

# Create systemd service
cat > /etc/systemd/system/test.service << 'EOL'
[Unit]
Description=test Flask App
After=network.target

[Service]
User=root
WorkingDirectory=/opt/test
Environment=PATH=/opt/test/venv/bin
ExecStart=/opt/test/venv/bin/gunicorn --bind 0.0.0.0:5000 app:app
Restart=always

[Install]
WantedBy=multi-user.target
EOL

# Start the service
systemctl daemon-reload
systemctl start test
systemctl enable test

# Configure nginx
cat > /etc/nginx/sites-available/test << 'EOL'
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

ln -s /etc/nginx/sites-available/test /etc/nginx/sites-enabled/
rm /etc/nginx/sites-enabled/default
systemctl restart nginx
systemctl enable nginx

echo "Flask application setup complete!"
EOF
  filename = "${path.module}/scripts/setup_flask.sh"
}


 