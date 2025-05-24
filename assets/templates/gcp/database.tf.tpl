# Cloud SQL Database Configuration
{{if .EnableDatabase}}

# Random password for database
resource "random_password" "db_password" {
  length  = 16
  special = true
}

# Cloud SQL instance
resource "google_sql_database_instance" "main" {
  name             = "{{.ProjectName}}-db-${random_id.suffix.hex}"
  database_version = "MYSQL_8_0"
  region           = "{{.Region}}"

  settings {
    tier = "db-f1-micro"

    ip_configuration {
      ipv4_enabled                                  = false
      private_network                              = google_compute_network.main.id
      enable_private_path_for_google_cloud_services = true
    }

    backup_configuration {
      enabled                        = true
      start_time                     = "07:00"
      location                       = "{{.Region}}"
      binary_log_enabled            = true
      transaction_log_retention_days = 7
    }

    maintenance_window {
      day  = 7
      hour = 7
    }

    database_flags {
      name  = "slow_query_log"
      value = "on"
    }

    user_labels = local.common_labels
  }

  deletion_protection = false
}

# Database
resource "google_sql_database" "main" {
  name     = "{{.ProjectName | replace "-" "_"}}"
  instance = google_sql_database_instance.main.name
}

# Database user
resource "google_sql_user" "main" {
  name     = "admin"
  instance = google_sql_database_instance.main.name
  password = random_password.db_password.result
}

# Private IP range for VPC peering
resource "google_compute_global_address" "private_ip_range" {
  name          = "{{.ProjectName}}-private-ip-range"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = 16
  network       = google_compute_network.main.id
}

# VPC peering connection
resource "google_service_networking_connection" "private_vpc_connection" {
  network                 = google_compute_network.main.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_ip_range.name]
}

# Secret Manager secrets for database credentials
resource "google_secret_manager_secret" "db_password" {
  secret_id = "{{.ProjectName}}-db-password"

  labels = local.common_labels

  replication {
    automatic = true
  }
}

resource "google_secret_manager_secret_version" "db_password" {
  secret      = google_secret_manager_secret.db_password.id
  secret_data = random_password.db_password.result
}

resource "google_secret_manager_secret" "db_connection_string" {
  secret_id = "{{.ProjectName}}-db-connection"

  labels = local.common_labels

  replication {
    automatic = true
  }
}

resource "google_secret_manager_secret_version" "db_connection_string" {
  secret = google_secret_manager_secret.db_connection_string.id
  secret_data = "mysql://${google_sql_user.main.name}:${random_password.db_password.result}@${google_sql_database_instance.main.public_ip_address}:3306/${google_sql_database.main.name}"
}

# IAM permissions for accessing secrets
{{if eq .AppType "nodejs" "flask"}}
resource "google_secret_manager_secret_iam_member" "db_password_access" {
  secret_id = google_secret_manager_secret.db_password.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.app.email}"
}

resource "google_secret_manager_secret_iam_member" "db_connection_access" {
  secret_id = google_secret_manager_secret.db_connection_string.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.app.email}"
}
{{end}}

{{end}} 