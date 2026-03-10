terraform {
  required_providers {
    costory = {
      source  = "costory-io/costory"
      version = ">= 0.1.0"
    }
  }
}

variable "costory_token" {
  type        = string
  description = "Costory API token."
  sensitive   = true
}

variable "azure_sas_url" {
  type        = string
  description = "Full Azure blob SAS URL including query string."
  sensitive   = true
}

variable "azure_storage_account_name" {
  type        = string
  description = "Azure storage account name hosting the export container."
}

variable "azure_container_name" {
  type        = string
  description = "Azure storage container with billing exports."
}

variable "azure_actuals_path" {
  type        = string
  description = "Path prefix for actual cost exports."
  default     = "actuals"
}

variable "azure_amortized_path" {
  type        = string
  description = "Path prefix for amortized cost exports."
  default     = "amortized"
}

provider "costory" {
  token = var.costory_token
}

resource "costory_azure_datasource" "main" {
  name                 = "Azure Billing"
  sas_url              = var.azure_sas_url
  storage_account_name = var.azure_storage_account_name
  container_name       = var.azure_container_name
  actuals_path         = var.azure_actuals_path
  amortized_path       = var.azure_amortized_path
}
