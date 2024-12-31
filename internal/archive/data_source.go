package archive

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &archiveDataSource{}

type archiveDataSource struct{}

func NewArchiveDataSource() datasource.DataSource {
	return &archiveDataSource{}
}

func (a *archiveDataSource) Metadata(ctx context.Context,
	req datasource.MetadataRequest, resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_file"
}

func (a *archiveDataSource) Schema(ctx context.Context,
	req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
}

func (a *archiveDataSource) Read(ctx context.Context,
	req datasource.ReadRequest, resp *datasource.ReadResponse) {
}
