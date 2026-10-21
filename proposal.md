# COS-1968: Add Terraform pending message for AWS imports

## Problem

During onboarding, users repeatedly relaunched AWS billing datasource imports while the first export was still in progress. AWS can take up to 24 hours to deliver the first billing export batch, but Terraform gave no signal to wait—unlike the web UI, which shows an "import in progress" message.

## Solution

Emit a Terraform diagnostic **warning** (not an error) when `costory_billing_datasource_aws` has `status = "PENDING"` after create or read. Warnings surface in `terraform apply` / `terraform plan` output without failing the run.

Use `resp.Diagnostics.AddWarning` with this detail text (per issue):

> AWS takes 12hours + to export the first batch of data. Costory will check tomorrow morning and will let you know by email when your domain is ready.

## Implementation

1. Add a small helper in `internal/provider/billingdatasource/aws_resource.go` that checks `status == "PENDING"` and appends the warning.
2. Call it at the end of `Create` and `Read`, after the API response is merged into state.
3. Add a unit test for the helper to lock in the warning behavior.

## Out of scope

- Warnings for other cloud providers (GCP, Azure, etc.)—issue is AWS-specific.
- Provider schema or API client changes.
- Docs regeneration (warning is runtime-only, not a schema attribute).
