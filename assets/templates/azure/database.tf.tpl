# Azure Database for MySQL Configuration
{{if .EnableDatabase}}

# Random password for database
resource "random_password" "db_password" {
  length  = 16
  special = true
}

# MySQL Server
resource "azurerm_mysql_flexible_server" "main" {
  name                   = "mysql-{{.ProjectName}}-${random_id.suffix.hex}"
  resource_group_name    = azurerm_resource_group.main.name
  location               = azurerm_resource_group.main.location
  administrator_login    = "adminuser"
  administrator_password = random_password.db_password.result
  
  backup_retention_days        = 7
  geo_redundant_backup_enabled = false
  
  sku_name   = "B_Standard_B1s"
  storage {
    size_gb = 20
    auto_grow_enabled = true
  }

  tags = azurerm_resource_group.main.tags
}

# Database
resource "azurerm_mysql_flexible_database" "main" {
  name                = "{{.ProjectName | replace "-" "_"}}"
  resource_group_name = azurerm_resource_group.main.name
  server_name         = azurerm_mysql_flexible_server.main.name
  charset             = "utf8"
  collation           = "utf8_unicode_ci"
}

# Firewall rule to allow Azure services
resource "azurerm_mysql_flexible_server_firewall_rule" "azure_services" {
  name                = "AllowAzureServices"
  resource_group_name = azurerm_resource_group.main.name
  server_name         = azurerm_mysql_flexible_server.main.name
  start_ip_address    = "0.0.0.0"
  end_ip_address      = "0.0.0.0"
}

# Firewall rule to allow all IPs (for development)
resource "azurerm_mysql_flexible_server_firewall_rule" "allow_all" {
  name                = "AllowAll"
  resource_group_name = azurerm_resource_group.main.name
  server_name         = azurerm_mysql_flexible_server.main.name
  start_ip_address    = "0.0.0.0"
  end_ip_address      = "255.255.255.255"
}

# Key Vault for storing database credentials
resource "azurerm_key_vault" "main" {
  name                        = "kv-{{.ProjectName}}-${random_id.suffix.hex}"
  location                    = azurerm_resource_group.main.location
  resource_group_name         = azurerm_resource_group.main.name
  enabled_for_disk_encryption = true
  tenant_id                   = data.azurerm_client_config.current.tenant_id
  soft_delete_retention_days  = 7
  purge_protection_enabled    = false

  sku_name = "standard"

  access_policy {
    tenant_id = data.azurerm_client_config.current.tenant_id
    object_id = data.azurerm_client_config.current.object_id

    key_permissions = [
      "Get",
    ]

    secret_permissions = [
      "Get",
      "Set",
      "Delete",
      "Purge",
      "Recover"
    ]

    storage_permissions = [
      "Get",
    ]
  }

  tags = azurerm_resource_group.main.tags
}

# Get current client configuration
data "azurerm_client_config" "current" {}

# Store database password in Key Vault
resource "azurerm_key_vault_secret" "db_password" {
  name         = "db-password"
  value        = random_password.db_password.result
  key_vault_id = azurerm_key_vault.main.id
}

# Store database connection string in Key Vault
resource "azurerm_key_vault_secret" "db_connection_string" {
  name         = "db-connection-string"
  value        = "mysql://${azurerm_mysql_flexible_server.main.administrator_login}:${random_password.db_password.result}@${azurerm_mysql_flexible_server.main.fqdn}:3306/${azurerm_mysql_flexible_database.main.name}"
  key_vault_id = azurerm_key_vault.main.id
}

{{end}} 