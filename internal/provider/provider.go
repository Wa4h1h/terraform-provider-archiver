package provider

import (
	"context"

	"github.com/Wa4h1h/terraform-provider-archiver/internal/archive"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// make sure we conform to Provider
var _ provider.Provider = &ArchiverProvider{}

type ArchiverProvider struct {
	version string
}

func NewArchiverProvider(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ArchiverProvider{
			version: version,
		}
	}
}

func (t *ArchiverProvider) Metadata(_ context.Context,
	_ provider.MetadataRequest,
	resp *provider.MetadataResponse,
) {
	resp.TypeName = "archiver"
	resp.Version = t.version
}

func (t *ArchiverProvider) Schema(_ context.Context,
	_ provider.SchemaRequest,
	_ *provider.SchemaResponse,
) {
}

func (t *ArchiverProvider) Configure(_ context.Context,
	_ provider.ConfigureRequest,
	_ *provider.ConfigureResponse,
) {
}

func (t *ArchiverProvider) DataSources(_ context.Context,
) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (t *ArchiverProvider) Resources(_ context.Context,
) []func() resource.Resource {
	return []func() resource.Resource{
		archive.NewArchiveResource,
	}
}
