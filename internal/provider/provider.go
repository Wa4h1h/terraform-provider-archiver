package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// make sure we conform to Provider
var _ provider.Provider = &ToolsProvider{}

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

func (t ToolsProvider) Metadata(ctx context.Context,
	req provider.MetadataRequest,
	resp *provider.MetadataResponse,
) {
	resp.TypeName = "tool"
	resp.Version = t.version
}

func (t ToolsProvider) Schema(ctx context.Context,
	req provider.SchemaRequest,
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
				},
			},
		},
	}
}

func (t ToolsProvider) Configure(ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {

}

func (t ToolsProvider) DataSources(ctx context.Context,
) []func() datasource.DataSource {
	return nil
}

func (t ToolsProvider) Resources(ctx context.Context,
) []func() resource.Resource {
	return nil
}
