provider "costory" {
  token = var.costory_token
}

data "costory_service_account" "current" {}

locals {
  bigquery_roles = toset([
    "roles/bigquery.metadataViewer",
    "roles/bigquery.dataViewer",
  ])
}

resource "google_bigquery_dataset_iam_member" "costory_access" {
  for_each   = local.bigquery_roles
  dataset_id = "billing_export"
  role       = each.key
  member     = "serviceAccount:${data.costory_service_account.current.service_account}"
}

resource "costory_billing_datasource_gcp" "main" {
  name                = "GCP Billing Export"
  bq_uri              = "my-project.billing_export.gcp_billing_export_v1_0123"
  is_detailed_billing = true
  depends_on          = [google_bigquery_dataset_iam_member.costory_access]
}
