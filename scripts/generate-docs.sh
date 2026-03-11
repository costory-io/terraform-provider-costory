#!/usr/bin/env bash

set -euo pipefail

tfplugindocs_version="${TFPLUGINDOCS_VERSION:-latest}"
provider_name="${TF_PROVIDER_NAME:-costory}"

echo "Generating Terraform provider docs with tfplugindocs@${tfplugindocs_version} for provider ${provider_name}..."
go run "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@${tfplugindocs_version}" generate \
  --provider-name "${provider_name}"
echo "Patching subcategories in generated docs..."
docs_dir="docs/resources"
for f in "${docs_dir}"/*.md; do
  name="$(basename "$f" .md)"
  case "$name" in
    billing_datasource_*) subcategory="Billing Datasources" ;;
    metrics_datasource_*) subcategory="Metrics Datasources" ;;
    *)                    continue ;;
  esac
  sed -i '' "s/^subcategory: \".*\"/subcategory: \"${subcategory}\"/" "$f"
done
