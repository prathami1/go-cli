# Azure Storage Configuration
{{if .EnableStorage}}

# Storage Account for application storage
resource "azurerm_storage_account" "storage" {
  name                     = "st{{.ProjectName | replace "-" ""}}storage${random_id.suffix.hex}"
  resource_group_name      = azurerm_resource_group.main.name
  location                 = azurerm_resource_group.main.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = azurerm_resource_group.main.tags
}

# Storage Container
resource "azurerm_storage_container" "app_data" {
  name                  = "app-data"
  storage_account_name  = azurerm_storage_account.storage.name
  container_access_type = "private"
}

# Lifecycle management policy
resource "azurerm_storage_management_policy" "storage" {
  storage_account_id = azurerm_storage_account.storage.id

  rule {
    name    = "rule1"
    enabled = true
    filters {
      prefix_match = ["app-data/"]
      blob_types   = ["blockBlob"]
    }
    actions {
      base_blob {
        tier_to_cool_after_days_since_modification_greater_than    = 30
        tier_to_archive_after_days_since_modification_greater_than = 90
        delete_after_days_since_modification_greater_than          = 365
      }
    }
  }
}

# Store storage account connection string in Key Vault
{{if .EnableDatabase}}
resource "azurerm_key_vault_secret" "storage_connection_string" {
  name         = "storage-connection-string"
  value        = azurerm_storage_account.storage.primary_connection_string
  key_vault_id = azurerm_key_vault.main.id
}

resource "azurerm_key_vault_secret" "storage_account_name" {
  name         = "storage-account-name"
  value        = azurerm_storage_account.storage.name
  key_vault_id = azurerm_key_vault.main.id
}
{{else}}
# Create a Key Vault for storage secrets if database is not enabled
resource "azurerm_key_vault" "storage" {
  name                        = "kv-{{.ProjectName}}-storage-${random_id.suffix.hex}"
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

# Get current client configuration (if not already defined)
data "azurerm_client_config" "current" {}

resource "azurerm_key_vault_secret" "storage_connection_string" {
  name         = "storage-connection-string"
  value        = azurerm_storage_account.storage.primary_connection_string
  key_vault_id = azurerm_key_vault.storage.id
}

resource "azurerm_key_vault_secret" "storage_account_name" {
  name         = "storage-account-name"
  value        = azurerm_storage_account.storage.name
  key_vault_id = azurerm_key_vault.storage.id
}
{{end}}

{{end}} 