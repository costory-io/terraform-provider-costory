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

resource "costory_team" "engineering" {
  name       = "Engineering"
  visibility = "PRIVATE"
}

variable "team_members" {
  type = map(object({
    email = string
    role  = string
  }))
  default = {
    alice = { email = "alice@example.com", role = "OWNER" }
    bob   = { email = "bob@example.com", role = "MEMBER" }
    carol = { email = "carol@example.com", role = "MEMBER" }
  }
}

resource "costory_team_member" "members" {
  for_each = var.team_members

  team_id = costory_team.engineering.id
  email   = each.value.email
  role    = each.value.role
}
