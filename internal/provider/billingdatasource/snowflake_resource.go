package billingdatasource

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/costory-io/costory-terraform/internal/costoryapi"
)

var (
	_ resource.Resource                = &snowflakeResource{}
	_ resource.ResourceWithConfigure   = &snowflakeResource{}
	_ resource.ResourceWithImportState = &snowflakeResource{}
)

type snowflakeResource struct {
	client *costoryapi.Client
}

type snowflakeResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Status        types.String `tfsdk:"status"`
	Name          types.String `tfsdk:"name"`
	Type          types.String `tfsdk:"type"`
	IntegrationID types.String `tfsdk:"integration_id"`
	BQTableURIs   types.List   `tfsdk:"bq_table_uris"`
}

// NewSnowflakeResource returns the Snowflake billing datasource resource.
func NewSnowflakeResource() resource.Resource {
	return &snowflakeResource{}
}

func (r *snowflakeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_billing_datasource_snowflake", req.ProviderTypeName)
}

func (r *snowflakeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a Costory Snowflake billing datasource. See the full documentation [here](https://docs.costory.io/setup/billing#snowflake).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Billing datasource ID returned by Costory.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Datasource status returned by Costory.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Billing datasource display name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Computed:            true,
				Default:             stringdefault.StaticString("SNOWFLAKE"),
				MarkdownDescription: "Datasource type. Always `SNOWFLAKE` for this resource.",
			},
			"integration_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Existing Costory Snowflake integration ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bq_table_uris": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				Sensitive:           true,
				MarkdownDescription: "BigQuery table URIs created by Costory for billing data.",
			},
		},
	}
}

func (r *snowflakeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureCostoryClient(req, resp)
}

func (r *snowflakeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var plan snowflakeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := plan.toRequestModel()
	if err := r.client.ValidateSnowflakeBillingDatasource(ctx, createRequest); err != nil {
		resp.Diagnostics.AddError("Unable to validate Snowflake billing datasource", err.Error())
		return
	}

	created, err := r.client.CreateSnowflakeBillingDatasource(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Snowflake billing datasource", err.Error())
		return
	}

	plan.mergeAPIResponse(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *snowflakeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var state snowflakeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetSnowflakeBillingDatasource(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, costoryapi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read Snowflake billing datasource", err.Error())
		return
	}

	state.mergeAPIResponse(current)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *snowflakeResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateNotSupported(resp, "costory_billing_datasource_snowflake")
}

func (r *snowflakeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state snowflakeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteBillingDatasource(ctx, r.client, &resp.Diagnostics, state.ID.ValueString(), "Unable to delete Snowflake billing datasource")
}

func (r *snowflakeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m snowflakeResourceModel) toRequestModel() costoryapi.SnowflakeBillingDatasourceRequest {
	return costoryapi.SnowflakeBillingDatasourceRequest{
		Name:          m.Name.ValueString(),
		IntegrationID: m.IntegrationID.ValueString(),
	}
}

func (m *snowflakeResourceModel) mergeAPIResponse(apiResponse *costoryapi.SnowflakeBillingDatasource) {
	if apiResponse == nil {
		return
	}

	applyString(&m.ID, apiResponse.ID)
	applyStatus(&m.Status, apiResponse.Status)
	applyString(&m.Name, apiResponse.Name)
	applyString(&m.Type, apiResponse.Type)
	applyString(&m.IntegrationID, apiResponse.IntegrationID)

	values := make([]attr.Value, 0, len(apiResponse.BQTableURIs))
	for _, uri := range apiResponse.BQTableURIs {
		values = append(values, types.StringValue(uri))
	}
	m.BQTableURIs, _ = types.ListValue(types.StringType, values)
}
