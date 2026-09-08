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
	_ resource.Resource                = &gcpCUDExportResource{}
	_ resource.ResourceWithConfigure   = &gcpCUDExportResource{}
	_ resource.ResourceWithImportState = &gcpCUDExportResource{}
)

type gcpCUDExportResource struct {
	client *costoryapi.Client
}

type gcpCUDExportResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Status      types.String `tfsdk:"status"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	BQTablePath types.String `tfsdk:"bq_table_path"`
}

// NewGCPCUDExportResource returns the GCP CUD export billing datasource resource.
func NewGCPCUDExportResource() resource.Resource {
	return &gcpCUDExportResource{}
}

func (r *gcpCUDExportResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_billing_datasource_gcp_cud_export", req.ProviderTypeName)
}

func (r *gcpCUDExportResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a Costory GCP committed-use discount export billing datasource. See the full documentation [here](https://docs.costory.io/setup/billing/gcp-cud-metadata).",
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
				Default:             stringdefault.StaticString("GCP_CUD_EXPORT"),
				MarkdownDescription: "Datasource type. Always `GCP_CUD_EXPORT` for this resource.",
			},
			"bq_table_path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "BigQuery table path for the CUD export (`project_id.dataset_id.table_name`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *gcpCUDExportResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureCostoryClient(req, resp)
}

func (r *gcpCUDExportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var plan gcpCUDExportResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := plan.toRequestModel()
	if err := r.client.ValidateGCPCUDExportBillingDatasource(ctx, createRequest); err != nil {
		resp.Diagnostics.AddError("Unable to validate GCP CUD export billing datasource", err.Error())
		return
	}

	created, err := r.client.CreateGCPCUDExportBillingDatasource(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create GCP CUD export billing datasource", err.Error())
		return
	}

	plan.mergeAPIResponse(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gcpCUDExportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var state gcpCUDExportResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetGCPCUDExportBillingDatasource(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, costoryapi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read GCP CUD export billing datasource", err.Error())
		return
	}

	state.mergeAPIResponse(current)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gcpCUDExportResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateNotSupported(resp, "costory_billing_datasource_gcp_cud_export")
}

func (r *gcpCUDExportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state gcpCUDExportResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteBillingDatasource(ctx, r.client, &resp.Diagnostics, state.ID.ValueString(), "Unable to delete GCP CUD export billing datasource")
}

func (r *gcpCUDExportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m gcpCUDExportResourceModel) toRequestModel() costoryapi.GCPCUDExportBillingDatasourceRequest {
	return costoryapi.GCPCUDExportBillingDatasourceRequest{
		Name:        m.Name.ValueString(),
		BQTablePath: m.BQTablePath.ValueString(),
	}
}

func (m *gcpCUDExportResourceModel) mergeAPIResponse(apiResponse *costoryapi.GCPCUDExportBillingDatasource) {
	if apiResponse == nil {
		return
	}

	applyString(&m.ID, apiResponse.ID)
	applyStatus(&m.Status, apiResponse.Status)
	applyString(&m.Name, apiResponse.Name)
	applyString(&m.Type, apiResponse.Type)
	applyString(&m.BQTablePath, apiResponse.BQTablePath)
}
