package provider

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const defaultBaseURL = "https://app.costory.io"

var (
	_ provider.Provider = &costoryProvider{}
)

type costoryProvider struct {
	version string
}

type costoryProviderModel struct {
	Slug    types.String `tfsdk:"slug"`
	Token   types.String `tfsdk:"token"`
	BaseURL types.String `tfsdk:"base_url"`
}

// New returns a constructor for the Costory Terraform provider implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &costoryProvider{
			version: version,
		}
	}
}

func (p *costoryProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "costory"
	resp.Version = p.version
}

func (p *costoryProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Costory provider forwards API calls to the Costory app.",
		Attributes: map[string]schema.Attribute{
			"slug": schema.StringAttribute{
				MarkdownDescription: "Costory tenant slug.",
				Required:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Costory API token.",
				Required:            true,
				Sensitive:           true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Costory API base URL. Defaults to `https://app.costory.io`.",
				Optional:            true,
			},
		},
	}
}

func (p *costoryProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config costoryProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Slug.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("slug"),
			"Unknown Costory slug",
			"The provider cannot create the Costory client because the slug is unknown.",
		)
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Costory token",
			"The provider cannot create the Costory client because the token is unknown.",
		)
	}

	if config.BaseURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Unknown Costory base URL",
			"The provider cannot create the Costory client because the base URL is unknown.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	slug := strings.TrimSpace(config.Slug.ValueString())
	token := strings.TrimSpace(config.Token.ValueString())
	baseURL := strings.TrimSpace(config.BaseURL.ValueString())

	if slug == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("slug"),
			"Invalid Costory slug",
			"The provider cannot create the Costory client because the slug is empty.",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Invalid Costory token",
			"The provider cannot create the Costory client because the token is empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	client := NewClient(baseURL, slug, token, &http.Client{
		Timeout: 30 * time.Second,
	})

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *costoryProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewContextDataSource,
	}
}

func (p *costoryProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}
