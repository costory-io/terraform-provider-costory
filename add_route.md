# Endpoints Needed For Billing Datasource GCP

This file lists the backend endpoints expected by the Terraform provider for GCP billing datasource support.

## Base assumptions

- Base URL is provider-configured; paths below are relative.
- Auth headers expected by existing client:
  - `Authorization: Bearer <token>`
  - `X-Costory-Slug: <slug>`
  - `Accept: application/json`
  - `Content-Type: application/json` for JSON body requests

## 1) Validate GCP billing datasource

- Method: `POST`
- Path: `/terraform/billing-datasource/validate`
- Purpose: pre-check datasource configuration before creation
- Request body:

```json
{
  "type": "GCP",
  "name": "string",
  "bqTablePath": "project.dataset.table",
  "isDetailedBilling": true,
  "startDate": "YYYY-MM-DD",
  "endDate": "YYYY-MM-DD"
}
```

Notes:
- `isDetailedBilling`, `startDate`, and `endDate` are optional.
- Provider accepts any `2xx` response as success for validation.
- Validation can return a structured payload (for example `ValidationResult`).

## 2) Create GCP billing datasource

- Method: `POST`
- Path: `/terraform/billing-datasource`
- Purpose: create datasource and return created object
- Request body: same as validation payload
- Response body (expected):

```json
{
  "id": "datasource-id",
  "type": "GCP",
  "name": "string",
  "bqUri": "project.dataset.table",
  "isDetailedBilling": true,
  "startDate": "YYYY-MM-DD",
  "endDate": "YYYY-MM-DD"
}
```

Compatibility notes:
- Create and read responses use `bqUri` in Terraform-facing payloads.
- Provider expects an `id` in create response.
- Any non-`2xx` is treated as error.

## 3) Read billing datasource by ID

- Method: `GET`
- Path: `/terraform/billing-datasource/{id}`
- Purpose: refresh Terraform state and support import reads
- Response body: same shape as create response

Status handling expected by provider:
- `200`: success
- `404`: not found (resource removed from Terraform state)
- other statuses: error

## 4) Delete billing datasource by ID

- Method: `DELETE`
- Path: `/terraform/billing-datasource/{id}`
- Purpose: delete datasource

Status handling expected by provider:
- `204` or `200`: success
- `404`: treated as already deleted
- other statuses: error

## 5) Service account route (related, not datasource CRUD)

- Method: `GET`
- Path: `/terraform/`
- Purpose: return setup information used by provider examples/onboarding
- Response can include any of these compatible keys:
  - service account: `service_account` or `serviceAccount` or `serviceAccountEmail`
  - sub IDs: `sub_ids` or `subIds`

## Import behavior

There is no dedicated import endpoint.

Terraform import stores the ID and then calls:
- `GET /terraform/billing-datasource/{id}`
