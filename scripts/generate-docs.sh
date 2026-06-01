#!/usr/bin/env bash

set -euo pipefail

provider_name="${TF_PROVIDER_NAME:-costory}"

if ! command -v terraform >/dev/null 2>&1; then
  echo "Error: terraform must be installed and available on PATH." >&2
  echo "tfplugindocs uses Terraform to export the provider schema." >&2
  echo "Automatic Terraform downloads can fail with: openpgp: key expired (hc-install < v0.9.4)." >&2
  echo "Install the CLI: https://developer.hashicorp.com/terraform/install" >&2
  exit 1
fi

echo "Generating Terraform provider docs with tfplugindocs (pinned in go.mod) for provider ${provider_name}..."
echo "Using $(terraform version | head -1)"
go tool tfplugindocs generate --provider-name "${provider_name}"
echo "Patching subcategories in generated docs..."
docs_dir="docs/resources"
for f in "${docs_dir}"/*.md; do
  name="$(basename "$f" .md)"
  case "$name" in
    billing_datasource_*) subcategory="Billing Datasources" ;;
    metrics_datasource_*) subcategory="Metrics Datasources" ;;
    team_*) subcategory="Teams" ;;
    *)                    continue ;;
  esac
  tmp="${f}.tmp"
  sed "s/^subcategory: \".*\"/subcategory: \"${subcategory}\"/" "$f" > "$tmp" && mv "$tmp" "$f"
done
