package billingdatasource

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/costory-io/costory-terraform/internal/costoryapi"
)

func configureCostoryClient(req resource.ConfigureRequest, resp *resource.ConfigureResponse) *costoryapi.Client {
	if req.ProviderData == nil {
		return nil
	}

	client, ok := req.ProviderData.(*costoryapi.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource configure type",
			fmt.Sprintf("Expected *costoryapi.Client, got: %T. This is always a provider implementation bug.", req.ProviderData),
		)
		return nil
	}

	return client
}

func addUnconfiguredClientError(diagnostics *diag.Diagnostics) {
	diagnostics.AddError(
		"Unconfigured Costory client",
		"The provider did not configure the Costory API client for the resource.",
	)
}

func optionalString(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}

	result := value.ValueString()
	return &result
}

func applyString(dst *types.String, value string) {
	if value != "" {
		*dst = types.StringValue(value)
	}
}

func applyOptionalString(dst *types.String, value *string) {
	if value != nil {
		*dst = types.StringValue(*value)
	}
}

func applyStatus(dst *types.String, value *string) {
	*dst = types.StringNull()
	if value != nil {
		*dst = types.StringValue(*value)
	}
}

func deleteBillingDatasource(
	ctx context.Context,
	client *costoryapi.Client,
	diagnostics *diag.Diagnostics,
	datasourceID string,
	errorTitle string,
) {
	if client == nil {
		addUnconfiguredClientError(diagnostics)
		return
	}

	err := client.DeleteBillingDatasource(ctx, datasourceID)
	if err != nil && !errors.Is(err, costoryapi.ErrNotFound) {
		diagnostics.AddError(errorTitle, err.Error())
	}
}

func updateNotSupported(resp *resource.UpdateResponse, resourceName string) {
	resp.Diagnostics.AddError(
		"Update not supported",
		fmt.Sprintf("All attributes are immutable for %s. Terraform should replace the resource instead.", resourceName),
	)
}
