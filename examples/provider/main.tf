terraform {
  required_providers {
    costory = {
      source = "costory-io/costory"
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
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
  default     = "http://localhost:8000"
}

variable "gcp_project_id" {
  type        = string
  description = "GCP project ID used for billing export access."
}

variable "bq_dataset_id" {
  type        = string
  description = "BigQuery dataset ID containing billing export table."
}

variable "bq_table_id" {
  type        = string
  description = "BigQuery billing export table ID."
}

variable "billing_datasource_name" {
  type        = string
  description = "Display name for the Costory billing datasource."
  default     = "costory-gcp-billing"
}

provider "costory" {
  slug     = var.costory_slug
  token    = var.costory_token
  base_url = var.costory_base_url
}

provider "google" {
  project = var.gcp_project_id
}

data "costory_service_account" "current" {}

locals {
  costory_dataset_roles = toset([
    "roles/bigquery.dataViewer",
    "roles/bigquery.metadataViewer",
  ])
}

resource "google_bigquery_dataset_iam_member" "costory_dataset_access" {
  for_each   = local.costory_dataset_roles
  project    = var.gcp_project_id
  dataset_id = var.bq_dataset_id
  role       = each.value
  member     = "serviceAccount:${data.costory_service_account.current.service_account}"
}

resource "costory_billing_datasource_aws" "example" {
  name                   = "costory-aws-billing"
  bucket_name            = "costory-cur-29-12-3-381492251657"
  prefix                 = "costory-cur//costory-cur-costory-cur-29-12-3"
  role_arn               = "arn:aws:iam::381492251657:role/costory-trust"
  eks_split_data_enabled = true
}

output "service_account" {
  value = data.costory_service_account.current.service_account
}

output "sub_ids" {
  value = data.costory_service_account.current.sub_ids
}
