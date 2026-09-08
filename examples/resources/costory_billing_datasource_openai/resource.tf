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

variable "openai_admin_api_key" {
  type        = string
  description = "OpenAI admin API key used to fetch billing data."
  sensitive   = true
}

provider "costory" {
  token = var.costory_api_token
}

resource "costory_billing_datasource_openai" "main" {
  name          = "OpenAI Billing"
  admin_api_key = var.openai_admin_api_key
}
