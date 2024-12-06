package provider

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/Wa4h1h/terraform-provider-tools/internal/archive"
	"github.com/Wa4h1h/terraform-provider-tools/internal/httpclient"
	"github.com/Wa4h1h/terraform-provider-tools/internal/random"
	"github.com/Wa4h1h/terraform-provider-tools/internal/template"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// make sure we conform to Provider
var _ provider.ProviderWithValidateConfig = &ToolsProvider{}

type Tools struct {
	archive.Archiver
	httpclient.HTTPRunner
	random.Randomizer
	template.TemplateRunner
}

func NewTools(httpRunner httpclient.HTTPRunner,
	archiverOpts ...archive.ArchiverOpt,
) *Tools {
	return &Tools{
		Archiver:       archive.NewArchiver(archiverOpts...),
		HTTPRunner:     httpRunner,
		Randomizer:     random.NewRandomizer(),
		TemplateRunner: template.NewTemplateRunner(),
	}
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
	resp.TypeName = "tools"
	resp.Version = t.version
}

func (t *ToolsProvider) Schema(_ context.Context,
	_ provider.SchemaRequest,
	resp *provider.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"http": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"headers": schema.MapAttribute{
						Optional:    true,
						ElementType: types.StringType,
					},
					"hostname": schema.StringAttribute{
						Optional: true,
					},
					"timeout": schema.Int32Attribute{
						Optional: true,
					},
				},
			},
		},
	}
}

type ToolsProviderModel struct {
	Http types.Object `tfsdk:"http"`
}

func (t *ToolsProvider) ValidateConfig(ctx context.Context,
	req provider.ValidateConfigRequest,
	resp *provider.ValidateConfigResponse,
) {
	tflog.Info(ctx, "Validation Tools configuration")

	// load provider config
	var data ToolsProviderModel

	d := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpPath := path.Root("http")

	// assert if http config object value is known by the time the provider is being initialized
	if data.Http.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			httpPath,
			"Unknown http config",
			"The provider cannot create http client",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	httpTimeoutAttr, ok := data.Http.Attributes()["timeout"]
	if ok {
		if httpTimeoutAttr.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				httpPath.AtMapKey("timeout"),
				"Unknown http timeout",
				"The provider cannot create http client",
			)
		}
	}

	httpHeadersAttr, ok := data.Http.Attributes()["headers"]
	if ok {
		if httpHeadersAttr.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				httpPath.AtMapKey("headers"),
				"Unknown http headers",
				"The provider cannot create http client",
			)
		}
	}

	httpHostname, ok := data.Http.Attributes()["hostname"]
	if ok {
		if httpHostname.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				httpPath.AtMapKey("hostname"),
				"Unknown http hostname",
				"The provider cannot create http client",
			)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}
}

func (t *ToolsProvider) Configure(ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	tflog.Info(ctx, "Configuring Tools client")

	// load provider config
	var (
		data      ToolsProviderModel
		tfHeaders map[string]tftypes.Value
		tfTimeout big.Float
		hostname  string
	)

	d := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpAttrs := data.Http.Attributes()

	headersVal, ok := httpAttrs["headers"]
	if ok {
		tfVal, err := headersVal.ToTerraformValue(ctx)
		if err == nil {
			err = tfVal.As(&tfHeaders)
			if err != nil {
				d.AddAttributeWarning(path.Root("http").AtMapKey("headers"),
					"Fail to get headers",
					fmt.Sprintf("The provider can not populate default headers: %s", err))
			}
		}

		d.AddAttributeWarning(path.Root("http").AtMapKey("headers"),
			"Fail to get headers",
			fmt.Sprintf("The provider can not populate default headers: %s", err))
	}

	timeoutVal, ok := httpAttrs["timeout"]
	if ok {
		tfVal, err := timeoutVal.ToTerraformValue(ctx)
		if err == nil {
			err = tfVal.As(&tfTimeout)
			if err != nil {
				d.AddAttributeWarning(path.Root("http").AtMapKey("timeout"),
					"Fail to get timeout",
					fmt.Sprintf("The provider can not populate default timeout: %s", err))
			}
		}

		d.AddAttributeWarning(path.Root("http").AtMapKey("timeout"),
			"Fail to get timeout",
			fmt.Sprintf("The provider can not populate default timeout: %s", err))
	}

	hostnameVal, ok := httpAttrs["hostname"]
	if ok {
		hostname = hostnameVal.String()
	}

	headers := make(map[string]string)
	for key, val := range tfHeaders {
		headers[key] = val.String()
	}

	var timeout big.Int

	tm, _ := tfTimeout.Int(&timeout)

	httpRunner := httpclient.NewHTTPRunner(httpclient.WithTimeout(int(tm.Int64())),
		httpclient.WithHostname(hostname), httpclient.WithHeaders(headers))

	tools := NewTools(httpRunner, archive.WithHTTPRunner(httpRunner))

	resp.DataSourceData = tools
	resp.ResourceData = tools
}

func (t *ToolsProvider) DataSources(ctx context.Context,
) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (t *ToolsProvider) Resources(ctx context.Context,
) []func() resource.Resource {
	return []func() resource.Resource{}
}
