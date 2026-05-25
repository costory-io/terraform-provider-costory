#!/usr/bin/env bash

set -euo pipefail

provider_name="${TF_PROVIDER_NAME:-costory}"

echo "Generating Terraform provider docs with tfplugindocs (pinned in go.mod) for provider ${provider_name}..."
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
