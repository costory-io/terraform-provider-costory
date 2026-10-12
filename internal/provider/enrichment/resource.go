package enrichment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gopkg.in/yaml.v3"

	"github.com/costory-io/costory-terraform/internal/costoryapi"
)

var (
	_ resource.Resource                = &enrichmentResource{}
	_ resource.ResourceWithConfigure   = &enrichmentResource{}
	_ resource.ResourceWithImportState = &enrichmentResource{}
)

type enrichmentResource struct {
	client *costoryapi.Client
}

type enrichmentResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	File          types.String `tfsdk:"file"`
	ContentSHA256 types.String `tfsdk:"content_sha256"`
}

// NewResource returns the enrichment resource.
func NewResource() resource.Resource {
	return &enrichmentResource{}
}

func (r *enrichmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_enrichment", req.ProviderTypeName)
}

func (r *enrichmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages one Costory enrichment from a YAML file. The file is the customer document, with `conditions` written in CEL and `targetColumns` as a map from column to `{ mode }`, in column order. Applying stores the document as a draft; `costory_enrichments_publish` makes the drafts live in one transaction. Destroying drafts the removal.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Enrichment name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Enrichment name. Must match `name` in the YAML file.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"file": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Path to the enrichment YAML file.",
			},
			"content_sha256": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "SHA-256 of the YAML file, empty when the live document differs from the file. Pass it to `costory_enrichments_publish`.",
				PlanModifiers: []planmodifier.String{
					fileHashModifier{},
				},
			},
		},
	}
}

func (r *enrichmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *enrichmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Costory client", "The provider did not configure the Costory API client for the resource.")
		return
	}
	var plan enrichmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, plan.Name.ValueString(), plan.File.ValueString()); err != nil {
		addApplyError(&resp.Diagnostics, err)
		return
	}
	plan.ID = plan.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *enrichmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Costory client", "The provider did not configure the Costory API client for the resource.")
		return
	}
	var state enrichmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, err := r.client.GetEnrichment(ctx, state.Name.ValueString())
	if err != nil {
		if errors.Is(err, costoryapi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read enrichment", err.Error())
		return
	}

	if !state.File.IsNull() && !state.File.IsUnknown() && state.File.ValueString() != "" {
		document, readErr := readDocument(state.File.ValueString())
		if readErr != nil {
			resp.Diagnostics.AddError("Unable to read enrichment file", readErr.Error())
			return
		}
		hash, hashErr := hashFile(state.File.ValueString())
		if hashErr != nil {
			resp.Diagnostics.AddError("Unable to read enrichment file", hashErr.Error())
			return
		}
		if documentsEqual(document, current) {
			state.ContentSHA256 = types.StringValue(hash)
		} else {
			state.ContentSHA256 = types.StringValue("")
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *enrichmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Costory client", "The provider did not configure the Costory API client for the resource.")
		return
	}
	var plan enrichmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, plan.Name.ValueString(), plan.File.ValueString()); err != nil {
		addApplyError(&resp.Diagnostics, err)
		return
	}
	plan.ID = plan.Name
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *enrichmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Unconfigured Costory client", "The provider did not configure the Costory API client for the resource.")
		return
	}
	var state enrichmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteEnrichment(ctx, state.Name.ValueString()); err != nil && !errors.Is(err, costoryapi.ErrNotFound) {
		addApplyError(&resp.Diagnostics, err)
	}
}

func (r *enrichmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}

func (r *enrichmentResource) apply(ctx context.Context, name, file string) error {
	document, err := readDocument(file)
	if err != nil {
		return err
	}
	documentName, _ := document["name"].(string)
	if documentName != name {
		return fmt.Errorf("name in %s is %q, resource name is %q", file, documentName, name)
	}
	return r.client.PutEnrichment(ctx, name, document)
}

type fileHashModifier struct{}

func (m fileHashModifier) Description(context.Context) string {
	return "Hashes the enrichment YAML file."
}

func (m fileHashModifier) MarkdownDescription(context.Context) string {
	return "Hashes the enrichment YAML file."
}

func (m fileHashModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	var file types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("file"), &file)...)
	if resp.Diagnostics.HasError() || file.IsNull() || file.IsUnknown() || file.ValueString() == "" {
		return
	}
	hash, err := hashFile(file.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("file"), "Unable to read enrichment file", err.Error())
		return
	}
	resp.PlanValue = types.StringValue(hash)
}

// readDocument reads a YAML file into the API form of the document.
func readDocument(file string) (map[string]any, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return parseDocument(file, raw)
}

func parseDocument(file string, raw []byte) (map[string]any, error) {
	var node yaml.Node
	if err := yaml.Unmarshal(raw, &node); err != nil {
		return nil, fmt.Errorf("parse %s: %w", file, err)
	}
	var document map[string]any
	if err := node.Decode(&document); err != nil {
		return nil, fmt.Errorf("parse %s: %w", file, err)
	}
	if document == nil {
		return nil, fmt.Errorf("%s must contain one object", file)
	}
	columns, err := targetColumnList(&node)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", file, err)
	}
	if columns != nil {
		document["targetColumns"] = columns
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		return nil, err
	}
	var normalized map[string]any
	if err := json.Unmarshal(encoded, &normalized); err != nil {
		return nil, err
	}
	return normalized, nil
}

// targetColumnList turns `targetColumns: { column: { mode } }` into the API
// list `[{ name, mode }]`, in file order: the compiler compares column order.
// It returns nil when the file already holds a list.
func targetColumnList(node *yaml.Node) ([]any, error) {
	if node.Kind != yaml.DocumentNode || len(node.Content) == 0 {
		return nil, nil
	}
	root := node.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, nil
	}
	for index := 0; index+1 < len(root.Content); index += 2 {
		if root.Content[index].Value != "targetColumns" {
			continue
		}
		value := root.Content[index+1]
		if value.Kind != yaml.MappingNode {
			return nil, nil
		}
		columns := []any{}
		for at := 0; at+1 < len(value.Content); at += 2 {
			column := map[string]any{}
			if err := value.Content[at+1].Decode(&column); err != nil {
				return nil, fmt.Errorf("targetColumns.%s: %w", value.Content[at].Value, err)
			}
			column["name"] = value.Content[at].Value
			columns = append(columns, column)
		}
		return columns, nil
	}
	return nil, nil
}

func hashFile(file string) (string, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func documentsEqual(fileDocument, stored map[string]any) bool {
	return reflect.DeepEqual(comparableDocument(fileDocument), comparableDocument(stored))
}

// comparableDocument drops what the API fills in: an absent description reads as ""
// and a null condition is left out.
func comparableDocument(document map[string]any) any {
	normalized, _ := normalizeDocument(document).(map[string]any)
	if normalized["description"] == "" {
		delete(normalized, "description")
	}
	return normalized
}

func normalizeDocument(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		normalized := map[string]any{}
		for key, item := range typed {
			if item == nil {
				continue
			}
			normalized[key] = normalizeDocument(item)
		}
		return normalized
	case []any:
		normalized := make([]any, len(typed))
		for index, item := range typed {
			normalized[index] = normalizeDocument(item)
		}
		return normalized
	default:
		return value
	}
}

func addApplyError(diagnostics *diag.Diagnostics, err error) {
	var statusErr *costoryapi.StatusError
	if errors.As(err, &statusErr) && len(statusErr.Errors) > 0 {
		title := "Invalid enrichment"
		if statusErr.Status == http.StatusConflict {
			title = "Enrichment conflict"
		}
		for _, message := range statusErr.Errors {
			diagnostics.AddError(title, message)
		}
		return
	}
	diagnostics.AddError("Unable to save enrichment", err.Error())
}
