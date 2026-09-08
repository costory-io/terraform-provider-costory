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
	_ resource.Resource                = &confluentResource{}
	_ resource.ResourceWithConfigure   = &confluentResource{}
	_ resource.ResourceWithImportState = &confluentResource{}
)

type confluentResource struct {
	client *costoryapi.Client
}

type confluentResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Status     types.String `tfsdk:"status"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	APIKey     types.String `tfsdk:"api_key"`
	APISecret  types.String `tfsdk:"api_secret"`
	BQTableURI types.String `tfsdk:"bq_table_uri"`
}

// NewConfluentResource returns the Confluent billing datasource resource.
func NewConfluentResource() resource.Resource {
	return &confluentResource{}
}

func (r *confluentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_billing_datasource_confluent", req.ProviderTypeName)
}

func (r *confluentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a Costory Confluent billing datasource. See the full documentation [here](https://docs.costory.io/setup/billing#confluent).",
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
				Default:             stringdefault.StaticString("CONFLUENT"),
				MarkdownDescription: "Datasource type. Always `CONFLUENT` for this resource.",
			},
			"api_key": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Confluent Cloud API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"api_secret": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Confluent Cloud API secret.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bq_table_uri": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "BigQuery table URI created by Costory for billing data.",
			},
		},
	}
}

func (r *confluentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureCostoryClient(req, resp)
}

func (r *confluentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var plan confluentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := plan.toRequestModel()
	if err := r.client.ValidateConfluentBillingDatasource(ctx, createRequest); err != nil {
		resp.Diagnostics.AddError("Unable to validate Confluent billing datasource", err.Error())
		return
	}

	created, err := r.client.CreateConfluentBillingDatasource(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Confluent billing datasource", err.Error())
		return
	}

	plan.mergeAPIResponse(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *confluentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var state confluentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetConfluentBillingDatasource(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, costoryapi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read Confluent billing datasource", err.Error())
		return
	}

	state.mergeAPIResponse(current)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *confluentResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateNotSupported(resp, "costory_billing_datasource_confluent")
}

func (r *confluentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state confluentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteBillingDatasource(ctx, r.client, &resp.Diagnostics, state.ID.ValueString(), "Unable to delete Confluent billing datasource")
}

func (r *confluentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m confluentResourceModel) toRequestModel() costoryapi.ConfluentBillingDatasourceRequest {
	return costoryapi.ConfluentBillingDatasourceRequest{
		Name:      m.Name.ValueString(),
		APIKey:    m.APIKey.ValueString(),
		APISecret: m.APISecret.ValueString(),
	}
}

func (m *confluentResourceModel) mergeAPIResponse(apiResponse *costoryapi.ConfluentBillingDatasource) {
	if apiResponse == nil {
		return
	}

	applyString(&m.ID, apiResponse.ID)
	applyStatus(&m.Status, apiResponse.Status)
	applyString(&m.Name, apiResponse.Name)
	applyString(&m.Type, apiResponse.Type)
	applyString(&m.BQTableURI, apiResponse.BQTableURI)
}
