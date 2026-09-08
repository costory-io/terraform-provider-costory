package billingdatasource

import (
	"context"
	"errors"
	"fmt"

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
	_ resource.Resource                = &clickHouseCloudResource{}
	_ resource.ResourceWithConfigure   = &clickHouseCloudResource{}
	_ resource.ResourceWithImportState = &clickHouseCloudResource{}
)

type clickHouseCloudResource struct {
	client *costoryapi.Client
}

type clickHouseCloudResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Status         types.String `tfsdk:"status"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	KeyID          types.String `tfsdk:"key_id"`
	KeySecret      types.String `tfsdk:"key_secret"`
	OrganizationID types.String `tfsdk:"organization_id"`
	BQTableURI     types.String `tfsdk:"bq_table_uri"`
	StartDate      types.String `tfsdk:"start_date"`
}

// NewClickHouseCloudResource returns the ClickHouse Cloud billing datasource resource.
func NewClickHouseCloudResource() resource.Resource {
	return &clickHouseCloudResource{}
}

func (r *clickHouseCloudResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_billing_datasource_clickhouse_cloud", req.ProviderTypeName)
}

func (r *clickHouseCloudResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a Costory ClickHouse Cloud billing datasource. See the full documentation [here](https://docs.costory.io/setup/billing#clickhouse-cloud).",
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
				Default:             stringdefault.StaticString("ClickHouseCloud"),
				MarkdownDescription: "Datasource type. Always `ClickHouseCloud` for this resource.",
			},
			"key_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ClickHouse Cloud API key ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_secret": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "ClickHouse Cloud API key secret.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ClickHouse Cloud organization ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bq_table_uri": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "BigQuery table URI created by Costory for billing data.",
			},
			"start_date": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional filter start date (ISO-8601).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *clickHouseCloudResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureCostoryClient(req, resp)
}

func (r *clickHouseCloudResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var plan clickHouseCloudResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := plan.toRequestModel()
	if err := r.client.ValidateClickHouseCloudBillingDatasource(ctx, createRequest); err != nil {
		resp.Diagnostics.AddError("Unable to validate ClickHouse Cloud billing datasource", err.Error())
		return
	}

	created, err := r.client.CreateClickHouseCloudBillingDatasource(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create ClickHouse Cloud billing datasource", err.Error())
		return
	}

	plan.mergeAPIResponse(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clickHouseCloudResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var state clickHouseCloudResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetClickHouseCloudBillingDatasource(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, costoryapi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read ClickHouse Cloud billing datasource", err.Error())
		return
	}

	state.mergeAPIResponse(current)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *clickHouseCloudResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateNotSupported(resp, "costory_billing_datasource_clickhouse_cloud")
}

func (r *clickHouseCloudResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state clickHouseCloudResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteBillingDatasource(ctx, r.client, &resp.Diagnostics, state.ID.ValueString(), "Unable to delete ClickHouse Cloud billing datasource")
}

func (r *clickHouseCloudResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m clickHouseCloudResourceModel) toRequestModel() costoryapi.ClickHouseCloudBillingDatasourceRequest {
	return costoryapi.ClickHouseCloudBillingDatasourceRequest{
		Name:           m.Name.ValueString(),
		KeyID:          m.KeyID.ValueString(),
		KeySecret:      m.KeySecret.ValueString(),
		OrganizationID: m.OrganizationID.ValueString(),
		StartDate:      optionalString(m.StartDate),
	}
}

func (m *clickHouseCloudResourceModel) mergeAPIResponse(apiResponse *costoryapi.ClickHouseCloudBillingDatasource) {
	if apiResponse == nil {
		return
	}

	applyString(&m.ID, apiResponse.ID)
	applyStatus(&m.Status, apiResponse.Status)
	applyString(&m.Name, apiResponse.Name)
	applyString(&m.Type, apiResponse.Type)
	applyString(&m.BQTableURI, apiResponse.BQTableURI)
	applyOptionalString(&m.StartDate, apiResponse.StartDate)
}
