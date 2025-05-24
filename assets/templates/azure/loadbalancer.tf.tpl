# Azure Application Gateway Configuration
{{if .EnableLoadBalancer}}

{{if ne .AppType "static-site"}}
# Public IP for Application Gateway
resource "azurerm_public_ip" "app_gateway" {
  name                = "pip-{{.ProjectName}}-agw"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  allocation_method   = "Static"
  sku                 = "Standard"

  tags = azurerm_resource_group.main.tags
}

# Subnet for Application Gateway
resource "azurerm_subnet" "app_gateway" {
  name                 = "subnet-{{.ProjectName}}-agw"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = ["10.0.2.0/24"]
}

# Application Gateway
resource "azurerm_application_gateway" "main" {
  name                = "agw-{{.ProjectName}}"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location

  sku {
    name     = "Standard_v2"
    tier     = "Standard_v2"
    capacity = 1
  }

  gateway_ip_configuration {
    name      = "gateway-ip-config"
    subnet_id = azurerm_subnet.app_gateway.id
  }

  frontend_port {
    name = "frontend-port"
    port = 80
  }

  frontend_ip_configuration {
    name                 = "frontend-ip-config"
    public_ip_address_id = azurerm_public_ip.app_gateway.id
  }

  {{if eq .AppType "nodejs" "flask"}}
  backend_address_pool {
    name         = "backend-pool"
    ip_addresses = [azurerm_network_interface.main.private_ip_address]
  }
  {{else if eq .AppType "docker"}}
  backend_address_pool {
    name  = "backend-pool"
    fqdns = [azurerm_container_group.app.fqdn]
  }
  {{end}}

  backend_http_settings {
    name                  = "backend-http-settings"
    cookie_based_affinity = "Disabled"
    path                  = "/"
    port                  = 80
    protocol              = "Http"
    request_timeout       = 60
    
    probe_name = "health-probe"
  }

  probe {
    name                = "health-probe"
    protocol            = "Http"
    path                = "/"
    interval            = 30
    timeout             = 30
    unhealthy_threshold = 3
    
    {{if eq .AppType "nodejs" "flask"}}
    host = azurerm_public_ip.main.ip_address
    {{else if eq .AppType "docker"}}
    host = azurerm_container_group.app.fqdn
    {{end}}
  }

  http_listener {
    name                           = "http-listener"
    frontend_ip_configuration_name = "frontend-ip-config"
    frontend_port_name             = "frontend-port"
    protocol                       = "Http"
  }

  request_routing_rule {
    name                       = "routing-rule"
    rule_type                  = "Basic"
    http_listener_name         = "http-listener"
    backend_address_pool_name  = "backend-pool"
    backend_http_settings_name = "backend-http-settings"
    priority                   = 100
  }

  tags = azurerm_resource_group.main.tags
}

{{else}}
# For static sites, use Azure CDN with Storage Account
resource "azurerm_cdn_profile" "main" {
  name                = "cdn-{{.ProjectName}}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  sku                 = "Standard_Microsoft"

  tags = azurerm_resource_group.main.tags
}

resource "azurerm_cdn_endpoint" "main" {
  name                = "cdn-{{.ProjectName}}-${random_id.suffix.hex}"
  profile_name        = azurerm_cdn_profile.main.name
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name

  origin {
    name      = "storage-origin"
    host_name = azurerm_storage_account.static_site.primary_web_host
  }

  origin_host_header = azurerm_storage_account.static_site.primary_web_host

  delivery_rule {
    name  = "DefaultRule"
    order = 1

    request_uri_condition {
      operator     = "Any"
      match_values = []
    }

    cache_expiration_action {
      behavior = "Override"
      duration = "1.00:00:00"
    }
  }

  tags = azurerm_resource_group.main.tags
}
{{end}}

{{end}} 