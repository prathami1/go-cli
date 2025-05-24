# Google Cloud Platform Provider Configuration
terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.1"
    }
  }
}

provider "google" {
  region = "{{.Region}}"
  zone   = "{{.Region}}-a"
}

# Random ID for unique resource names
resource "random_id" "suffix" {
  byte_length = 4
}

# VPC Network
resource "google_compute_network" "main" {
  name                    = "{{.ProjectName}}-vpc"
  auto_create_subnetworks = false
  description             = "VPC network for {{.ProjectName}}"
}

# Subnet
resource "google_compute_subnetwork" "main" {
  name          = "{{.ProjectName}}-subnet"
  ip_cidr_range = "10.0.1.0/24"
  region        = "{{.Region}}"
  network       = google_compute_network.main.id
  description   = "Main subnet for {{.ProjectName}}"
}

# Firewall rules
resource "google_compute_firewall" "allow_http" {
  name    = "{{.ProjectName}}-allow-http"
  network = google_compute_network.main.name

  allow {
    protocol = "tcp"
    ports    = ["80", "8080"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["http-server"]
  description   = "Allow HTTP traffic"
}

resource "google_compute_firewall" "allow_https" {
  name    = "{{.ProjectName}}-allow-https"
  network = google_compute_network.main.name

  allow {
    protocol = "tcp"
    ports    = ["443"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["https-server"]
  description   = "Allow HTTPS traffic"
}

{{if eq .AppType "nodejs" "flask"}}
resource "google_compute_firewall" "allow_ssh" {
  name    = "{{.ProjectName}}-allow-ssh"
  network = google_compute_network.main.name

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["ssh-server"]
  description   = "Allow SSH access"
}
{{end}}

# Labels for all resources
locals {
  common_labels = {
    project     = "{{.ProjectName}}"
    environment = "{{.Environment}}"
    managed_by  = "clouddeploy"
    app_type    = "{{.AppType}}"
  }
} 