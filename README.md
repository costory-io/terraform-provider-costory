# Terraform Provider Costory (Template)

This repository is a **template** for a custom Terraform provider built with the
[Terraform Plugin Framework](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework?product_intent=terraform).

The provider is intentionally simple:
- Configure provider with `slug` and `token`
- Send API calls to the Costory backend for:
  - service-account discovery (`data.costory_service_account`)
  - GCP billing datasource lifecycle (`resource.costory_billing_datasource_gcp`)
- Expose these values through a data source:
  - `service_account` (string)
  - `sub_ids` (list of string)

---

## What this template includes

- Go provider implementation using Terraform Plugin Framework
- `costory_service_account` data source
- `costory_billing_datasource_gcp` resource (create/read/delete/import)
- HTTP client abstraction and unit tests
- GitHub Actions CI workflow for:
  - formatting check (`gofmt`)
  - lint (`golangci-lint`)
  - vet (`go vet`)
  - tests (`go test`)
  - compilation (`go build`)

---

## API contract used by this template

Current implementation calls:

- `GET /api/v1/terraform/context`
- `POST /billing-datasources/terraform/validate`
- `POST /billing-datasources/terraform`
- `GET /billing-datasources/terraform/:id`
- `DELETE /billing-datasources/terraform/:id`

With headers:

- `Authorization: Bearer <token>`
- `X-Costory-Slug: <slug>`

Expected JSON response:

```json
{
  "service_account": "my-service-account",
  "sub_ids": ["sub-1", "sub-2"]
}
```

All API routes are centralized in `internal/provider/routes.go` so you can easily
swap route paths once backend endpoints are finalized.

---

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

See `examples/provider/provider.tf` for a full example.

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
  slug  = var.costory_slug
  token = var.costory_token
}

data "costory_service_account" "current" {}

resource "costory_billing_datasource_gcp" "main" {
  name                = "GCP Billing Export"
  bq_table_path       = "my-project.billing_export.gcp_billing_export_v1_0123"
  is_detailed_billing = true
}

output "service_account" {
  value = data.costory_service_account.current.service_account
}

output "sub_ids" {
  value = data.costory_service_account.current.sub_ids
}
```

---

## Development notes

- Entry point: `main.go`
- Provider config and registration: `internal/provider/provider.go`
- Costory API client: `internal/provider/client.go`
- API route definitions: `internal/provider/routes.go`
- Data source implementation: `internal/provider/context_data_source.go`
- GCP resource implementation: `internal/provider/gcp_billing_datasource_resource.go`
- CI workflow: `.github/workflows/ci.yml`