package billingdatasource

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
	_ resource.Resource                = &azureResource{}
	_ resource.ResourceWithConfigure   = &azureResource{}
	_ resource.ResourceWithImportState = &azureResource{}
)

type azureResource struct {
	client *costoryapi.Client
}

type azureResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Status             types.String `tfsdk:"status"`
	Name               types.String `tfsdk:"name"`
	SASURL             types.String `tfsdk:"sas_url"`
	StorageAccountName types.String `tfsdk:"storage_account_name"`
	ContainerName      types.String `tfsdk:"container_name"`
	ActualsPath        types.String `tfsdk:"actuals_path"`
	AmortizedPath      types.String `tfsdk:"amortized_path"`
}

// NewAzureResource returns the Azure billing datasource resource.
func NewAzureResource() resource.Resource {
	return &azureResource{}
}

func (r *azureResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_billing_datasource_azure", req.ProviderTypeName)
}

func (r *azureResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a Costory Azure billing datasource.",
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
			"sas_url": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Full Azure blob SAS URL including the query string.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"storage_account_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Azure storage account name hosting the export container.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"container_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Azure storage container name with billing exports.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"actuals_path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Path prefix for actual cost exports in the container.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"amortized_path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Path prefix for amortized cost exports in the container.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *azureResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *azureResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var plan azureResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := plan.toRequestModel()

	if err := r.client.ValidateAzureBillingDatasource(ctx, createRequest); err != nil {
		resp.Diagnostics.AddError(
			"Unable to validate Azure billing datasource",
			err.Error(),
		)
		return
	}

	created, err := r.client.CreateAzureBillingDatasource(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Azure billing datasource",
			err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.mergeAPIResponse(created)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *azureResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var state azureResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetAzureBillingDatasource(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, costoryapi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Unable to read Azure billing datasource",
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

func (r *azureResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"All attributes are immutable for costory_billing_datasource_azure. Terraform should replace the resource instead.",
	)
}

func (r *azureResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var state azureResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteBillingDatasource(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, costoryapi.ErrNotFound) {
		resp.Diagnostics.AddError(
			"Unable to delete Azure billing datasource",
			err.Error(),
		)
		return
	}
}

func (r *azureResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m azureResourceModel) toRequestModel() costoryapi.AzureBillingDatasourceRequest {
	return costoryapi.AzureBillingDatasourceRequest{
		Name:               m.Name.ValueString(),
		SASURL:             m.SASURL.ValueString(),
		StorageAccountName: m.StorageAccountName.ValueString(),
		ContainerName:      m.ContainerName.ValueString(),
		ActualsPath:        m.ActualsPath.ValueString(),
		AmortizedPath:      m.AmortizedPath.ValueString(),
	}
}

func (m *azureResourceModel) mergeAPIResponse(apiResponse *costoryapi.AzureBillingDatasource) {
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

	if apiResponse.StorageAccountName != "" {
		m.StorageAccountName = types.StringValue(apiResponse.StorageAccountName)
	}

	if apiResponse.ContainerName != "" {
		m.ContainerName = types.StringValue(apiResponse.ContainerName)
	}

	if apiResponse.ActualsPath != "" {
		m.ActualsPath = types.StringValue(apiResponse.ActualsPath)
	}

	if apiResponse.AmortizedPath != "" {
		m.AmortizedPath = types.StringValue(apiResponse.AmortizedPath)
	}
}
