{{if eq .AppType "static-site"}}
# Storage Account for static website hosting
resource "azurerm_storage_account" "static_site" {
  name                     = "st{{.ProjectName | replace "-" ""}}${random_id.suffix.hex}"
  resource_group_name      = azurerm_resource_group.main.name
  location                 = azurerm_resource_group.main.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  static_website {
    index_document     = "index.html"
    error_404_document = "404.html"
  }

  tags = azurerm_resource_group.main.tags
}

# Upload sample index.html
resource "azurerm_storage_blob" "index" {
  name                   = "index.html"
  storage_account_name   = azurerm_storage_account.static_site.name
  storage_container_name = "$web"
  type                   = "Block"
  content_type           = "text/html"
  source_content = <<-EOF
<!DOCTYPE html>
<html>
<head>
    <title>{{.ProjectName}}</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; margin-top: 50px; }
        .container { max-width: 600px; margin: 0 auto; }
        .success { color: #0078d4; }
    </style>
</head>
<body>
    <div class="container">
        <h1 class="success">🎉 {{.ProjectName}} Deployed Successfully!</h1>
        <p>Your static site is now live on Azure Storage.</p>
        <p>Environment: {{.Environment}}</p>
        <p>Region: {{.Region}}</p>
    </div>
</body>
</html>
EOF
}

{{else if eq .AppType "docker"}}
# Container Group for Docker applications
resource "azurerm_container_group" "app" {
  name                = "ci-{{.ProjectName}}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  ip_address_type     = "Public"
  dns_name_label      = "{{.ProjectName}}-${random_id.suffix.hex}"
  os_type             = "Linux"

  container {
    name   = "{{.ProjectName}}"
    image  = "nginx:alpine"
    cpu    = "0.5"
    memory = "1.0"

    ports {
      port     = 80
      protocol = "TCP"
    }

    environment_variables = {
      PROJECT_NAME = "{{.ProjectName}}"
      ENVIRONMENT  = "{{.Environment}}"
    }
  }

  tags = azurerm_resource_group.main.tags
}

{{else}}
# Public IP
resource "azurerm_public_ip" "main" {
  name                = "pip-{{.ProjectName}}"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  allocation_method   = "Static"
  sku                 = "Standard"

  tags = azurerm_resource_group.main.tags
}

# Network Interface
resource "azurerm_network_interface" "main" {
  name                = "nic-{{.ProjectName}}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.public.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.main.id
  }

  tags = azurerm_resource_group.main.tags
}

# SSH Key
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

# Virtual Machine
resource "azurerm_linux_virtual_machine" "main" {
  name                = "vm-{{.ProjectName}}"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  size                = "Standard_B1s"
  admin_username      = "azureuser"

  disable_password_authentication = true

  network_interface_ids = [
    azurerm_network_interface.main.id,
  ]

  admin_ssh_key {
    username   = "azureuser"
    public_key = tls_private_key.ssh.public_key_openssh
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-focal"
    sku       = "20_04-lts-gen2"
    version   = "latest"
  }

  custom_data = base64encode(templatefile("${path.module}/custom_data_{{.AppType}}.sh", {
    project_name = "{{.ProjectName}}"
    environment  = "{{.Environment}}"
    region       = "{{.Region}}"
  }))

  tags = azurerm_resource_group.main.tags
}

# Custom data script for Node.js
resource "local_file" "custom_data_nodejs" {
  count    = "{{.AppType}}" == "nodejs" ? 1 : 0
  filename = "${path.module}/custom_data_nodejs.sh"
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
        <p>Node.js application deployed successfully on Azure</p>
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

# Custom data script for Flask
resource "local_file" "custom_data_flask" {
  count    = "{{.AppType}}" == "flask" ? 1 : 0
  filename = "${path.module}/custom_data_flask.sh"
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
        <p>Flask application deployed successfully on Azure</p>
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