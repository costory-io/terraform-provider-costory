## Proposal: Azure billing datasource

### Goal
Extend the Terraform provider to support Azure billing datasources
using the `/terraform/billingDatasources` and `/terraform/billingDatasources/validate`
API endpoints.

### Planned changes
- Add `costory_billing_datasource_azure` resource mapping to the Azure payload.
- Extend the Costory API client with request/response types and methods for
  Azure create/validate/get/delete flows.
- Mark `sas_url` as sensitive and ForceNew for all fields (no update endpoint).
- Persist state attributes: `id`, `name`, `type`, `storage_account_name`,
  `container_name`, `actuals_path`, `amortized_path`, `status`.
- Register the new resource in the provider.
- Add acceptance tests for create/read/delete and validate.

### Validation
- Run relevant acceptance tests for the Azure datasource.
