terraform {
  required_providers {
    costory = {
      source  = "costory-io/costory"
      version = ">= 0.1.0"
    }
  }
}

variable "costory_api_token" {
  type        = string
  description = "Costory API token."
  sensitive   = true
}

variable "datadog_api_key" {
  type        = string
  description = "Datadog API key."
  sensitive   = true
}

variable "datadog_application_key" {
  type        = string
  description = "Datadog application key."
  sensitive   = true
}

provider "costory" {
  token = var.costory_api_token
}

resource "costory_billing_datasource_datadog" "main" {
  name            = "Datadog Billing"
  api_key         = var.datadog_api_key
  application_key = var.datadog_application_key
  region          = "datadoghq.com"
}
