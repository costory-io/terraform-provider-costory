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

variable "cursor_admin_api_key" {
  type        = string
  description = "Cursor admin API key used to fetch billing data."
  sensitive   = true
}

provider "costory" {
  token = var.costory_token
}

resource "costory_billing_datasource_cursor" "main" {
  name          = "Cursor Billing"
  admin_api_key = var.cursor_admin_api_key
}
