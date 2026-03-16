package team

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/costory-io/costory-terraform/internal/costoryapi"
)

var (
	_ resource.Resource                = &teamResource{}
	_ resource.ResourceWithConfigure   = &teamResource{}
	_ resource.ResourceWithImportState = &teamResource{}
)

type teamResource struct {
	client *costoryapi.Client
}

type teamResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Visibility  types.String `tfsdk:"visibility"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// NewResource returns the team resource.
func NewResource() resource.Resource {
	return &teamResource{}
}

func (r *teamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_team", req.ProviderTypeName)
}

func (r *teamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Costory team.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Team ID returned by Costory.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Team display name.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Optional team description.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"visibility": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Team visibility. One of: `PRIVATE`, `PUBLIC`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Team creation timestamp.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Team last update timestamp.",
			},
		},
	}
}

func (r *teamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var plan teamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateTeam(ctx, plan.toCreateRequest())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create team",
			err.Error(),
		)
		return
	}

	plan.mergeAPIResponse(created)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var state teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetTeam(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, costoryapi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Unable to read team",
			err.Error(),
		)
		return
	}

	state.mergeAPIResponse(current)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var plan teamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.UpdateTeam(ctx, state.ID.ValueString(), plan.toUpdateRequest(state))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update team",
			err.Error(),
		)
		return
	}

	plan.mergeAPIResponse(updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var state teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteTeam(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, costoryapi.ErrNotFound) {
		resp.Diagnostics.AddError(
			"Unable to delete team",
			err.Error(),
		)
		return
	}
}

func (r *teamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m teamResourceModel) toCreateRequest() costoryapi.TeamCreateRequest {
	req := costoryapi.TeamCreateRequest{
		Name: m.Name.ValueString(),
	}

	if value := stringValueFromAttr(m.Description); value != nil {
		req.Description = value
	}

	if value := stringValueFromAttr(m.Visibility); value != nil {
		req.Visibility = value
	}

	return req
}

func (m teamResourceModel) toUpdateRequest(state teamResourceModel) costoryapi.TeamUpdateRequest {
	req := costoryapi.TeamUpdateRequest{
		Name: stringValueFromAttr(m.Name),
	}

	req.Description = stringValueFromPlanOrState(m.Description, state.Description)
	req.Visibility = stringValueFromPlanOrState(m.Visibility, state.Visibility)

	return req
}

func stringValueFromPlanOrState(plan types.String, state types.String) *string {
	if value := stringValueFromAttr(plan); value != nil {
		return value
	}

	return stringValueFromAttr(state)
}

func stringValueFromAttr(value types.String) *string {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	val := value.ValueString()
	return &val
}

func (m *teamResourceModel) mergeAPIResponse(apiResponse *costoryapi.Team) {
	if apiResponse == nil {
		return
	}

	if apiResponse.ID != "" {
		m.ID = types.StringValue(apiResponse.ID)
	}

	m.Name = types.StringValue(apiResponse.Name)
	m.Description = types.StringValue(apiResponse.Description)
	m.Visibility = types.StringValue(apiResponse.Visibility)
	m.CreatedAt = types.StringValue(apiResponse.CreatedAt)
	m.UpdatedAt = types.StringValue(apiResponse.UpdatedAt)
}
