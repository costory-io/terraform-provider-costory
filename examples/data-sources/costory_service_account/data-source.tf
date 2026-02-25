provider "costory" {
  token = var.costory_token
}

data "costory_service_account" "current" {}

output "service_account" {
  value = data.costory_service_account.current.service_account
}

output "sub_ids" {
  value = data.costory_service_account.current.sub_ids
}
