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

variable "elastic_cloud_api_key" {
  type        = string
  description = "Elastic Cloud API key used to fetch billing data."
  sensitive   = true
}

variable "elastic_cloud_organization_id" {
  type        = string
  description = "Elastic Cloud organization ID."
}

provider "costory" {
  token = var.costory_api_token
}

resource "costory_billing_datasource_elastic_cloud" "main" {
  name             = "Elastic Cloud Billing"
  api_key          = var.elastic_cloud_api_key
  organization_id  = var.elastic_cloud_organization_id
}
