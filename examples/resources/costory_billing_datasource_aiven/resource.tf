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

variable "aiven_api_secret" {
  type        = string
  description = "Aiven API token used to fetch billing data."
  sensitive   = true
}

provider "costory" {
  token = var.costory_api_token
}

resource "costory_billing_datasource_aiven" "main" {
  name            = "Aiven Billing"
  api_secret      = var.aiven_api_secret
  organization_id = "aiven-organization-id"
}
