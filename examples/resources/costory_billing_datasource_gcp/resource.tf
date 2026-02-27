provider "costory" {
  token = var.costory_token
}

data "costory_service_account" "current" {}

locals {
  bigquery_roles = toset([
    "roles/bigquery.metadataViewer",
    "roles/bigquery.dataViewer",
  ])
  # Set the BigQuery project, dataset, and table IDs where your detailed billing data is exported.
  bigquery_project_id = "my-project"
  bigquery_dataset_id = "billing_export"
  bigquery_table_id   = "gcp_billing_export_v1_0123"
}

resource "google_bigquery_dataset_iam_member" "costory_access" {
  for_each   = local.bigquery_roles
  project    = local.bigquery_project_id
  dataset_id = local.bigquery_dataset_id
  role       = each.key
  member     = "serviceAccount:${data.costory_service_account.current.service_account}"
}

resource "costory_billing_datasource_gcp" "main" {
  name                = "GCP Billing Export"
  bq_uri              = "${local.bigquery_project_id}.${local.bigquery_dataset_id}.${local.bigquery_table_id}"
  is_detailed_billing = true
  depends_on          = [google_bigquery_dataset_iam_member.costory_access]
}
