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

variable "anthropic_admin_api_key" {
  type        = string
  description = "Anthropic admin API key used to fetch billing data."
  sensitive   = true
}

provider "costory" {
  token = var.costory_api_token
}

resource "costory_billing_datasource_anthropic" "main" {
  name          = "Anthropic Billing"
  admin_api_key = var.anthropic_admin_api_key
}
