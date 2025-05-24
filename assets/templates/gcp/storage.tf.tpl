# Cloud Storage Configuration
{{if .EnableStorage}}

# Cloud Storage bucket for application storage
resource "google_storage_bucket" "storage" {
  name     = "{{.ProjectName}}-storage-${random_id.suffix.hex}"
  location = "{{.Region}}"

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }

  lifecycle_rule {
    condition {
      age = 30
    }
    action {
      type          = "SetStorageClass"
      storage_class = "NEARLINE"
    }
  }

  lifecycle_rule {
    condition {
      age = 90
    }
    action {
      type          = "SetStorageClass"
      storage_class = "COLDLINE"
    }
  }

  lifecycle_rule {
    condition {
      age = 365
    }
    action {
      type = "Delete"
    }
  }

  labels = local.common_labels
}

# IAM binding for service account access to storage bucket
{{if eq .AppType "nodejs" "flask"}}
resource "google_storage_bucket_iam_member" "storage_access" {
  bucket = google_storage_bucket.storage.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.app.email}"
}
{{else if eq .AppType "docker"}}
# For Cloud Run, we'll need to set up service account access
resource "google_service_account" "storage_sa" {
  account_id   = "{{.ProjectName}}-storage-sa"
  display_name = "Storage service account for {{.ProjectName}}"
  description  = "Service account for {{.ProjectName}} storage access"
}

resource "google_storage_bucket_iam_member" "storage_access" {
  bucket = google_storage_bucket.storage.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.storage_sa.email}"
}

# Bind service account to Cloud Run service
resource "google_cloud_run_service_iam_member" "storage_sa_binding" {
  service  = google_cloud_run_service.app.name
  location = google_cloud_run_service.app.location
  role     = "roles/run.serviceAgent"
  member   = "serviceAccount:${google_service_account.storage_sa.email}"
}
{{end}}

# Store bucket name and info in Secret Manager
resource "google_secret_manager_secret" "storage_bucket_name" {
  secret_id = "{{.ProjectName}}-storage-bucket-name"

  labels = local.common_labels

  replication {
    automatic = true
  }
}

resource "google_secret_manager_secret_version" "storage_bucket_name" {
  secret      = google_secret_manager_secret.storage_bucket_name.id
  secret_data = google_storage_bucket.storage.name
}

{{if eq .AppType "nodejs" "flask"}}
# IAM permissions for accessing storage secrets
resource "google_secret_manager_secret_iam_member" "storage_bucket_name_access" {
  secret_id = google_secret_manager_secret.storage_bucket_name.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.app.email}"
}
{{else if eq .AppType "docker"}}
resource "google_secret_manager_secret_iam_member" "storage_bucket_name_access" {
  secret_id = google_secret_manager_secret.storage_bucket_name.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.storage_sa.email}"
}
{{end}}

{{end}} 