package billingdatasource

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/costory-io/costory-terraform/internal/costoryapi"
)

var (
	_ resource.Resource                = &customBigQueryResource{}
	_ resource.ResourceWithConfigure   = &customBigQueryResource{}
	_ resource.ResourceWithImportState = &customBigQueryResource{}
)

type customBigQueryResource struct {
	client *costoryapi.Client
}

type customBigQueryMappingModel struct {
	BilledCost                 types.String `tfsdk:"billed_cost"`
	SubAccountID               types.String `tfsdk:"sub_account_id"`
	ServiceName                types.String `tfsdk:"service_name"`
	BillingPeriodStart         types.String `tfsdk:"billing_period_start"`
	BillingPeriodEnd           types.String `tfsdk:"billing_period_end"`
	ChargePeriodStart          types.String `tfsdk:"charge_period_start"`
	ChargePeriodEnd            types.String `tfsdk:"charge_period_end"`
	ChargeDescription          types.String `tfsdk:"charge_description"`
	BillingCurrency            types.String `tfsdk:"billing_currency"`
	ContractedCost             types.String `tfsdk:"contracted_cost"`
	BillingAccountID           types.String `tfsdk:"billing_account_id"`
	OriginalBillingCurrency    types.String `tfsdk:"original_billing_currency"`
	ChargeCategory             types.String `tfsdk:"charge_category"`
	ChargeFrequency            types.String `tfsdk:"charge_frequency"`
	CommitmentDiscountCategory types.String `tfsdk:"commitment_discount_category"`
	CommitmentDiscountID       types.String `tfsdk:"commitment_discount_id"`
	CommitmentDiscountType     types.String `tfsdk:"commitment_discount_type"`
	ConsumedQuantity           types.String `tfsdk:"consumed_quantity"`
	ConsumedUnit               types.String `tfsdk:"consumed_unit"`
	EffectiveCost              types.String `tfsdk:"effective_cost"`
	InvoiceIssuer              types.String `tfsdk:"invoice_issuer"`
	ListCost                   types.String `tfsdk:"list_cost"`
	ListUnitPrice              types.String `tfsdk:"list_unit_price"`
	PricingCategory            types.String `tfsdk:"pricing_category"`
	PricingQuantity            types.String `tfsdk:"pricing_quantity"`
	PricingUnit                types.String `tfsdk:"pricing_unit"`
	Provider                   types.String `tfsdk:"provider"`
	Publisher                  types.String `tfsdk:"publisher"`
	RegionID                   types.String `tfsdk:"region_id"`
	ResourceID                 types.String `tfsdk:"resource_id"`
	ResourceName               types.String `tfsdk:"resource_name"`
	ResourceType               types.String `tfsdk:"resource_type"`
	CategoryFocus              types.String `tfsdk:"category_focus"`
	SKUID                      types.String `tfsdk:"sku_id"`
	SKU                        types.String `tfsdk:"sku"`
	SKUPriceID                 types.String `tfsdk:"sku_price_id"`
	Tags                       types.String `tfsdk:"tags"`
}

type customBigQueryResourceModel struct {
	ID               types.String                `tfsdk:"id"`
	Status           types.String                `tfsdk:"status"`
	Name             types.String                `tfsdk:"name"`
	Type             types.String                `tfsdk:"type"`
	BQTablePath      types.String                `tfsdk:"bq_table_path"`
	ProviderName     types.String                `tfsdk:"provider_name"`
	BillingAccountID types.String                `tfsdk:"billing_account_id"`
	StartDate        types.String                `tfsdk:"start_date"`
	EndDate          types.String                `tfsdk:"end_date"`
	Mapping          *customBigQueryMappingModel `tfsdk:"mapping"`
}

// NewCustomBigQueryResource returns the custom BigQuery billing datasource resource.
func NewCustomBigQueryResource() resource.Resource {
	return &customBigQueryResource{}
}

func (r *customBigQueryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_billing_datasource_custom_bigquery", req.ProviderTypeName)
}

func optionalMappingAttribute(markdownDescription string) schema.StringAttribute {
	return schema.StringAttribute{
		Optional:            true,
		MarkdownDescription: markdownDescription,
	}
}

func (r *customBigQueryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Creates a Costory custom BigQuery billing datasource. See the full documentation [here](https://docs.costory.io/setup/billing#custom-bigquery).",
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
				Default:             stringdefault.StaticString("CustomBigQuery"),
				MarkdownDescription: "Datasource type. Always `CustomBigQuery` for this resource.",
			},
			"bq_table_path": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "BigQuery table path (`project_id.dataset_id.table_name`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"provider_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Provider name stored with the imported billing rows.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"billing_account_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional billing account identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
			"mapping": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "Column mapping from the BigQuery table to Costory billing fields. `billed_cost` is required; other fields are optional BigQuery column names.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"billed_cost": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "BigQuery column mapped to billed cost.",
					},
					"sub_account_id":               optionalMappingAttribute("BigQuery column mapped to sub-account ID."),
					"service_name":                 optionalMappingAttribute("BigQuery column mapped to service name."),
					"billing_period_start":         optionalMappingAttribute("BigQuery column mapped to billing period start."),
					"billing_period_end":           optionalMappingAttribute("BigQuery column mapped to billing period end."),
					"charge_period_start":          optionalMappingAttribute("BigQuery column mapped to charge period start."),
					"charge_period_end":            optionalMappingAttribute("BigQuery column mapped to charge period end."),
					"charge_description":           optionalMappingAttribute("BigQuery column mapped to charge description."),
					"billing_currency":             optionalMappingAttribute("BigQuery column mapped to billing currency."),
					"contracted_cost":              optionalMappingAttribute("BigQuery column mapped to contracted cost."),
					"billing_account_id":           optionalMappingAttribute("BigQuery column mapped to billing account ID."),
					"original_billing_currency":    optionalMappingAttribute("BigQuery column mapped to original billing currency."),
					"charge_category":              optionalMappingAttribute("BigQuery column mapped to charge category."),
					"charge_frequency":             optionalMappingAttribute("BigQuery column mapped to charge frequency."),
					"commitment_discount_category": optionalMappingAttribute("BigQuery column mapped to commitment discount category."),
					"commitment_discount_id":       optionalMappingAttribute("BigQuery column mapped to commitment discount ID."),
					"commitment_discount_type":     optionalMappingAttribute("BigQuery column mapped to commitment discount type."),
					"consumed_quantity":            optionalMappingAttribute("BigQuery column mapped to consumed quantity."),
					"consumed_unit":                optionalMappingAttribute("BigQuery column mapped to consumed unit."),
					"effective_cost":               optionalMappingAttribute("BigQuery column mapped to effective cost."),
					"invoice_issuer":               optionalMappingAttribute("BigQuery column mapped to invoice issuer."),
					"list_cost":                    optionalMappingAttribute("BigQuery column mapped to list cost."),
					"list_unit_price":              optionalMappingAttribute("BigQuery column mapped to list unit price."),
					"pricing_category":             optionalMappingAttribute("BigQuery column mapped to pricing category."),
					"pricing_quantity":             optionalMappingAttribute("BigQuery column mapped to pricing quantity."),
					"pricing_unit":                 optionalMappingAttribute("BigQuery column mapped to pricing unit."),
					"provider":                     optionalMappingAttribute("BigQuery column mapped to provider."),
					"publisher":                    optionalMappingAttribute("BigQuery column mapped to publisher."),
					"region_id":                    optionalMappingAttribute("BigQuery column mapped to region ID."),
					"resource_id":                  optionalMappingAttribute("BigQuery column mapped to resource ID."),
					"resource_name":                optionalMappingAttribute("BigQuery column mapped to resource name."),
					"resource_type":                optionalMappingAttribute("BigQuery column mapped to resource type."),
					"category_focus":               optionalMappingAttribute("BigQuery column mapped to category focus."),
					"sku_id":                       optionalMappingAttribute("BigQuery column mapped to SKU ID."),
					"sku":                          optionalMappingAttribute("BigQuery column mapped to SKU."),
					"sku_price_id":                 optionalMappingAttribute("BigQuery column mapped to SKU price ID."),
					"tags":                         optionalMappingAttribute("BigQuery column mapped to tags."),
				},
			},
		},
	}
}

func (r *customBigQueryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = configureCostoryClient(req, resp)
}

func (r *customBigQueryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var plan customBigQueryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := plan.toRequestModel()
	if err := r.client.ValidateCustomBigQueryBillingDatasource(ctx, createRequest); err != nil {
		resp.Diagnostics.AddError("Unable to validate custom BigQuery billing datasource", err.Error())
		return
	}

	created, err := r.client.CreateCustomBigQueryBillingDatasource(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create custom BigQuery billing datasource", err.Error())
		return
	}

	plan.mergeAPIResponse(created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *customBigQueryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		addUnconfiguredClientError(&resp.Diagnostics)
		return
	}

	var state customBigQueryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetCustomBigQueryBillingDatasource(ctx, state.ID.ValueString())
	if err != nil {
		if errors.Is(err, costoryapi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read custom BigQuery billing datasource", err.Error())
		return
	}

	state.mergeAPIResponse(current)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *customBigQueryResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	updateNotSupported(resp, "costory_billing_datasource_custom_bigquery")
}

func (r *customBigQueryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customBigQueryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteBillingDatasource(ctx, r.client, &resp.Diagnostics, state.ID.ValueString(), "Unable to delete custom BigQuery billing datasource")
}

func (r *customBigQueryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (m customBigQueryResourceModel) toRequestModel() costoryapi.CustomBigQueryBillingDatasourceRequest {
	return costoryapi.CustomBigQueryBillingDatasourceRequest{
		Name:             m.Name.ValueString(),
		BQTablePath:      m.BQTablePath.ValueString(),
		ProviderName:     m.ProviderName.ValueString(),
		BillingAccountID: optionalString(m.BillingAccountID),
		StartDate:        optionalString(m.StartDate),
		EndDate:          optionalString(m.EndDate),
		Mapping:          mappingToAPI(m.Mapping),
	}
}

func (m *customBigQueryResourceModel) mergeAPIResponse(apiResponse *costoryapi.CustomBigQueryBillingDatasource) {
	if apiResponse == nil {
		return
	}

	applyString(&m.ID, apiResponse.ID)
	applyStatus(&m.Status, apiResponse.Status)
	applyString(&m.Name, apiResponse.Name)
	applyString(&m.Type, apiResponse.Type)
	applyString(&m.BQTablePath, apiResponse.BQTablePath)
	applyString(&m.ProviderName, apiResponse.ProviderName)
	applyOptionalString(&m.BillingAccountID, apiResponse.BillingAccountID)
	applyOptionalString(&m.StartDate, apiResponse.StartDate)
	applyOptionalString(&m.EndDate, apiResponse.EndDate)
	if len(apiResponse.Mapping) > 0 {
		m.Mapping = mappingFromAPI(apiResponse.Mapping)
	}
}

func mappingToAPI(mapping *customBigQueryMappingModel) map[string]string {
	if mapping == nil {
		return map[string]string{}
	}

	out := map[string]string{
		"billedCost": mapping.BilledCost.ValueString(),
	}
	setMappingValue(out, "subAccountId", mapping.SubAccountID)
	setMappingValue(out, "serviceName", mapping.ServiceName)
	setMappingValue(out, "billingPeriodStart", mapping.BillingPeriodStart)
	setMappingValue(out, "billingPeriodEnd", mapping.BillingPeriodEnd)
	setMappingValue(out, "chargePeriodStart", mapping.ChargePeriodStart)
	setMappingValue(out, "chargePeriodEnd", mapping.ChargePeriodEnd)
	setMappingValue(out, "chargeDescription", mapping.ChargeDescription)
	setMappingValue(out, "billingCurrency", mapping.BillingCurrency)
	setMappingValue(out, "contractedCost", mapping.ContractedCost)
	setMappingValue(out, "billingAccountId", mapping.BillingAccountID)
	setMappingValue(out, "originalBillingCurrency", mapping.OriginalBillingCurrency)
	setMappingValue(out, "chargeCategory", mapping.ChargeCategory)
	setMappingValue(out, "chargeFrequency", mapping.ChargeFrequency)
	setMappingValue(out, "commitmentDiscountCategory", mapping.CommitmentDiscountCategory)
	setMappingValue(out, "commitmentDiscountId", mapping.CommitmentDiscountID)
	setMappingValue(out, "commitmentDiscountType", mapping.CommitmentDiscountType)
	setMappingValue(out, "consumedQuantity", mapping.ConsumedQuantity)
	setMappingValue(out, "consumedUnit", mapping.ConsumedUnit)
	setMappingValue(out, "effectiveCost", mapping.EffectiveCost)
	setMappingValue(out, "invoiceIssuer", mapping.InvoiceIssuer)
	setMappingValue(out, "listCost", mapping.ListCost)
	setMappingValue(out, "listUnitPrice", mapping.ListUnitPrice)
	setMappingValue(out, "pricingCategory", mapping.PricingCategory)
	setMappingValue(out, "pricingQuantity", mapping.PricingQuantity)
	setMappingValue(out, "pricingUnit", mapping.PricingUnit)
	setMappingValue(out, "provider", mapping.Provider)
	setMappingValue(out, "publisher", mapping.Publisher)
	setMappingValue(out, "regionId", mapping.RegionID)
	setMappingValue(out, "resourceId", mapping.ResourceID)
	setMappingValue(out, "resourceName", mapping.ResourceName)
	setMappingValue(out, "resourceType", mapping.ResourceType)
	setMappingValue(out, "categoryFocus", mapping.CategoryFocus)
	setMappingValue(out, "skuId", mapping.SKUID)
	setMappingValue(out, "sku", mapping.SKU)
	setMappingValue(out, "skuPriceId", mapping.SKUPriceID)
	setMappingValue(out, "tags", mapping.Tags)
	return out
}

func setMappingValue(out map[string]string, key string, value types.String) {
	if value.IsNull() || value.IsUnknown() {
		return
	}
	if raw := value.ValueString(); raw != "" {
		out[key] = raw
	}
}

func mappingFromAPI(mapping map[string]string) *customBigQueryMappingModel {
	get := func(key string) types.String {
		if value, ok := mapping[key]; ok && value != "" {
			return types.StringValue(value)
		}
		return types.StringNull()
	}

	billedCost := get("billedCost")
	if billedCost.IsNull() {
		billedCost = types.StringValue("")
	}

	return &customBigQueryMappingModel{
		BilledCost:                 billedCost,
		SubAccountID:               get("subAccountId"),
		ServiceName:                get("serviceName"),
		BillingPeriodStart:         get("billingPeriodStart"),
		BillingPeriodEnd:           get("billingPeriodEnd"),
		ChargePeriodStart:          get("chargePeriodStart"),
		ChargePeriodEnd:            get("chargePeriodEnd"),
		ChargeDescription:          get("chargeDescription"),
		BillingCurrency:            get("billingCurrency"),
		ContractedCost:             get("contractedCost"),
		BillingAccountID:           get("billingAccountId"),
		OriginalBillingCurrency:    get("originalBillingCurrency"),
		ChargeCategory:             get("chargeCategory"),
		ChargeFrequency:            get("chargeFrequency"),
		CommitmentDiscountCategory: get("commitmentDiscountCategory"),
		CommitmentDiscountID:       get("commitmentDiscountId"),
		CommitmentDiscountType:     get("commitmentDiscountType"),
		ConsumedQuantity:           get("consumedQuantity"),
		ConsumedUnit:               get("consumedUnit"),
		EffectiveCost:              get("effectiveCost"),
		InvoiceIssuer:              get("invoiceIssuer"),
		ListCost:                   get("listCost"),
		ListUnitPrice:              get("listUnitPrice"),
		PricingCategory:            get("pricingCategory"),
		PricingQuantity:            get("pricingQuantity"),
		PricingUnit:                get("pricingUnit"),
		Provider:                   get("provider"),
		Publisher:                  get("publisher"),
		RegionID:                   get("regionId"),
		ResourceID:                 get("resourceId"),
		ResourceName:               get("resourceName"),
		ResourceType:               get("resourceType"),
		CategoryFocus:              get("categoryFocus"),
		SKUID:                      get("skuId"),
		SKU:                        get("sku"),
		SKUPriceID:                 get("skuPriceId"),
		Tags:                       get("tags"),
	}
}
