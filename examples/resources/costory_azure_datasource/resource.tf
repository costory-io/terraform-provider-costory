terraform {
  required_version = ">= 1.5"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
    }
    azuread = {
      source  = "hashicorp/azuread"
      version = "~> 3.0"
    }
    azapi = {
      source  = "Azure/azapi"
      version = "~> 2.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.11"
    }
  }
}

provider "azurerm" {
  features {}
  subscription_id = var.subscription_id
}

provider "azuread" {}

# ──────────────────────────────────────────────
# Variables
# ──────────────────────────────────────────────

variable "subscription_id" {
  type        = string
  description = "Azure subscription ID."
}

variable "costory_google_unique_id" {
  type        = string
  description = "Costory Google service account unique ID used as the federated credential subject. Copy it from Connect Azure in Costory, or from data.costory_service_account.current.sub_ids[0]."
}

variable "location" {
  type        = string
  description = "Azure region for all resources."
  default     = "West Europe"
}

variable "resource_group_name" {
  type        = string
  description = "Name of the resource group to create."
  default     = "costory-cost-exports"
}

variable "storage_account_name_prefix" {
  type        = string
  description = "Prefix for the storage account name (a random suffix is appended for uniqueness)."
  default     = "costexports"
}

variable "backfill_month_count" {
  type        = number
  description = "Number of past months to backfill (0 to skip)."
  default     = 12
}

variable "run_backfill" {
  type        = bool
  description = "Set to true to trigger backfill runs. Use: terraform apply -var='run_backfill=true'"
  default     = false
}

data "azuread_client_config" "current" {}

resource "azurerm_resource_group" "cost_exports" {
  name     = var.resource_group_name
  location = var.location
}

resource "random_string" "storage_suffix" {
  length  = 8
  special = false
  upper   = false
}

resource "azurerm_storage_account" "cost_exports" {
  name                     = "${var.storage_account_name_prefix}${random_string.storage_suffix.result}"
  resource_group_name      = azurerm_resource_group.cost_exports.name
  location                 = azurerm_resource_group.cost_exports.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_container" "billing" {
  name               = "billing-exports"
  storage_account_id = azurerm_storage_account.cost_exports.id
}

resource "time_static" "export_start" {}

locals {
  actuals_export_name   = "costory-actuals-${random_string.storage_suffix.result}"
  amortized_export_name = "costory-amortized-${random_string.storage_suffix.result}"
}

resource "azapi_resource" "actuals" {
  type      = "Microsoft.CostManagement/exports@2025-03-01"
  name      = local.actuals_export_name
  parent_id = "/subscriptions/${var.subscription_id}"

  body = {
    properties = {
      definition = {
        type      = "ActualCost"
        timeframe = "MonthToDate"
        dataSet = {
          granularity = "Daily"
        }
      }
      schedule = {
        status     = "Active"
        recurrence = "Daily"
        recurrencePeriod = {
          from = time_static.export_start.rfc3339
          to   = "2099-01-01T00:00:00Z"
        }
      }
      format                = "Parquet"
      partitionData         = true
      dataOverwriteBehavior = "OverwritePreviousReport"
      deliveryInfo = {
        destination = {
          container      = azurerm_storage_container.billing.name
          resourceId     = azurerm_storage_account.cost_exports.id
          rootFolderPath = "actuals"
        }
      }
    }
  }
}

resource "azapi_resource" "amortized" {
  type      = "Microsoft.CostManagement/exports@2025-03-01"
  name      = local.amortized_export_name
  parent_id = "/subscriptions/${var.subscription_id}"

  body = {
    properties = {
      definition = {
        type      = "AmortizedCost"
        timeframe = "MonthToDate"
        dataSet = {
          granularity = "Daily"
        }
      }
      schedule = {
        status     = "Active"
        recurrence = "Daily"
        recurrencePeriod = {
          from = time_static.export_start.rfc3339
          to   = "2099-01-01T00:00:00Z"
        }
      }
      format                = "Parquet"
      partitionData         = true
      dataOverwriteBehavior = "OverwritePreviousReport"
      deliveryInfo = {
        destination = {
          container      = azurerm_storage_container.billing.name
          resourceId     = azurerm_storage_account.cost_exports.id
          rootFolderPath = "amortized"
        }
      }
    }
  }
}

locals {
  absolute_month = tonumber(formatdate("YYYY", plantimestamp())) * 12 + tonumber(formatdate("M", plantimestamp())) - 1

  backfill_months = [
    for i in range(1, var.backfill_month_count + 1) : format(
      "%04d-%02d",
      floor((local.absolute_month - i) / 12),
      (local.absolute_month - i) % 12 + 1,
    )
  ]

  backfill_ranges = {
    for m in local.backfill_months : m => {
      from = "${m}-01T00:00:00Z"
      to = "${formatdate(
        "YYYY-MM-DD",
        timeadd(
          format(
            "%04d-%02d-01T00:00:00Z",
            tonumber(split("-", m)[1]) == 12 ? tonumber(split("-", m)[0]) + 1 : tonumber(split("-", m)[0]),
            tonumber(split("-", m)[1]) == 12 ? 1 : tonumber(split("-", m)[1]) + 1,
          ),
          "-24h",
        ),
      )}T00:00:00Z"
    }
  }
}

resource "azapi_resource_action" "backfill_actuals" {
  for_each = var.run_backfill ? local.backfill_ranges : {}

  type        = "Microsoft.CostManagement/exports@2025-03-01"
  resource_id = azapi_resource.actuals.id
  action      = "run"
  method      = "POST"
  when        = "apply"

  body = {
    timePeriod = {
      from = each.value.from
      to   = each.value.to
    }
  }
}

resource "azapi_resource_action" "backfill_amortized" {
  for_each = var.run_backfill ? local.backfill_ranges : {}

  type        = "Microsoft.CostManagement/exports@2025-03-01"
  resource_id = azapi_resource.amortized.id
  action      = "run"
  method      = "POST"
  when        = "apply"

  body = {
    timePeriod = {
      from = each.value.from
      to   = each.value.to
    }
  }
}

# ──────────────────────────────────────────────
# Azure AD federated identity
# ──────────────────────────────────────────────

resource "azuread_application" "costory" {
  display_name = "costory-billing-${random_string.storage_suffix.result}"
}

resource "azuread_service_principal" "costory" {
  client_id = azuread_application.costory.client_id
}

resource "azuread_application_federated_identity_credential" "costory" {
  application_id = azuread_application.costory.id
  display_name   = "costory-gcp"
  description    = "Allow Costory to read Azure billing exports via Google federated identity."
  audiences      = ["api://AzureADTokenExchange"]
  issuer         = "https://accounts.google.com"
  subject        = var.costory_google_unique_id
}

resource "azurerm_role_assignment" "costory_blob_reader" {
  scope                            = "${azurerm_storage_account.cost_exports.id}/blobServices/default/containers/${azurerm_storage_container.billing.name}"
  role_definition_name             = "Storage Blob Data Reader"
  principal_id                     = azuread_service_principal.costory.object_id
  skip_service_principal_aad_check = true
}

output "resource_group_name" {
  value = azurerm_resource_group.cost_exports.name
}

output "storage_account_name" {
  value = azurerm_storage_account.cost_exports.name
}

output "storage_container_name" {
  value = azurerm_storage_container.billing.name
}

output "client_id" {
  description = "Azure AD application (client) ID for Costory federated identity."
  value       = azuread_application.costory.client_id
}

output "tenant_id" {
  description = "Azure AD tenant ID for Costory federated identity."
  value       = data.azuread_client_config.current.tenant_id
}
