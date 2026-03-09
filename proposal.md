## Proposal: Cursor + Anthropic billing datasources

### Goal
Extend the Terraform provider to support Cursor and Anthropic billing datasources
using the existing `/terraform/billingDatasources` and `/terraform/billingDatasources/validate`
API endpoints.

### Planned changes
- Add new billing datasource resources:
  - `costory_billing_datasource_cursor`
  - `costory_billing_datasource_anthropic`
- Extend the Costory API client with request/response types and methods for
  Cursor + Anthropic create/validate/get flows.
- Mark `admin_api_key` as sensitive and optional `start_date`/`end_date` inputs.
- Persist state attributes: `id`, `name`, `type`, `bq_table_uri`, `start_date`,
  `end_date`, `status`.
- Register the new resources in the provider.
- Add API client tests and generate docs for the new resources.

### Validation
- Run `go test ./...` (and regenerate docs if schema changes).
