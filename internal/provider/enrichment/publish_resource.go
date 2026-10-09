package enrichment

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/costory-io/costory-terraform/internal/costoryapi"
)

var (
	_ resource.Resource              = &publishResource{}
	_ resource.ResourceWithConfigure = &publishResource{}
)

const publishID = "enrichments"

type publishResource struct {
	client *costoryapi.Client
}

type publishResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enrichments types.Map    `tfsdk:"enrichments"`
}

// NewPublishResource returns the resource that publishes enrichment drafts.
func NewPublishResource() resource.Resource {
	return &publishResource{}
}

func (r *publishResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_enrichments_publish", req.ProviderTypeName)
}

func (r *publishResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Publishes the enrichment drafts written by `costory_enrichment`: Costory compiles the live set with every draft applied, makes it live in one transaction, and starts the billing update. " +
			"A set that does not compile is refused with the pipeline's error and nothing changes. " +
			"Use one per organization. Destroying it publishes nothing; the enrichments stay live.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Always `enrichments`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enrichments": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
				MarkdownDescription: "Enrichment name to `content_sha256`, usually `{ for e in costory_enrichment.all : e.name => e.content_sha256 }`. " +
					"Any change publishes. A name that leaves the map is removed before publishing, so its removal goes out in the same apply.",
			},
		},
	}
}

func (r *publishResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *publishResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Costory client", "The provider did not configure the Costory API client for the resource.")
		return
	}
	var plan publishResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.publish(ctx, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(publishID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read keeps the state: drift is on each costory_enrichment, whose
// content_sha256 empties when the live document differs from its file.
func (r *publishResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state publishResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *publishResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Costory client", "The provider did not configure the Costory API client for the resource.")
		return
	}
	var plan, state publishResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var before, after map[string]string
	resp.Diagnostics.Append(state.Enrichments.ElementsAs(ctx, &before, false)...)
	resp.Diagnostics.Append(plan.Enrichments.ElementsAs(ctx, &after, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.publish(ctx, removedNames(before, after), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(publishID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *publishResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {}

// publish drafts the removal of names that left the map, then publishes.
// Terraform updates this resource before it destroys the enrichments it no
// longer references, so their own Delete would come after the publish.
func (r *publishResource) publish(ctx context.Context, removed []string, diagnostics *diag.Diagnostics) {
	for _, name := range removed {
		if err := r.client.DeleteEnrichment(ctx, name); err != nil && !errors.Is(err, costoryapi.ErrNotFound) {
			diagnostics.AddError("Unable to remove enrichment "+name, err.Error())
			return
		}
	}
	result, err := r.client.PublishEnrichments(ctx)
	if err != nil {
		addPublishError(diagnostics, err)
		return
	}
	if result == nil || result.WorkflowStarted || result.WorkflowError == "" {
		return
	}
	diagnostics.AddWarning(
		"Billing update not started",
		fmt.Sprintf("The enrichments are live (%s); the next billing run applies them. %s", changeSummary(result.Changes), result.WorkflowError),
	)
}

func removedNames(before, after map[string]string) []string {
	var removed []string
	for name := range before {
		if _, ok := after[name]; !ok {
			removed = append(removed, name)
		}
	}
	sort.Strings(removed)
	return removed
}

func changeSummary(changes costoryapi.EnrichmentChanges) string {
	parts := []string{}
	for _, part := range []struct {
		sign  string
		names []string
	}{{"+", changes.Added}, {"~", changes.Changed}, {"-", changes.Removed}} {
		for _, name := range part.names {
			parts = append(parts, part.sign+name)
		}
	}
	return strings.Join(parts, " ")
}

func addPublishError(diagnostics *diag.Diagnostics, err error) {
	var statusErr *costoryapi.StatusError
	if errors.As(err, &statusErr) && len(statusErr.Errors) > 0 {
		title := "Enrichments do not compile"
		if statusErr.Status == http.StatusConflict {
			title = "Enrichment drafts changed while publishing"
		}
		for _, message := range statusErr.Errors {
			diagnostics.AddError(title, message)
		}
		return
	}
	diagnostics.AddError("Unable to publish enrichments", err.Error())
}
