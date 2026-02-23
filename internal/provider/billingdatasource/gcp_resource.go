package billingdatasource

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/costory-io/costory-terraform/internal/costoryapi"
)

var (
	_ resource.Resource                = &gcpResource{}
	_ resource.ResourceWithConfigure   = &gcpResource{}
	_ resource.ResourceWithImportState = &gcpResource{}
)

type gcpResource struct {
	client *costoryapi.Client
}

type gcpResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	BQTablePath       types.String `tfsdk:"bq_table_path"`
	IsDetailedBilling types.Bool   `tfsdk:"is_detailed_billing"`
	StartDate         types.String `tfsdk:"start_date"`
	EndDate           types.String `tfsdk:"end_date"`
}

// NewGCPResource returns the GCP billing datasource resource.
func NewGCPResource() resource.Resource {
	return &gcpResource{}
}

func (r *gcpResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_billing_datasource_gcp", req.ProviderTypeName)
}

func (r *gcpResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a Costory GCP billing datasource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Billing datasource ID returned by Costory.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Billing datasource display name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bq_table_path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "BigQuery table path used for billing export.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_detailed_billing": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether Costory should use detailed billing rows.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"start_date": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional filter start date (YYYY-MM-DD).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"end_date": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional filter end date (YYYY-MM-DD).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *gcpResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*costoryapi.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource configure type",
			fmt.Sprintf("Expected *costoryapi.Client, got: %T. This is always a provider implementation bug.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *gcpResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var plan gcpResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := plan.toRequestModel()

	if err := r.client.ValidateGCPBillingDatasource(ctx, createRequest); err != nil {
		resp.Diagnostics.AddError(
			"Unable to validate GCP billing datasource",
			err.Error(),
		)
		return
	}

	created, err := r.client.CreateGCPBillingDatasource(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create GCP billing datasource",
			err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.mergeAPIResponse(created)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gcpResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var state gcpResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetGCPBillingDatasource(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, costoryapi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Unable to read GCP billing datasource",
			err.Error(),
		)
		return
	}

	state.mergeAPIResponse(current)
	if state.ID.IsNull() || state.ID.IsUnknown() {
		state.ID = types.StringValue(current.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gcpResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"All attributes are immutable for costory_billing_datasource_gcp. Terraform should replace the resource instead.",
	)
}

func (r *gcpResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var state gcpResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteBillingDatasource(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, costoryapi.ErrNotFound) {
		resp.Diagnostics.AddError(
			"Unable to delete GCP billing datasource",
			err.Error(),
		)
		return
	}
}

func (r *gcpResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m gcpResourceModel) toRequestModel() costoryapi.GCPBillingDatasourceRequest {
	req := costoryapi.GCPBillingDatasourceRequest{
		Name:        m.Name.ValueString(),
		BQTablePath: m.BQTablePath.ValueString(),
	}

	if !m.IsDetailedBilling.IsNull() && !m.IsDetailedBilling.IsUnknown() {
		value := m.IsDetailedBilling.ValueBool()
		req.IsDetailedBilling = &value
	}

	if !m.StartDate.IsNull() && !m.StartDate.IsUnknown() {
		value := m.StartDate.ValueString()
		req.StartDate = &value
	}

	if !m.EndDate.IsNull() && !m.EndDate.IsUnknown() {
		value := m.EndDate.ValueString()
		req.EndDate = &value
	}

	return req
}

func (m *gcpResourceModel) mergeAPIResponse(apiResponse *costoryapi.GCPBillingDatasource) {
	if apiResponse == nil {
		return
	}

	if apiResponse.ID != "" {
		m.ID = types.StringValue(apiResponse.ID)
	}

	if apiResponse.Name != "" {
		m.Name = types.StringValue(apiResponse.Name)
	}

	if apiResponse.BQTablePath != "" {
		m.BQTablePath = types.StringValue(apiResponse.BQTablePath)
	}

	if apiResponse.IsDetailedBilling != nil {
		m.IsDetailedBilling = types.BoolValue(*apiResponse.IsDetailedBilling)
	}

	if apiResponse.StartDate != nil {
		m.StartDate = types.StringValue(*apiResponse.StartDate)
	}

	if apiResponse.EndDate != nil {
		m.EndDate = types.StringValue(*apiResponse.EndDate)
	}
}
