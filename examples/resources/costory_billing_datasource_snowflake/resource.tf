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

resource "costory_billing_datasource_snowflake" "main" {
  name           = "Snowflake Billing"
  integration_id = "snowflake-integration-id"
}
