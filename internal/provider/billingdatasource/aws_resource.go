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
	_ resource.Resource                = &awsResource{}
	_ resource.ResourceWithConfigure   = &awsResource{}
	_ resource.ResourceWithImportState = &awsResource{}
)

type awsResource struct {
	client *costoryapi.Client
}

type awsResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Status              types.String `tfsdk:"status"`
	Name                types.String `tfsdk:"name"`
	BucketName          types.String `tfsdk:"bucket_name"`
	RoleARN             types.String `tfsdk:"role_arn"`
	Prefix              types.String `tfsdk:"prefix"`
	EKSSplitDataEnabled types.Bool   `tfsdk:"eks_split_data_enabled"`
	StartDate           types.String `tfsdk:"start_date"`
	EndDate             types.String `tfsdk:"end_date"`
	EKSSplit            types.Bool   `tfsdk:"eks_split"`
}

// NewAWSResource returns the AWS billing datasource resource.
func NewAWSResource() resource.Resource {
	return &awsResource{}
}

func (r *awsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_billing_datasource_aws", req.ProviderTypeName)
}

func (r *awsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a Costory AWS billing datasource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Billing datasource ID returned by Costory.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Datasource status returned by Costory (for example ACTIVE or PENDING).",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Billing datasource display name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bucket_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "S3 bucket containing AWS billing exports.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_arn": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IAM role ARN used by Costory to access AWS billing exports.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prefix": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Object prefix path inside the billing export bucket.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"eks_split_data_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether EKS split data is enabled in ingestion.",
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
			"eks_split": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Optional EKS split mode flag used by the API.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *awsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *awsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var plan awsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := plan.toRequestModel()

	created, err := r.client.CreateAWSBillingDatasource(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create AWS billing datasource",
			err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.mergeAPIResponse(created)

	// Refresh after create so state reflects observed backend status (for example PENDING -> ACTIVE lifecycle).
	current, err := r.client.GetAWSBillingDatasource(ctx, created.ID)
	if err != nil {
		if errors.Is(err, costoryapi.ErrNotFound) {
			resp.Diagnostics.AddWarning(
				"Created datasource not yet readable",
				"Costory accepted datasource creation, but the datasource was not immediately readable. The current create response was stored in state and the next refresh will reconcile observed status.",
			)
		} else {
			resp.Diagnostics.AddWarning(
				"Unable to refresh datasource after create",
				err.Error(),
			)
		}
	} else {
		plan.mergeAPIResponse(current)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *awsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var state awsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetAWSBillingDatasource(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, costoryapi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Unable to read AWS billing datasource",
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

func (r *awsResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"All attributes are immutable for costory_billing_datasource_aws. Terraform should replace the resource instead.",
	)
}

func (r *awsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var state awsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteBillingDatasource(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, costoryapi.ErrNotFound) {
		resp.Diagnostics.AddError(
			"Unable to delete AWS billing datasource",
			err.Error(),
		)
		return
	}
}

func (r *awsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m awsResourceModel) toRequestModel() costoryapi.AWSBillingDatasourceRequest {
	req := costoryapi.AWSBillingDatasourceRequest{
		Name:       m.Name.ValueString(),
		BucketName: m.BucketName.ValueString(),
		RoleARN:    m.RoleARN.ValueString(),
		Prefix:     m.Prefix.ValueString(),
	}

	if !m.EKSSplitDataEnabled.IsNull() && !m.EKSSplitDataEnabled.IsUnknown() {
		value := m.EKSSplitDataEnabled.ValueBool()
		req.EKSSplitDataEnabled = &value
	}

	if !m.StartDate.IsNull() && !m.StartDate.IsUnknown() {
		value := m.StartDate.ValueString()
		req.StartDate = &value
	}

	if !m.EndDate.IsNull() && !m.EndDate.IsUnknown() {
		value := m.EndDate.ValueString()
		req.EndDate = &value
	}

	if !m.EKSSplit.IsNull() && !m.EKSSplit.IsUnknown() {
		value := m.EKSSplit.ValueBool()
		req.EKSSplit = &value
	}

	return req
}

func (m *awsResourceModel) mergeAPIResponse(apiResponse *costoryapi.AWSBillingDatasource) {
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

	if apiResponse.BucketName != "" {
		m.BucketName = types.StringValue(apiResponse.BucketName)
	}

	if apiResponse.RoleARN != "" {
		m.RoleARN = types.StringValue(apiResponse.RoleARN)
	}

	if apiResponse.Prefix != "" {
		m.Prefix = types.StringValue(apiResponse.Prefix)
	}

	if apiResponse.EKSSplitDataEnabled != nil {
		m.EKSSplitDataEnabled = types.BoolValue(*apiResponse.EKSSplitDataEnabled)
	}

	if apiResponse.StartDate != nil {
		m.StartDate = types.StringValue(*apiResponse.StartDate)
	}

	if apiResponse.EndDate != nil {
		m.EndDate = types.StringValue(*apiResponse.EndDate)
	}

	if apiResponse.EKSSplit != nil {
		m.EKSSplit = types.BoolValue(*apiResponse.EKSSplit)
	}
}
