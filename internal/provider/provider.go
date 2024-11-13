package provider

import (
	"context"

	"github.com/Wa4h1h/terraform-provider-tools/internal/archive"
	"github.com/Wa4h1h/terraform-provider-tools/internal/httpclient"
	"github.com/Wa4h1h/terraform-provider-tools/internal/random"
	"github.com/Wa4h1h/terraform-provider-tools/internal/template"
	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// make sure we conform to Provider
var _ provider.Provider = &ToolsProvider{}

type Tools struct {
	archive.Archiver
	httpclient.HTTPRunner
	random.Randomizer
	template.TemplateRunner
}

func NewTools() *Tools {
	return &Tools{
		archive.NewArchiver(),
		httpclient.NewHTTPRunner(),
		random.NewRandomizer(),
		template.NewTemplateRunner(),
	}
}

type ToolsProviderModel struct {
	http types.Object `tfsdk:"http"`
}

type ToolsProvider struct {
	version string
}

func NewToolsProvider(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ToolsProvider{
			version: version,
		}
	}
}

func (t *ToolsProvider) Metadata(ctx context.Context,
	req provider.MetadataRequest,
	resp *provider.MetadataResponse,
) {
	resp.TypeName = "tool"
	resp.Version = t.version
}

func (t *ToolsProvider) Schema(_ context.Context,
	_ provider.SchemaRequest,
	resp *provider.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"http": schema.ObjectAttribute{
				Optional:  true,
				Sensitive: true,
				AttributeTypes: map[string]attr.Type{
					"access_token_url": types.StringType,
					"client_id":        types.StringType,
					"client_secret":    types.StringType,
					"scope":            types.StringType,
					"grant_type":       types.StringType,
				},
			},
		},
	}
}

func (t *ToolsProvider) Configure(ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	// load provider config
	var data ToolsProviderModel

	d := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(d...)

	// assert if http config object value is known by the time the provider is being initialized
	if data.http.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("http"),
			"Unknown http config",
			"The provider cannot create http client",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// init API
	tools := NewTools()

	resp.DataSourceData = tools
	resp.ResourceData = tools
}

func (t *ToolsProvider) DataSources(ctx context.Context,
) []func() datasource.DataSource {
	return nil
}

func (t *ToolsProvider) Resources(ctx context.Context,
) []func() resource.Resource {
	return nil
}
