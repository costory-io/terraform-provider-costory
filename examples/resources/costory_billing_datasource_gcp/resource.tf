provider "costory" {
  token = var.costory_token
}

resource "costory_billing_datasource_gcp" "main" {
  name                = "GCP Billing Export"
  bq_uri              = "my-project.billing_export.gcp_billing_export_v1_0123"
  is_detailed_billing = true
}
