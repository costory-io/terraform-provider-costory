package team

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/costory-io/costory-terraform/internal/costoryapi"
)

var (
	_ resource.Resource                = &teamMemberResource{}
	_ resource.ResourceWithConfigure   = &teamMemberResource{}
	_ resource.ResourceWithImportState = &teamMemberResource{}
)

type teamMemberResource struct {
	client *costoryapi.Client
}

type teamMemberResourceModel struct {
	ID     types.String `tfsdk:"id"`
	TeamID types.String `tfsdk:"team_id"`
	UserID types.String `tfsdk:"user_id"`
	Email  types.String `tfsdk:"email"`
	Role   types.String `tfsdk:"role"`
}

// NewMemberResource returns the team member resource.
func NewMemberResource() resource.Resource {
	return &teamMemberResource{}
}

func (r *teamMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_team_member", req.ProviderTypeName)
}

func (r *teamMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a single Costory team member.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Team member resource ID.",
			},
			"team_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Team ID to manage membership for.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Existing Costory user ID to add.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Email address to invite to the team.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Role for the member. One of: `OWNER`, `MEMBER`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *teamMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *teamMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var plan teamMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !validateTeamMemberIdentity(resp, plan.UserID, plan.Email) {
		return
	}

	teamID := plan.TeamID.ValueString()
	memberKey := ""
	request := costoryapi.TeamMemberRequest{}
	if !plan.UserID.IsNull() && !plan.UserID.IsUnknown() {
		userID := plan.UserID.ValueString()
		request.UserID = &userID
		memberKey = userID
	} else {
		email := plan.Email.ValueString()
		request.Email = &email
		memberKey = email
	}

	if value := stringValueFromAttr(plan.Role); value != nil {
		request.Role = value
	}

	if err := r.client.AddTeamMember(ctx, teamID, request); err != nil {
		resp.Diagnostics.AddError(
			"Unable to add team member",
			err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(formatTeamMemberID(teamID, memberKey))
	if plan.Role.IsNull() || plan.Role.IsUnknown() {
		plan.Role = types.StringValue("MEMBER")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teamMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *teamMemberResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"All attributes are immutable for costory_team_member. Terraform should replace the resource instead.",
	)
}

func (r *teamMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var state teamMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.UserID.IsNull() || state.UserID.IsUnknown() {
		resp.Diagnostics.AddWarning(
			"Unable to remove team member without user_id",
			"The Costory API requires a user_id to remove a team member. The resource will be removed from state, but the invitation may remain until manually removed.",
		)
		return
	}

	err := r.client.RemoveTeamMember(ctx, state.TeamID.ValueString(), state.UserID.ValueString())
	if err != nil && !errors.Is(err, costoryapi.ErrNotFound) {
		resp.Diagnostics.AddError(
			"Unable to remove team member",
			err.Error(),
		)
		return
	}
}

func (r *teamMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	teamID, userID, err := parseTeamMemberImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import identifier",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_id"), types.StringValue(teamID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), types.StringValue(userID))...)
}

func validateTeamMemberIdentity(resp *resource.CreateResponse, userID types.String, email types.String) bool {
	hasUserID := !userID.IsNull() && !userID.IsUnknown()
	hasEmail := !email.IsNull() && !email.IsUnknown()

	if hasUserID == hasEmail {
		resp.Diagnostics.AddError(
			"Invalid team member identity",
			"Exactly one of user_id or email must be set.",
		)
		return false
	}

	return true
}

func formatTeamMemberID(teamID, memberKey string) string {
	return fmt.Sprintf("%s:%s", teamID, memberKey)
}

func parseTeamMemberImportID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("expected import identifier in the form {team_id}:{user_id}")
	}
	return parts[0], parts[1], nil
}
