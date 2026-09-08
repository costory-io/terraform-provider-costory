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
	_ resource.Resource                = &anthropicClaudeAiResource{}
	_ resource.ResourceWithConfigure   = &anthropicClaudeAiResource{}
	_ resource.ResourceWithImportState = &anthropicClaudeAiResource{}
)

type anthropicClaudeAiResource struct {
	client *costoryapi.Client
}

type anthropicClaudeAiResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Status          types.String `tfsdk:"status"`
	Name            types.String `tfsdk:"name"`
	Type            types.String `tfsdk:"type"`
	AnalyticsAPIKey types.String `tfsdk:"analytics_api_key"`
	AccountName     types.String `tfsdk:"account_name"`
	BQTableURI      types.String `tfsdk:"bq_table_uri"`
	StartDate       types.String `tfsdk:"start_date"`
}

// NewAnthropicClaudeAiResource returns the Anthropic Claude AI billing datasource resource.
func NewAnthropicClaudeAiResource() resource.Resource {
	return &anthropicClaudeAiResource{}
}

func (r *anthropicClaudeAiResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_billing_datasource_anthropic_claude_ai", req.ProviderTypeName)
}

func (r *anthropicClaudeAiResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a Costory Anthropic Claude AI billing datasource. See the full documentation [here](https://docs.costory.io/setup/billing#anthropic-claude-ai).",
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
				Default:             stringdefault.StaticString("AnthropicClaudeAi"),
				MarkdownDescription: "Datasource type. Always `AnthropicClaudeAi` for this resource.",
			},
			"analytics_api_key": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Anthropic Claude AI analytics API key used to fetch billing data.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"account_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Anthropic Claude AI account name.",
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

func (r *anthropicClaudeAiResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureCostoryClient(req, resp)
}

func (r *anthropicClaudeAiResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var plan anthropicClaudeAiResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := plan.toRequestModel()
	if err := r.client.ValidateAnthropicClaudeAiBillingDatasource(ctx, createRequest); err != nil {
		resp.Diagnostics.AddError("Unable to validate Anthropic Claude AI billing datasource", err.Error())
		return
	}

	created, err := r.client.CreateAnthropicClaudeAiBillingDatasource(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Anthropic Claude AI billing datasource", err.Error())
		return
	}

	plan.mergeAPIResponse(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *anthropicClaudeAiResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var state anthropicClaudeAiResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetAnthropicClaudeAiBillingDatasource(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, costoryapi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read Anthropic Claude AI billing datasource", err.Error())
		return
	}

	state.mergeAPIResponse(current)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *anthropicClaudeAiResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateNotSupported(resp, "costory_billing_datasource_anthropic_claude_ai")
}

func (r *anthropicClaudeAiResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state anthropicClaudeAiResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteBillingDatasource(ctx, r.client, &resp.Diagnostics, state.ID.ValueString(), "Unable to delete Anthropic Claude AI billing datasource")
}

func (r *anthropicClaudeAiResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m anthropicClaudeAiResourceModel) toRequestModel() costoryapi.AnthropicClaudeAiBillingDatasourceRequest {
	return costoryapi.AnthropicClaudeAiBillingDatasourceRequest{
		Name:            m.Name.ValueString(),
		AnalyticsAPIKey: m.AnalyticsAPIKey.ValueString(),
		AccountName:     m.AccountName.ValueString(),
		StartDate:       optionalString(m.StartDate),
	}
}

func (m *anthropicClaudeAiResourceModel) mergeAPIResponse(apiResponse *costoryapi.AnthropicClaudeAiBillingDatasource) {
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
