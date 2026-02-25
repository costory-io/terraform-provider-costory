#!/usr/bin/env bash

set -euo pipefail

tfplugindocs_version="${TFPLUGINDOCS_VERSION:-latest}"
provider_name="${TF_PROVIDER_NAME:-costory}"

echo "Generating Terraform provider docs with tfplugindocs@${tfplugindocs_version} for provider ${provider_name}..."
go run "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@${tfplugindocs_version}" generate \
  --provider-name "${provider_name}"
