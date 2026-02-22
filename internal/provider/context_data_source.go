package provider

import (
	"context"
	"fmt"

	"github.com/costory-io/costory-terraform/internal/costoryapi"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &serviceAccountDataSource{}
	_ datasource.DataSourceWithConfigure = &serviceAccountDataSource{}
)

type serviceAccountDataSource struct {
	client *costoryapi.Client
}

type serviceAccountDataSourceModel struct {
	ServiceAccount types.String `tfsdk:"service_account"`
	SubIDs         types.List   `tfsdk:"sub_ids"`
}

// NewServiceAccountDataSource returns the Costory service-account data source.
func NewServiceAccountDataSource() datasource.DataSource {
	return &serviceAccountDataSource{}
}

func (d *serviceAccountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_service_account", req.ProviderTypeName)
}

func (d *serviceAccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Returns Costory service-account data from the Costory API.",
		Attributes: map[string]schema.Attribute{
			"service_account": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Service account name returned by Costory.",
			},
			"sub_ids": schema.ListAttribute{
				Computed:            true,
				MarkdownDescription: "Subscription IDs returned by Costory.",
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *serviceAccountDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*costoryapi.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected data source configure type",
			fmt.Sprintf("Expected *costoryapi.Client, got: %T. This is always a provider implementation bug.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *serviceAccountDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the data source.",
		)
		return
	}

	serviceAccountResponse, err := d.client.GetServiceAccount(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Costory service account",
			err.Error(),
		)
		return
	}

	var state serviceAccountDataSourceModel
	var diags diag.Diagnostics

	state.ServiceAccount = types.StringValue(serviceAccountResponse.ServiceAccount)
	state.SubIDs, diags = types.ListValueFrom(ctx, types.StringType, serviceAccountResponse.SubIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
