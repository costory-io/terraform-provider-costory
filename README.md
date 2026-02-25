# Terraform Provider Costory

This repository contains the Terraform provider for Costory, built with the
[Terraform Plugin Framework](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework?product_intent=terraform).

The provider currently supports:

- Configure provider with `token`
- Setup Costory:
  - service-account discovery (`data.costory_service_account`)
  - GCP billing datasource lifecycle (`resource.costory_billing_datasource_gcp`)
  - AWS billing datasource lifecycle (`resource.costory_billing_datasource_aws`)

## Prerequisites

Install:

- [Go](https://go.dev/doc/install) `>= 1.24`
- [Terraform CLI](https://developer.hashicorp.com/terraform/downloads)
- [golangci-lint](https://golangci-lint.run/welcome/install/) (for local linting)

---

## Local setup

```bash
git clone https://github.com/costory-io/costory-terraform.git
cd costory-terraform
go mod download
```

Build provider binary:

```bash
go build -o bin/terraform-provider-costory
```

---

## Run checks locally

Formatting:

```bash
gofmt -w .
```

Lint:

```bash
golangci-lint run
```

Type/quality checks and compilation:

```bash
go vet ./...
go test ./...
go build ./...
```

---

## Example usage

This minimal example is self-contained:

```hcl
terraform {
  required_providers {
    costory = {
      source  = "costory-io/costory"
      version = ">= 0.1.0"
    }
  }
}

provider "costory" {
  token = var.costory_token
}

data "costory_service_account" "current" {}

resource "costory_billing_datasource_gcp" "main" {
  name                = "GCP Billing Export"
  bq_uri              = "my-project.billing_export.gcp_billing_export_v1_0123"
  is_detailed_billing = true
}

resource "costory_billing_datasource_aws" "main" {
  name        = "AWS CUR"
  bucket_name = "my-cur-bucket"
  role_arn    = resource.aws_iam_role.costory_billing_read.arn
  prefix      = "cur/"
}

output "service_account" {
  value = data.costory_service_account.current.service_account
}

output "sub_ids" {
  value = data.costory_service_account.current.sub_ids
}
```

---
