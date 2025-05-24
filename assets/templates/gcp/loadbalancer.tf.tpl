# HTTP(S) Load Balancer Configuration
{{if .EnableLoadBalancer}}

{{if ne .AppType "static-site"}}
# Health check
resource "google_compute_health_check" "app" {
  name               = "{{.ProjectName}}-health-check"
  timeout_sec        = 5
  check_interval_sec = 10

  http_health_check {
    port         = 80
    request_path = "/"
  }
}

{{if eq .AppType "nodejs" "flask"}}
# Instance group for compute instances
resource "google_compute_instance_group" "app" {
  name        = "{{.ProjectName}}-ig"
  description = "Instance group for {{.ProjectName}}"
  zone        = "{{.Region}}-a"

  instances = [google_compute_instance.app.id]

  named_port {
    name = "http"
    port = "80"
  }
}

# Backend service
resource "google_compute_backend_service" "app" {
  name        = "{{.ProjectName}}-backend-service"
  description = "Backend service for {{.ProjectName}}"

  protocol    = "HTTP"
  port_name   = "http"
  timeout_sec = 30

  backend {
    group = google_compute_instance_group.app.id
  }

  health_checks = [google_compute_health_check.app.id]
}

{{else if eq .AppType "docker"}}
# Network Endpoint Group for Cloud Run
resource "google_compute_region_network_endpoint_group" "app" {
  name                  = "{{.ProjectName}}-neg"
  network_endpoint_type = "SERVERLESS"
  region                = "{{.Region}}"

  cloud_run {
    service = google_cloud_run_service.app.name
  }
}

# Backend service for Cloud Run
resource "google_compute_backend_service" "app" {
  name        = "{{.ProjectName}}-backend-service"
  description = "Backend service for {{.ProjectName}} Cloud Run"

  protocol            = "HTTP"
  load_balancing_scheme = "EXTERNAL"

  backend {
    group = google_compute_region_network_endpoint_group.app.id
  }
}
{{end}}

# URL map
resource "google_compute_url_map" "app" {
  name            = "{{.ProjectName}}-url-map"
  description     = "URL map for {{.ProjectName}}"
  default_service = google_compute_backend_service.app.id
}

# HTTP proxy
resource "google_compute_target_http_proxy" "app" {
  name    = "{{.ProjectName}}-http-proxy"
  url_map = google_compute_url_map.app.id
}

# Global IP address
resource "google_compute_global_address" "app" {
  name         = "{{.ProjectName}}-ip"
  description  = "Global IP address for {{.ProjectName}}"
  address_type = "EXTERNAL"
}

# Global forwarding rule
resource "google_compute_global_forwarding_rule" "app" {
  name                  = "{{.ProjectName}}-forwarding-rule"
  description           = "Global forwarding rule for {{.ProjectName}}"
  ip_protocol           = "TCP"
  load_balancing_scheme = "EXTERNAL"
  port_range            = "80"
  target                = google_compute_target_http_proxy.app.id
  ip_address            = google_compute_global_address.app.id
}

{{else}}
# For static sites, use Cloud CDN with Cloud Storage backend
resource "google_compute_backend_bucket" "static_site" {
  name        = "{{.ProjectName}}-backend-bucket"
  description = "Backend bucket for {{.ProjectName}} static site"
  bucket_name = google_storage_bucket.static_site.name
  enable_cdn  = true

  cdn_policy {
    cache_mode                   = "CACHE_ALL_STATIC"
    default_ttl                  = 3600
    max_ttl                      = 86400
    client_ttl                   = 3600
    negative_caching             = true
    serve_while_stale            = 86400
  }
}

# URL map for static site
resource "google_compute_url_map" "static_site" {
  name            = "{{.ProjectName}}-url-map"
  description     = "URL map for {{.ProjectName}} static site"
  default_service = google_compute_backend_bucket.static_site.id
}

# HTTP proxy for static site
resource "google_compute_target_http_proxy" "static_site" {
  name    = "{{.ProjectName}}-http-proxy"
  url_map = google_compute_url_map.static_site.id
}

# Global IP address for static site
resource "google_compute_global_address" "static_site" {
  name         = "{{.ProjectName}}-ip"
  description  = "Global IP address for {{.ProjectName}} static site"
  address_type = "EXTERNAL"
}

# Global forwarding rule for static site
resource "google_compute_global_forwarding_rule" "static_site" {
  name                  = "{{.ProjectName}}-forwarding-rule"
  description           = "Global forwarding rule for {{.ProjectName}} static site"
  ip_protocol           = "TCP"
  load_balancing_scheme = "EXTERNAL"
  port_range            = "80"
  target                = google_compute_target_http_proxy.static_site.id
  ip_address            = google_compute_global_address.static_site.id
}
{{end}}

# Store load balancer IP in Secret Manager
resource "google_secret_manager_secret" "lb_ip" {
  secret_id = "{{.ProjectName}}-lb-ip"

  labels = local.common_labels

  replication {
    automatic = true
  }
}

resource "google_secret_manager_secret_version" "lb_ip" {
  secret = google_secret_manager_secret.lb_ip.id
  {{if eq .AppType "static-site"}}
  secret_data = google_compute_global_address.static_site.address
  {{else}}
  secret_data = google_compute_global_address.app.address
  {{end}}
}

{{end}} 