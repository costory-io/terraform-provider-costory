# COS-1643: Elastic Cloud billing datasource (Terraform)

## Goal

Expose **Elastic Cloud** as a first-class billing datasource in the Costory Terraform provider, aligned with the app/API (`type: "ElasticCloud"`, credentials `apiKey` + `organizationId`, optional `startDate`, computed `bq_table_uri`, `status`, etc.).

## Approach

1. **API client (`internal/costoryapi`)**  
   - Add `billingDatasourceTypeElasticCloud = "ElasticCloud"`.  
   - Introduce request/response structs with JSON tags matching the Costory Terraform API: `apiKey`, `organizationId`, `startDate`, `bqTableUri`, `status`.  
   - Register validate/create/get endpoints (same URL patterns as other billing datasources).  
   - Implement `ValidateElasticCloudBillingDatasource`, `CreateElasticCloudBillingDatasource`, `GetElasticCloudBillingDatasource`.

2. **Terraform resource**  
   - New resource `costory_billing_datasource_elastic_cloud`, modeled on `costory_billing_datasource_anthropic`: immutable inputs with `RequiresReplace`, sensitive `api_key`, required `organization_id`, optional `start_date`, computed `id`, `type`, `status`, `bq_table_uri`.

3. **Provider registration**  
   - Register the resource in `internal/provider/provider.go`.

4. **Tests**  
   - Add an httptest CRUD test for the Elastic Cloud client (mirror Anthropic external test).

5. **Docs and examples**  
   - Add `examples/resources/costory_billing_datasource_elastic_cloud/resource.tf`.  
   - Run `scripts/generate-docs.sh` to produce `docs/resources/billing_datasource_elastic_cloud.md`.  
   - Update `README.md` feature list and optional usage snippet.

## Out of scope

- UX docs (`ux_impact.md` / `ux_proposal.md`) — not requested for this backend/IaC change.
