terraform {
  required_version = ">= 1.5"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
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

# ──────────────────────────────────────────────
# Variables
# ──────────────────────────────────────────────

variable "subscription_id" {
  type        = string
  description = "Azure subscription ID."
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

variable "sas_token_validity_days" {
  type        = number
  description = "Number of days the SAS token remains valid."
  default     = 900
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

resource "azapi_resource" "actuals" {
  type      = "Microsoft.CostManagement/exports@2025-03-01"
  name      = "costory-actuals-${random_string.storage_suffix.result}"
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
      format = "Parquet"
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
  name      = "costory-amortized-${random_string.storage_suffix.result}"
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
      format = "Parquet"
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
# SAS Token
# ──────────────────────────────────────────────

resource "time_static" "sas_start" {}

data "azurerm_storage_account_sas" "billing" {
  connection_string = azurerm_storage_account.cost_exports.primary_connection_string
  https_only        = true
  signed_version    = "2022-11-02"

  start  = time_static.sas_start.rfc3339
  expiry = timeadd(time_static.sas_start.rfc3339, "${var.sas_token_validity_days * 24}h")

  resource_types {
    service   = false
    container = true
    object    = true
  }

  services {
    blob  = true
    queue = false
    table = false
    file  = false
  }

  permissions {
    read    = true
    write   = false
    delete  = false
    list    = true
    add     = false
    create  = false
    update  = false
    process = false
    tag     = false
    filter  = false
  }
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

output "sas_token" {
  value     = data.azurerm_storage_account_sas.billing.sas
  sensitive = true
}

output "blob_endpoint_with_sas" {
  description = "Full blob endpoint URL with SAS token for accessing billing exports."
  value       = "${azurerm_storage_account.cost_exports.primary_blob_endpoint}${azurerm_storage_container.billing.name}${data.azurerm_storage_account_sas.billing.sas}"
  sensitive   = true
}

output "sas_token_expiry" {
  value = timeadd(time_static.sas_start.rfc3339, "${var.sas_token_validity_days * 24}h")
}
