{{if eq .AppType "static-site"}}
# Cloud Storage bucket for static site hosting
resource "google_storage_bucket" "static_site" {
  name     = "{{.ProjectName}}-static-site-${random_id.suffix.hex}"
  location = "{{.Region}}"

  website {
    main_page_suffix = "index.html"
    not_found_page   = "404.html"
  }

  uniform_bucket_level_access = true

  labels = local.common_labels
}

# Make bucket publicly readable
resource "google_storage_bucket_iam_member" "public_read" {
  bucket = google_storage_bucket.static_site.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

# Upload sample index.html
resource "google_storage_bucket_object" "index" {
  name         = "index.html"
  bucket       = google_storage_bucket.static_site.name
  content_type = "text/html"
  content = <<-EOF
<!DOCTYPE html>
<html>
<head>
    <title>{{.ProjectName}}</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; margin-top: 50px; }
        .container { max-width: 600px; margin: 0 auto; }
        .success { color: #4285f4; }
    </style>
</head>
<body>
    <div class="container">
        <h1 class="success">🎉 {{.ProjectName}} Deployed Successfully!</h1>
        <p>Your static site is now live on Google Cloud Storage.</p>
        <p>Environment: {{.Environment}}</p>
        <p>Region: {{.Region}}</p>
    </div>
</body>
</html>
EOF
}

# Upload sample 404.html
resource "google_storage_bucket_object" "not_found" {
  name         = "404.html"
  bucket       = google_storage_bucket.static_site.name
  content_type = "text/html"
  content = <<-EOF
<!DOCTYPE html>
<html>
<head>
    <title>404 - Page Not Found</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; margin-top: 50px; }
    </style>
</head>
<body>
    <h1>404 - Page Not Found</h1>
    <p>The page you're looking for doesn't exist.</p>
    <a href="/">Go back to home</a>
</body>
</html>
EOF
}

{{else if eq .AppType "docker"}}
# Cloud Run service for Docker applications
resource "google_cloud_run_service" "app" {
  name     = "{{.ProjectName}}-service"
  location = "{{.Region}}"

  template {
    spec {
      containers {
        image = "gcr.io/cloudrun/hello"
        
        ports {
          container_port = 8080
        }

        env {
          name  = "PROJECT_NAME"
          value = "{{.ProjectName}}"
        }

        env {
          name  = "ENVIRONMENT"
          value = "{{.Environment}}"
        }

        resources {
          limits = {
            cpu    = "1000m"
            memory = "512Mi"
          }
        }
      }

      container_concurrency = 100
      timeout_seconds      = 300
    }

    metadata {
      labels = local.common_labels
      annotations = {
        "autoscaling.knative.dev/maxScale" = "10"
        "run.googleapis.com/client-name"   = "terraform"
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  lifecycle {
    ignore_changes = [
      template[0].metadata[0].annotations,
    ]
  }
}

# Make Cloud Run service publicly accessible
resource "google_cloud_run_service_iam_member" "public_access" {
  service  = google_cloud_run_service.app.name
  location = google_cloud_run_service.app.location
  role     = "roles/run.invoker"
  member   = "allUsers"
}

{{else}}
# Compute Engine instance for Node.js and Flask applications
resource "google_compute_instance" "app" {
  name         = "{{.ProjectName}}-instance"
  machine_type = "e2-micro"
  zone         = "{{.Region}}-a"

  tags = ["http-server", "https-server", "ssh-server"]

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2004-lts"
      size  = 10
      type  = "pd-standard"
    }
  }

  network_interface {
    network    = google_compute_network.main.name
    subnetwork = google_compute_subnetwork.main.name
    access_config {
      // Ephemeral public IP
    }
  }

  service_account {
    email  = google_service_account.app.email
    scopes = ["cloud-platform"]
  }

  metadata = {
    ssh-keys = "ubuntu:${tls_private_key.ssh.public_key_openssh}"
  }

  metadata_startup_script = templatefile("${path.module}/startup_{{.AppType}}.sh", {
    project_name = "{{.ProjectName}}"
    environment  = "{{.Environment}}"
    region       = "{{.Region}}"
  })

  labels = local.common_labels
}

# Service account for the instance
resource "google_service_account" "app" {
  account_id   = "{{.ProjectName}}-sa"
  display_name = "Service account for {{.ProjectName}}"
  description  = "Service account for {{.ProjectName}} application"
}

# SSH key pair
resource "tls_private_key" "ssh" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "local_file" "private_key" {
  content  = tls_private_key.ssh.private_key_pem
  filename = "${path.module}/{{.ProjectName}}-key.pem"

  provisioner "local-exec" {
    command = "chmod 600 ${path.module}/{{.ProjectName}}-key.pem"
  }
}

# Startup script for Node.js
resource "local_file" "startup_nodejs" {
  count    = "{{.AppType}}" == "nodejs" ? 1 : 0
  filename = "${path.module}/startup_nodejs.sh"
  content = <<-EOF
#!/bin/bash
set -e

# Update system
apt-get update
apt-get install -y curl software-properties-common

# Install Node.js
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
apt-get install -y nodejs

# Create application directory
mkdir -p /opt/{{.ProjectName}}
cd /opt/{{.ProjectName}}

# Create package.json
cat > package.json << 'JSON'
{
  "name": "{{.ProjectName}}",
  "version": "1.0.0",
  "description": "Node.js application deployed with CloudDeploy",
  "main": "server.js",
  "scripts": {
    "start": "node server.js"
  },
  "dependencies": {
    "express": "^4.18.0"
  }
}
JSON

# Create server.js
cat > server.js << 'JS'
const express = require('express');
const app = express();
const port = 80;

app.get('/', (req, res) => {
  res.send(`
    <html>
      <head><title>{{.ProjectName}}</title></head>
      <body style="font-family: Arial; text-align: center; margin-top: 50px;">
        <h1>🚀 {{.ProjectName}} is Running!</h1>
        <p>Node.js application deployed successfully on Google Cloud</p>
        <p>Environment: {{.Environment}}</p>
        <p>Region: {{.Region}}</p>
      </body>
    </html>
  `);
});

app.listen(port, () => {
  console.log('{{.ProjectName}} listening on port', port);
});
JS

# Install dependencies and start the app
npm install

# Create systemd service
cat > /etc/systemd/system/{{.ProjectName}}.service << 'SERVICE'
[Unit]
Description={{.ProjectName}} Node.js App
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/{{.ProjectName}}
ExecStart=/usr/bin/node server.js
Restart=on-failure

[Install]
WantedBy=multi-user.target
SERVICE

# Enable and start the service
systemctl enable {{.ProjectName}}
systemctl start {{.ProjectName}}
EOF
}

# Startup script for Flask
resource "local_file" "startup_flask" {
  count    = "{{.AppType}}" == "flask" ? 1 : 0
  filename = "${path.module}/startup_flask.sh"
  content = <<-EOF
#!/bin/bash
set -e

# Update system
apt-get update
apt-get install -y python3 python3-pip

# Create application directory
mkdir -p /opt/{{.ProjectName}}
cd /opt/{{.ProjectName}}

# Create Flask application
cat > app.py << 'PYTHON'
from flask import Flask

app = Flask(__name__)

@app.route('/')
def hello():
    return '''
    <html>
      <head><title>{{.ProjectName}}</title></head>
      <body style="font-family: Arial; text-align: center; margin-top: 50px;">
        <h1>🐍 {{.ProjectName}} is Running!</h1>
        <p>Flask application deployed successfully on Google Cloud</p>
        <p>Environment: {{.Environment}}</p>
        <p>Region: {{.Region}}</p>
      </body>
    </html>
    '''

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=80, debug=False)
PYTHON

# Install Flask
pip3 install flask

# Create systemd service
cat > /etc/systemd/system/{{.ProjectName}}.service << 'SERVICE'
[Unit]
Description={{.ProjectName}} Flask App
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/{{.ProjectName}}
ExecStart=/usr/bin/python3 app.py
Restart=on-failure

[Install]
WantedBy=multi-user.target
SERVICE

# Enable and start the service
systemctl enable {{.ProjectName}}
systemctl start {{.ProjectName}}
EOF
}

{{end}} 