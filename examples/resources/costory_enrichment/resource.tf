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
  description = "Costory API key of an organization admin."
  sensitive   = true
}

provider "costory" {
  token = var.costory_api_token
}

# One resource per YAML file in ./enrichments. Applying stores drafts;
# costory_enrichments_publish makes them live.
resource "costory_enrichment" "all" {
  for_each = fileset("${path.module}/enrichments", "*.yaml")

  name = trimsuffix(each.value, ".yaml")
  file = "${path.module}/enrichments/${each.value}"
}
