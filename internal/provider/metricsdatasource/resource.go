package metricsdatasource

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	_ resource.Resource                = &metricsDatasourceResource{}
	_ resource.ResourceWithConfigure   = &metricsDatasourceResource{}
	_ resource.ResourceWithImportState = &metricsDatasourceResource{}
)

type metricsDatasourceResource struct {
	client *costoryapi.Client
}

type metricsDatasourceResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Status            types.String `tfsdk:"status"`
	Name              types.String `tfsdk:"name"`
	Type              types.String `tfsdk:"type"`
	BucketName        types.String `tfsdk:"bucket_name"`
	Prefix            types.String `tfsdk:"prefix"`
	RoleARN           types.String `tfsdk:"role_arn"`
	MetricsDefinition types.List   `tfsdk:"metrics_definition"`
}

var metricsDefinitionAttrTypes = map[string]attr.Type{
	"metric_name":  types.StringType,
	"gap_filling":  types.StringType,
	"aggregation":  types.StringType,
	"value_column": types.StringType,
	"date_column":  types.StringType,
	"dimensions":   types.ListType{ElemType: types.StringType},
	"unit":         types.StringType,
}

// NewResource returns the metrics datasource resource.
func NewResource() resource.Resource {
	return &metricsDatasourceResource{}
}

func (r *metricsDatasourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_metrics_datasource_s3_parquet", req.ProviderTypeName)
}

func (r *metricsDatasourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a Costory AwsS3V2 metrics datasource for parquet files in S3.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Datasource ID returned by Costory.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Datasource status returned by Costory.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name of the datasource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Datasource type. Always `AwsS3V2` for this resource.",
				Default:             stringdefault.StaticString("AwsS3V2"),
			},
			"bucket_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "S3 bucket name containing parquet files.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prefix": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "S3 key prefix for parquet files.",
				Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_arn": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IAM role ARN (must match `arn:aws:iam::`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"metrics_definition": schema.ListNestedAttribute{
				Required:            true,
				MarkdownDescription: "List of metric definitions.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"metric_name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Pretty name of the metric.",
						},
						"gap_filling": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "One of: `ZERO`, `FORWARD_FILL`, `LINEAR_INTERPOLATION`, `SPREAD`.",
						},
						"aggregation": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "One of: `SUM`, `Average`.",
						},
						"value_column": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Parquet column containing metric values.",
						},
						"date_column": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Parquet column containing dates.",
						},
						"dimensions": schema.ListAttribute{
							Optional:            true,
							MarkdownDescription: "Parquet columns used as dimensional axes.",
							ElementType:         types.StringType,
						},
						"unit": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Unit label for the metric.",
						},
					},
				},
			},
		},
	}
}

func (r *metricsDatasourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *metricsDatasourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var plan metricsDatasourceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := plan.toRequestModel()

	if err := r.client.ValidateMetricsDatasource(ctx, createRequest); err != nil {
		resp.Diagnostics.AddError(
			"Unable to validate metrics datasource",
			err.Error(),
		)
		return
	}

	created, err := r.client.CreateMetricsDatasource(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create metrics datasource",
			err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(created.ID)
	plan.mergeAPIResponse(created)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *metricsDatasourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var state metricsDatasourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetMetricsDatasource(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, costoryapi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Unable to read metrics datasource",
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

func (r *metricsDatasourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var plan metricsDatasourceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state metricsDatasourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := plan.toRequestModel()

	if err := r.client.ValidateMetricsDatasource(ctx, createRequest); err != nil {
		resp.Diagnostics.AddError(
			"Unable to validate metrics datasource",
			err.Error(),
		)
		return
	}

	defs := createRequest.MetricsDefinitions
	if err := r.client.UpdateMetricsDatasource(ctx, state.ID.ValueString(), defs); err != nil {
		resp.Diagnostics.AddError(
			"Unable to update metrics datasource",
			err.Error(),
		)
		return
	}

	current, err := r.client.GetMetricsDatasource(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Unable to refresh metrics datasource after update",
			err.Error(),
		)
		plan.ID = state.ID
		plan.Status = state.Status
	} else {
		plan.mergeAPIResponse(current)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *metricsDatasourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Costory client",
			"The provider did not configure the Costory API client for the resource.",
		)
		return
	}

	var state metricsDatasourceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMetricsDatasource(ctx, state.ID.ValueString())
	if err != nil && !errors.Is(err, costoryapi.ErrNotFound) {
		resp.Diagnostics.AddError(
			"Unable to delete metrics datasource",
			err.Error(),
		)
		return
	}
}

func (r *metricsDatasourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m *metricsDatasourceResourceModel) toRequestModel() costoryapi.MetricsDatasourceRequest {
	defs := make([]costoryapi.MetricsDefinition, 0, len(m.MetricsDefinition.Elements()))
	for _, elem := range m.MetricsDefinition.Elements() {
		obj, ok := elem.(types.Object)
		if !ok {
			continue
		}
		attrs := obj.Attributes()
		md := costoryapi.MetricsDefinition{
			MetricName:  getAttrString(attrs, "metric_name"),
			GapFilling:  getAttrString(attrs, "gap_filling"),
			Aggregation: getAttrString(attrs, "aggregation"),
			ValueColumn: getAttrString(attrs, "value_column"),
			DateColumn:  getAttrString(attrs, "date_column"),
			Unit:        getAttrString(attrs, "unit"),
		}
		if dims, ok := attrs["dimensions"]; ok && !dims.IsNull() && !dims.IsUnknown() {
			if list, ok := dims.(types.List); ok {
				for _, d := range list.Elements() {
					if s, ok := d.(types.String); ok {
						md.Dimensions = append(md.Dimensions, s.ValueString())
					}
				}
			}
		}
		defs = append(defs, md)
	}

	prefix := ""
	if !m.Prefix.IsNull() && !m.Prefix.IsUnknown() {
		prefix = m.Prefix.ValueString()
	}

	return costoryapi.MetricsDatasourceRequest{
		Name:               m.Name.ValueString(),
		Type:               m.Type.ValueString(),
		BucketName:         m.BucketName.ValueString(),
		Prefix:             prefix,
		RoleARN:            m.RoleARN.ValueString(),
		MetricsDefinitions: defs,
	}
}

func getAttrString(attrs map[string]attr.Value, key string) string {
	if v, ok := attrs[key]; ok && !v.IsNull() && !v.IsUnknown() {
		if s, ok := v.(types.String); ok {
			return s.ValueString()
		}
	}
	return ""
}

func (m *metricsDatasourceResourceModel) mergeAPIResponse(apiResponse *costoryapi.MetricsDatasource) {
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

	if apiResponse.BucketName != "" {
		m.BucketName = types.StringValue(apiResponse.BucketName)
	}

	m.Prefix = types.StringValue(apiResponse.Prefix)

	if apiResponse.RoleARN != "" {
		m.RoleARN = types.StringValue(apiResponse.RoleARN)
	}

	if len(apiResponse.MetricsDefinition) > 0 {
		objType := types.ObjectType{AttrTypes: metricsDefinitionAttrTypes}
		objs := make([]attr.Value, len(apiResponse.MetricsDefinition))
		for i, d := range apiResponse.MetricsDefinition {
			var dims types.List
			if len(d.Dimensions) > 0 {
				dims, _ = types.ListValueFrom(context.Background(), types.StringType, d.Dimensions)
			} else {
				dims = types.ListNull(types.StringType)
			}
			objs[i] = types.ObjectValueMust(metricsDefinitionAttrTypes, map[string]attr.Value{
				"metric_name":  types.StringValue(d.MetricName),
				"gap_filling":  types.StringValue(d.GapFilling),
				"aggregation":  types.StringValue(d.Aggregation),
				"value_column": types.StringValue(d.ValueColumn),
				"date_column":  types.StringValue(d.DateColumn),
				"dimensions":   dims,
				"unit":         types.StringValue(d.Unit),
			})
		}
		m.MetricsDefinition = types.ListValueMust(objType, objs)
	}
}
