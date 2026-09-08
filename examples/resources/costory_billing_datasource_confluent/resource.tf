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

variable "confluent_api_key" {
  type        = string
  description = "Confluent Cloud API key."
  sensitive   = true
}

variable "confluent_api_secret" {
  type        = string
  description = "Confluent Cloud API secret."
  sensitive   = true
}

provider "costory" {
  token = var.costory_api_token
}

resource "costory_billing_datasource_confluent" "main" {
  name       = "Confluent Billing"
  api_key    = var.confluent_api_key
  api_secret = var.confluent_api_secret
}
