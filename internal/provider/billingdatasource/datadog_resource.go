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
	_ resource.Resource                = &datadogResource{}
	_ resource.ResourceWithConfigure   = &datadogResource{}
	_ resource.ResourceWithImportState = &datadogResource{}
)

type datadogResource struct {
	client *costoryapi.Client
}

type datadogResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Status         types.String `tfsdk:"status"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	APIKey         types.String `tfsdk:"api_key"`
	ApplicationKey types.String `tfsdk:"application_key"`
	Region         types.String `tfsdk:"region"`
	IntegrationID  types.String `tfsdk:"integration_id"`
	BQTableURI     types.String `tfsdk:"bq_table_uri"`
}

// NewDatadogResource returns the Datadog billing datasource resource.
func NewDatadogResource() resource.Resource {
	return &datadogResource{}
}

func (r *datadogResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_billing_datasource_datadog", req.ProviderTypeName)
}

func (r *datadogResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a Costory Datadog billing datasource. See the full documentation [here](https://docs.costory.io/setup/billing#datadog).",
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
				Default:             stringdefault.StaticString("Datadog"),
				MarkdownDescription: "Datasource type. Always `Datadog` for this resource.",
			},
			"api_key": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Datadog API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"application_key": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Datadog application key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Datadog site hostname. One of `datadoghq.com`, `us3.datadoghq.com`, `us5.datadoghq.com`, `datadoghq.eu`, `ddog-gov.com`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"integration_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Datadog integration identifier returned by Costory.",
			},
			"bq_table_uri": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "BigQuery table URI created by Costory for billing data.",
			},
		},
	}
}

func (r *datadogResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureCostoryClient(req, resp)
}

func (r *datadogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var plan datadogResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := plan.toRequestModel()
	if err := r.client.ValidateDatadogBillingDatasource(ctx, createRequest); err != nil {
		resp.Diagnostics.AddError("Unable to validate Datadog billing datasource", err.Error())
		return
	}

	created, err := r.client.CreateDatadogBillingDatasource(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Datadog billing datasource", err.Error())
		return
	}

	plan.mergeAPIResponse(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *datadogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var state datadogResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetDatadogBillingDatasource(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, costoryapi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read Datadog billing datasource", err.Error())
		return
	}

	state.mergeAPIResponse(current)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *datadogResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateNotSupported(resp, "costory_billing_datasource_datadog")
}

func (r *datadogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state datadogResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteBillingDatasource(ctx, r.client, &resp.Diagnostics, state.ID.ValueString(), "Unable to delete Datadog billing datasource")
}

func (r *datadogResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m datadogResourceModel) toRequestModel() costoryapi.DatadogBillingDatasourceRequest {
	return costoryapi.DatadogBillingDatasourceRequest{
		Name:           m.Name.ValueString(),
		APIKey:         m.APIKey.ValueString(),
		ApplicationKey: m.ApplicationKey.ValueString(),
		Region:         m.Region.ValueString(),
	}
}

func (m *datadogResourceModel) mergeAPIResponse(apiResponse *costoryapi.DatadogBillingDatasource) {
	if apiResponse == nil {
		return
	}

	applyString(&m.ID, apiResponse.ID)
	applyStatus(&m.Status, apiResponse.Status)
	applyString(&m.Name, apiResponse.Name)
	applyString(&m.Type, apiResponse.Type)
	applyString(&m.IntegrationID, apiResponse.IntegrationID)
	applyString(&m.BQTableURI, apiResponse.BQTableURI)
}
