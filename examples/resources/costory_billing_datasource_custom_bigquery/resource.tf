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

provider "costory" {
  token = var.costory_api_token
}

resource "costory_billing_datasource_custom_bigquery" "main" {
  name          = "Custom BigQuery Billing"
  bq_table_path = "my-project.custom.costs"
  provider_name = "acme"

  mapping = {
    billed_cost  = "cost"
    service_name = "sku"
  }
}
