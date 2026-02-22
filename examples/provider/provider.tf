terraform {
  required_providers {
    costory = {
      source  = "costory-io/costory"
      version = ">= 0.1.0"
    }
  }
}

variable "costory_slug" {
  type        = string
  description = "Costory tenant slug."
}

variable "costory_token" {
  type        = string
  description = "Costory API token."
  sensitive   = true
}

variable "costory_base_url" {
  type        = string
  description = "Costory API base URL."
  default     = "https://app.costory.io"
}

provider "costory" {
  slug     = var.costory_slug
  token    = var.costory_token
  base_url = var.costory_base_url
}

data "costory_service_account" "current" {}

output "service_account" {
  value = data.costory_service_account.current.service_account
}

output "sub_ids" {
  value = data.costory_service_account.current.sub_ids
}
