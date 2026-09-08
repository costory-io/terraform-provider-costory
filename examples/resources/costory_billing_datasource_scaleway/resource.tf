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

variable "scaleway_secret_key" {
  type        = string
  description = "Scaleway secret key used to fetch billing data."
  sensitive   = true
}

provider "costory" {
  token = var.costory_api_token
}

resource "costory_billing_datasource_scaleway" "main" {
  name            = "Scaleway Billing"
  secret_key      = var.scaleway_secret_key
  organization_id = "scaleway-organization-id"
}
