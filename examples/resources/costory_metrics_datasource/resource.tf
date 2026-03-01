terraform {
  required_providers {
    costory = {
      source  = "costory-io/costory"
      version = ">= 0.1.0"
    }
  }
}

variable "costory_token" {
  type        = string
  description = "Costory API token."
  sensitive   = true
}

variable "s3_bucket_name" {
  type        = string
  description = "S3 bucket containing parquet metrics files."
}

variable "s3_prefix" {
  type        = string
  description = "S3 key prefix for parquet files."
  default     = ""
}

variable "role_arn" {
  type        = string
  description = "IAM role ARN for Costory to assume (must match arn:aws:iam::)."
}

provider "costory" {
  token = var.costory_token
}

resource "costory_metrics_datasource_s3_parquet" "main" {
  name        = "AWS S3 Metrics"
  bucket_name = var.s3_bucket_name
  prefix      = var.s3_prefix
  role_arn    = var.role_arn

  metrics_definition {
    metric_name  = "Usage"
    gap_filling  = "ZERO"
    aggregation  = "SUM"
    value_column = "usage_amount"
    date_column  = "usage_start_date"
    dimensions   = ["service", "region"]
    unit         = "count"
  }

  metrics_definition {
    metric_name  = "Cost"
    gap_filling  = "FORWARD_FILL"
    aggregation  = "Average"
    value_column = "unblended_cost"
    date_column  = "usage_start_date"
    dimensions   = ["service"]
  }
}
