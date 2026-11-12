package billingdatasource

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUpgradeAzureStateV0toV1DropsSASURL(t *testing.T) {
	t.Parallel()

	prior := azureResourceModelV0{
		ID:                 types.StringValue("azure-ds-1"),
		Status:             types.StringValue("ACTIVE"),
		Name:               types.StringValue("Azure Billing"),
		SASURL:             types.StringValue("https://account.blob.core.windows.net/billing-exports?sv=2024-01-01&sig=example"),
		StorageAccountName: types.StringValue("storageaccount"),
		ContainerName:      types.StringValue("billing-exports"),
		ActualsPath:        types.StringValue("actuals"),
		AmortizedPath:      types.StringValue("amortized"),
	}

	got := upgradeAzureStateV0toV1(prior)

	if got.ID.ValueString() != "azure-ds-1" {
		t.Fatalf("unexpected id: got %q", got.ID.ValueString())
	}

	if !got.ClientID.IsNull() || !got.TenantID.IsNull() {
		t.Fatalf("expected client_id and tenant_id to be null after SAS state upgrade, got %q / %q", got.ClientID.ValueString(), got.TenantID.ValueString())
	}

	if got.Name.ValueString() != "Azure Billing" || got.StorageAccountName.ValueString() != "storageaccount" {
		t.Fatalf("expected unchanged identity fields, got name=%q storage=%q", got.Name.ValueString(), got.StorageAccountName.ValueString())
	}

	if got.ContainerName.ValueString() != "billing-exports" || got.ActualsPath.ValueString() != "actuals" || got.AmortizedPath.ValueString() != "amortized" {
		t.Fatalf("expected unchanged export fields")
	}
}
