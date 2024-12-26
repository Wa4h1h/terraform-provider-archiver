package archive

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.ResourceWithValidateConfig = &archiveResource{}
	_ resource.Resource                   = &archiveResource{}
)

type archiveResource struct{}

type File struct {
	Path types.String `tfsdk:"path"`
}

type Dir struct {
	Path types.String `tfsdk:"path"`
}

type Content struct {
	Src      types.String `tfsdk:"src"`
	FilePath types.String `tfsdk:"file_path"`
}

type ResourceModel struct {
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	OutMode        types.String `tfsdk:"out_mode"`
	ExcludeList    types.List   `tfsdk:"exclude_list"`
	ResolveSymLink types.Bool   `tfsdk:"resolve_symlink"`
	FileBlocks     types.Set    `tfsdk:"file"`
	DirBlocks      types.Set    `tfsdk:"dir"`
	ContentBlocks  types.Set    `tfsdk:"content"`
}

func NewArchiveResource() resource.Resource {
	return &archiveResource{}
}

func (a *archiveResource) Metadata(_ context.Context,
	req resource.MetadataRequest, resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_file"
}

func (a *archiveResource) Schema(_ context.Context,
	_ resource.SchemaRequest, resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "name of the produced archive",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "archive type: zip or tar.gz",
			},
			"out_mode": schema.StringAttribute{
				Optional:    true,
				Description: "archive file mode: default is 666",
			},
			"resolve_symlink": schema.BoolAttribute{
				Optional:    true,
				Description: "resolve symbolic link: default is false",
			},
			"exclude_list": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "list of paths to exclude from the produced archive",
			},
		},
		Blocks: map[string]schema.Block{
			"file": schema.SetNestedBlock{
				Description: "file to include in the archive",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							Required:    true,
							Description: "file path",
						},
					},
				},
			},
			"dir": schema.SetNestedBlock{
				Description: "directory to include in the archive",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							Required:    true,
							Description: "directory path",
						},
					},
				},
			},
			"content": schema.SetNestedBlock{
				Description: "base64 content to include in the archive",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"src": schema.StringAttribute{
							Required:    true,
							Description: "base64 encoded bytes",
						},
						"file_path": schema.StringAttribute{
							Required:    true,
							Description: "file containing the decoded base64 bytes",
						},
					},
				},
			},
		},
	}
}

func (a *archiveResource) ValidateConfig(ctx context.Context,
	req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse,
) {
	tflog.Debug(ctx, "validating resource config...")

	var plan ResourceModel

	d := req.Config.Get(ctx, &plan)
	resp.Diagnostics.Append(d...)

	if resp.Diagnostics.HasError() {
		return
	}

	if plan.OutMode.IsNull() {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("out_mode"),
			"out_mode is null",
			"out_mode is null, the default mode 666 will be used")
	}

	if plan.ResolveSymLink.IsNull() {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("resolve_symlink"),
			"resolve_symlink is null",
			"resolve_symlink is null, the default value false will be used "+
				"and symlinks will not be resolved")
	}
}

func (a *archiveResource) Create(ctx context.Context,
	req resource.CreateRequest, resp *resource.CreateResponse,
) {
	tflog.Debug(ctx, "creating archive....")
	var plan ResourceModel

	d := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(d...)

	if resp.Diagnostics.HasError() {
		return
	}

	archiver := GetArchiver(plan.Type.String())
	if archiver == nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("type"),
			"unsupported archive type",
			fmt.Sprintf("unsported archive type %s, "+
				"only zip and tar.gz are supported", plan.Type.String()))

		return
	}

	mode, err := strconv.ParseUint(plan.OutMode.String(), 10, 32)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("out_mode"),
			fmt.Sprintf("can not parse file mode %s", plan.OutMode.String()),
			err.Error())

		return
	}

	sym, err := strconv.ParseBool(plan.ResolveSymLink.String())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("resolve_symlink"),
			fmt.Sprintf("can not parse resolve_symlink %s", plan.ResolveSymLink.String()),
			"false will be used a value: "+err.Error())
	}

	list := make([]string, 0, len(plan.ExcludeList.Elements()))

	resp.Diagnostics.Append(plan.ExcludeList.ElementsAs(ctx, &list, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err = archiver.Open(plan.Name.String(),
		WithFileMode(os.FileMode(mode)),
		WithSymLink(sym),
		WithExcludeList(list))
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to create %s", plan.Name.String()),
			err.Error())

		return
	}

	files := make([]File, 0, len(plan.FileBlocks.Elements()))

	resp.Diagnostics.Append(plan.FileBlocks.ElementsAs(ctx, &files, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, f := range files {
		tflog.Debug(ctx, "path --> "+f.Path.String())
	}

	err = archiver.Close()
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to close %s", plan.Name.String()),
			err.Error())

		return
	}
}

func (a *archiveResource) Read(ctx context.Context,
	req resource.ReadRequest, resp *resource.ReadResponse,
) {
}

func (a *archiveResource) Update(ctx context.Context,
	req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
}

func (a *archiveResource) Delete(ctx context.Context,
	req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
}
