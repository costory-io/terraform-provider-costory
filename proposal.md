# COS-1844: Update Azure Terraform onboarding examples

## Goal

Align the Azure billing datasource Terraform example with feedback from a customer onboarding run: provider version, export blob paths, and required `partitionData` on Cost Management exports.

## Changes

1. **Provider version** — Replace deprecated `>= 0.1.0` with `~> 0.2` so examples do not resolve to unsupported 0.1.x releases.

2. **`partitionData`** — Set `partitionData = true` on both `azapi_resource` Cost Management exports (`actuals` and `amortized`). Required by current Azure/azapi APIs.

3. **Export paths** — With partitioned exports, blobs land under `{rootFolderPath}/{exportName}/`. Introduce locals for export names and set:
   - `actuals_path = "actuals/${local.actuals_export_name}"`
   - `amortized_path = "amortized/${local.amortized_export_name}"`

4. **Infrastructure-only example** — Apply `partitionData` to `examples/resources/costory_azure_datasource/resource.tf` for consistency.

5. **Docs** — Regenerate `docs/resources/billing_datasource_azure.md` from the example via `scripts/generate-docs.sh`.

## Out of scope

- Other billing datasource examples (GCP, AWS, etc.) — not mentioned in the issue.
- Provider code or schema changes.
