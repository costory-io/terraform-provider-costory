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

variable "clickhouse_cloud_key_secret" {
  type        = string
  description = "ClickHouse Cloud API key secret."
  sensitive   = true
}

provider "costory" {
  token = var.costory_api_token
}

resource "costory_billing_datasource_clickhouse_cloud" "main" {
  name            = "ClickHouse Cloud Billing"
  key_id          = "key-id"
  key_secret      = var.clickhouse_cloud_key_secret
  organization_id = "org-123"
}
