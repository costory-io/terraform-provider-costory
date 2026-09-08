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

variable "anthropic_claude_ai_analytics_api_key" {
  type        = string
  description = "Anthropic Claude AI analytics API key used to fetch billing data."
  sensitive   = true
}

provider "costory" {
  token = var.costory_api_token
}

resource "costory_billing_datasource_anthropic_claude_ai" "main" {
  name              = "Anthropic Claude AI Billing"
  analytics_api_key = var.anthropic_claude_ai_analytics_api_key
  account_name      = "acme"
}
