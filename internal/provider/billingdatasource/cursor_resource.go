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
	_ resource.Resource                = &cursorResource{}
	_ resource.ResourceWithConfigure   = &cursorResource{}
	_ resource.ResourceWithImportState = &cursorResource{}
)

type cursorResource struct {
	client *costoryapi.Client
}

type cursorResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Status      types.String `tfsdk:"status"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	AdminAPIKey types.String `tfsdk:"admin_api_key"`
	BQTableURI  types.String `tfsdk:"bq_table_uri"`
	StartDate   types.String `tfsdk:"start_date"`
	EndDate     types.String `tfsdk:"end_date"`
}

// NewCursorResource returns the Cursor billing datasource resource.
func NewCursorResource() resource.Resource {
	return &cursorResource{}
}

func (r *cursorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_billing_datasource_cursor", req.ProviderTypeName)
}

func (r *cursorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a Costory Cursor billing datasource. See the full documentation [here](https://docs.costory.io/setup/billing#cursor).",
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
				Default:             stringdefault.StaticString("Cursor"),
				MarkdownDescription: "Datasource type. Always `Cursor` for this resource.",
			},
			"admin_api_key": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Cursor admin API key used to fetch billing data.",
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
			"end_date": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional filter end date (ISO-8601).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *cursorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *cursorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var plan cursorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := plan.toRequestModel()

	if err := r.client.ValidateCursorBillingDatasource(ctx, createRequest); err != nil {
		resp.Diagnostics.AddError(
			"Unable to validate Cursor billing datasource",
			err.Error(),
		)
		return
	}

	created, err := r.client.CreateCursorBillingDatasource(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Cursor billing datasource",
			err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.mergeAPIResponse(created)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cursorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var state cursorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetCursorBillingDatasource(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, costoryapi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Unable to read Cursor billing datasource",
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

func (r *cursorResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"All attributes are immutable for costory_billing_datasource_cursor. Terraform should replace the resource instead.",
	)
}

func (r *cursorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var state cursorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteBillingDatasource(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, costoryapi.ErrNotFound) {
		resp.Diagnostics.AddError(
			"Unable to delete Cursor billing datasource",
			err.Error(),
		)
		return
	}
}

func (r *cursorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m cursorResourceModel) toRequestModel() costoryapi.CursorBillingDatasourceRequest {
	req := costoryapi.CursorBillingDatasourceRequest{
		Name:        m.Name.ValueString(),
		AdminAPIKey: m.AdminAPIKey.ValueString(),
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

func (m *cursorResourceModel) mergeAPIResponse(apiResponse *costoryapi.CursorBillingDatasource) {
	if apiResponse == nil {
		return
	}

	if apiResponse.ID != "" {
		m.ID = types.StringValue(apiResponse.ID)
	}

	m.Status = types.StringNull()
	if apiResponse.Status != nil {
		m.Status = types.StringValue(*apiResponse.Status)
	}

	if apiResponse.Name != "" {
		m.Name = types.StringValue(apiResponse.Name)
	}

	if apiResponse.Type != "" {
		m.Type = types.StringValue(apiResponse.Type)
	}

	if apiResponse.BQTableURI != "" {
		m.BQTableURI = types.StringValue(apiResponse.BQTableURI)
	}

	if apiResponse.StartDate != nil {
		m.StartDate = types.StringValue(*apiResponse.StartDate)
	}

	if apiResponse.EndDate != nil {
		m.EndDate = types.StringValue(*apiResponse.EndDate)
	}
}
