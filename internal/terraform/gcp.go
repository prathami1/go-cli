package terraform

import (
	"github.com/prathami1/go-cli/internal/config"
)

// generateGCPMainTF generates the main Terraform configuration for GCP
func generateGCPMainTF(cfg *config.DeploymentConfig) string {
	return `# Terraform configuration for GCP deployment
terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  region = var.region
}

# VPC Network
resource "google_compute_network" "main" {
  name                    = "${var.project_name}-vpc"
  auto_create_subnetworks = false
}

# Subnet
resource "google_compute_subnetwork" "main" {
  name          = "${var.project_name}-subnet"
  ip_cidr_range = "10.0.1.0/24"
  region        = var.region
  network       = google_compute_network.main.id
}

# Firewall rule
resource "google_compute_firewall" "allow_http" {
  name    = "${var.project_name}-allow-http"
  network = google_compute_network.main.name

  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["http-server"]
}

# Compute Engine instance for apps
resource "google_compute_instance" "app" {
  name         = "${var.project_name}-instance"
  machine_type = "e2-micro"
  zone         = "${var.region}-a"

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2004-lts"
    }
  }

  network_interface {
    network    = google_compute_network.main.name
    subnetwork = google_compute_subnetwork.main.name
    access_config {
      // Ephemeral IP
    }
  }

  metadata_startup_script = "echo 'Hello from ${var.project_name} on GCP!' > /var/www/html/index.html"

  tags = ["http-server"]
}
`
}

// generateGCPOutputsTF generates outputs for GCP deployment
func generateGCPOutputsTF(cfg *config.DeploymentConfig) string {
	return `# Outputs for GCP deployment

output "vpc_network_name" {
  description = "Name of the VPC network"
  value       = google_compute_network.main.name
}

output "region" {
  description = "GCP region"
  value       = var.region
}

output "project_name" {
  description = "Project name"
  value       = var.project_name
}

output "instance_external_ip" {
  description = "External IP address of the instance"
  value       = google_compute_instance.app.network_interface[0].access_config[0].nat_ip
}

output "application_url" {
  description = "Application URL"
  value       = "http://${google_compute_instance.app.network_interface[0].access_config[0].nat_ip}"
}
`
}
