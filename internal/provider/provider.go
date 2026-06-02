package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.ProviderWithFunctions = &OpenAIProvider{}

type OpenAIProvider struct{ version string }

type OpenAIProviderModel struct {
	APIKey  types.String `tfsdk:"api_key"`
	OrgID   types.String `tfsdk:"organization_id"`
	BaseURL types.String `tfsdk:"base_url"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider { return &OpenAIProvider{version: version} }
}

func (p *OpenAIProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "openai"
	resp.Version = p.version
}

func (p *OpenAIProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage OpenAI platform resources — projects, API keys, and service accounts — via the OpenAI Platform API.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "OpenAI admin API key. Can be set via `OPENAI_API_KEY` env var.",
				Optional:            true,
				Sensitive:           true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "OpenAI organization ID (optional). Can be set via `OPENAI_ORG_ID` env var.",
				Optional:            true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Override the OpenAI API base URL. Defaults to `https://api.openai.com`.",
				Optional:            true,
			},
		},
	}
}

func (p *OpenAIProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg OpenAIProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if !cfg.APIKey.IsNull() {
		apiKey = cfg.APIKey.ValueString()
	}
	if apiKey == "" {
		resp.Diagnostics.AddError("Missing API key", "Set api_key in the provider block or OPENAI_API_KEY env var.")
		return
	}

	orgID := os.Getenv("OPENAI_ORG_ID")
	if !cfg.OrgID.IsNull() {
		orgID = cfg.OrgID.ValueString()
	}

	baseURL := ""
	if !cfg.BaseURL.IsNull() {
		baseURL = cfg.BaseURL.ValueString()
	}

	client := newClient(apiKey, orgID, baseURL)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *OpenAIProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewAPIKeyResource,
		NewServiceAccountResource,
	}
}

func (p *OpenAIProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *OpenAIProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{}
}
