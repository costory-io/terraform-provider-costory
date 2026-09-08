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
	_ resource.Resource                = &cloudflareResource{}
	_ resource.ResourceWithConfigure   = &cloudflareResource{}
	_ resource.ResourceWithImportState = &cloudflareResource{}
)

type cloudflareResource struct {
	client *costoryapi.Client
}

type cloudflareResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Status     types.String `tfsdk:"status"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	APIToken   types.String `tfsdk:"api_token"`
	AccountID  types.String `tfsdk:"account_id"`
	BQTableURI types.String `tfsdk:"bq_table_uri"`
	StartDate  types.String `tfsdk:"start_date"`
}

// NewCloudflareResource returns the Cloudflare billing datasource resource.
func NewCloudflareResource() resource.Resource {
	return &cloudflareResource{}
}

func (r *cloudflareResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_billing_datasource_cloudflare", req.ProviderTypeName)
}

func (r *cloudflareResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a Costory Cloudflare billing datasource. See the full documentation [here](https://docs.costory.io/setup/billing#cloudflare).",
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
				Default:             stringdefault.StaticString("Cloudflare"),
				MarkdownDescription: "Datasource type. Always `Cloudflare` for this resource.",
			},
			"api_token": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Cloudflare API token used to fetch billing data.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"account_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Cloudflare account ID.",
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

func (r *cloudflareResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureCostoryClient(req, resp)
}

func (r *cloudflareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var plan cloudflareResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := plan.toRequestModel()
	if err := r.client.ValidateCloudflareBillingDatasource(ctx, createRequest); err != nil {
		resp.Diagnostics.AddError("Unable to validate Cloudflare billing datasource", err.Error())
		return
	}

	created, err := r.client.CreateCloudflareBillingDatasource(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Cloudflare billing datasource", err.Error())
		return
	}

	plan.mergeAPIResponse(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cloudflareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var state cloudflareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetCloudflareBillingDatasource(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, costoryapi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read Cloudflare billing datasource", err.Error())
		return
	}

	state.mergeAPIResponse(current)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cloudflareResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateNotSupported(resp, "costory_billing_datasource_cloudflare")
}

func (r *cloudflareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cloudflareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteBillingDatasource(ctx, r.client, &resp.Diagnostics, state.ID.ValueString(), "Unable to delete Cloudflare billing datasource")
}

func (r *cloudflareResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m cloudflareResourceModel) toRequestModel() costoryapi.CloudflareBillingDatasourceRequest {
	return costoryapi.CloudflareBillingDatasourceRequest{
		Name:      m.Name.ValueString(),
		APIToken:  m.APIToken.ValueString(),
		AccountID: m.AccountID.ValueString(),
		StartDate: optionalString(m.StartDate),
	}
}

func (m *cloudflareResourceModel) mergeAPIResponse(apiResponse *costoryapi.CloudflareBillingDatasource) {
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
