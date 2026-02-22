package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &contextDataSource{}
	_ datasource.DataSourceWithConfigure = &contextDataSource{}
)

type contextDataSource struct {
	client *Client
}

type contextDataSourceModel struct {
	ServiceAccount types.String `tfsdk:"service_account"`
	SubIDs         types.List   `tfsdk:"sub_ids"`
}

// NewContextDataSource returns the Costory context data source.
func NewContextDataSource() datasource.DataSource {
	return &contextDataSource{}
}

func (d *contextDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_context", req.ProviderTypeName)
}

func (d *contextDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Returns Costory service-account context from the Costory API.",
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

func (d *contextDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected data source configure type",
			fmt.Sprintf("Expected *provider.Client, got: %T. This is always a provider implementation bug.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *contextDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the data source.",
		)
		return
	}

	contextResponse, err := d.client.GetContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Costory context",
			err.Error(),
		)
		return
	}

	var state contextDataSourceModel
	var diags diag.Diagnostics

	state.ServiceAccount = types.StringValue(contextResponse.ServiceAccount)
	state.SubIDs, diags = types.ListValueFrom(ctx, types.StringType, contextResponse.SubIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
