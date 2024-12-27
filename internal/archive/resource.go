package archive

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

	archiver := GetArchiver(plan.Type.ValueString())
	if archiver == nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("type"),
			"unsupported archive type",
			fmt.Sprintf("unsported archive type %s, only zip and tar.gz are supported",
				plan.Type.ValueString()))

		return
	}

	var (
		mode    = uint64(DefaultArchiveMode)
		symLink = false
		err     error
	)

	if !plan.OutMode.IsNull() {
		mode, err = strconv.ParseUint(plan.OutMode.ValueString(), 10, 32)
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("out_mode"),
				fmt.Sprintf("can not parse file mode %s", plan.OutMode.ValueString()),
				err.Error())

			return
		}
	}

	if !plan.ResolveSymLink.IsNull() {
		symLink = plan.ResolveSymLink.ValueBool()
	}

	list := make([]string, 0, len(plan.ExcludeList.Elements()))

	resp.Diagnostics.Append(plan.ExcludeList.ElementsAs(ctx, &list, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	archName, err := filepath.Abs(plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"can not resolve path",
			fmt.Sprintf("can not resolve absolute path %s: %s",
				plan.Name.ValueString(), err.Error()))

		return
	}

	err = archiver.Open(archName,
		WithFileMode(os.FileMode(mode)),
		WithSymLink(symLink),
		WithExcludeList(list))
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to create %s", plan.Name.ValueString()),
			err.Error())

		return
	}

	files := make([]File, 0, len(plan.FileBlocks.Elements()))
	resp.Diagnostics.Append(plan.FileBlocks.ElementsAs(ctx, &files, false)...)
	if resp.Diagnostics.HasError() {
		tflog.Warn(ctx, "failed to files to archive")
	}

	dirs := make([]Dir, 0, len(plan.DirBlocks.Elements()))
	resp.Diagnostics.Append(plan.DirBlocks.ElementsAs(ctx, &dirs, false)...)
	if resp.Diagnostics.HasError() {
		tflog.Warn(ctx, "failed to dirs to archive")
	}

	contents := make([]Content, 0, len(plan.ContentBlocks.Elements()))
	resp.Diagnostics.Append(plan.ContentBlocks.ElementsAs(ctx, &contents, false)...)
	if resp.Diagnostics.HasError() {
		tflog.Warn(ctx, "failed to contents to archive")
	}

	for _, f := range files {
		orgPath := f.Path.ValueString()

		absPath, relPath, err := a.cleanPath(orgPath)
		if err != nil {
			tflog.Error(ctx, "can not resolve abs path", map[string]interface{}{
				"org_path": orgPath,
				"err":      err,
			})

			continue
		}

		if err := archiver.ArchiveFile(absPath, relPath); err != nil {
			tflog.Error(ctx, "can not add file to archive",
				map[string]interface{}{
					"path": orgPath,
					"err":  err,
				})
		}
	}

	for _, d := range dirs {
		orgPath := d.Path.ValueString()

		absPath, relPath, err := a.cleanPath(orgPath)
		if err != nil {
			tflog.Error(ctx, "can not resolve abs path", map[string]interface{}{
				"org_path": orgPath,
				"err":      err,
			})

			continue
		}

		if err := archiver.ArchiveDir(absPath, relPath); err != nil {
			tflog.Error(ctx, "can not add file to archive",
				map[string]interface{}{
					"path": orgPath,
					"err":  err,
				})
		}
	}

	for _, c := range contents {
		b, err := base64.StdEncoding.DecodeString(c.Src.ValueString())
		if err != nil {
			tflog.Error(ctx, "can not decode content",
				map[string]interface{}{
					"dst": c.FilePath.ValueString(),
					"err": err,
				})
		}

		relPath := filepath.Clean(c.FilePath.ValueString())

		for strings.HasPrefix(relPath, "../") {
			relPath = strings.TrimPrefix(relPath, "../")
		}

		if err := archiver.ArchiveContent(b, relPath); err != nil {
			tflog.Error(ctx, "can not add content to archive",
				map[string]interface{}{
					"path": relPath,
					"err":  err,
				})
		}
	}

	err = archiver.Close()
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to close %s", plan.Name.ValueString()),
			err.Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (a *archiveResource) Read(ctx context.Context,
	req resource.ReadRequest, resp *resource.ReadResponse,
) {
}

func (a *archiveResource) Update(ctx context.Context,
	req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
	tflog.Debug(ctx, "updating archive....")
	var plan ResourceModel

	d := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(d...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (a *archiveResource) Delete(ctx context.Context,
	req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
	var (
		archive string
		err     error
	)

	d := req.State.GetAttribute(ctx, path.Root("name"), &archive)
	resp.Diagnostics.Append(d...)

	if resp.Diagnostics.HasError() {
		return
	}

	archive, err = filepath.Abs(archive)
	if err != nil {
		resp.Diagnostics.AddError(
			"can not resolve path",
			fmt.Sprintf("can not resolve absolute path %s: %s",
				archive, err.Error()))

		return
	}

	err = os.Remove(archive)
	if err != nil {
		resp.Diagnostics.AddError(
			"can not delete archive",
			fmt.Sprintf("can not delete %s: %s",
				archive, err.Error()))
	}
}

func (a *archiveResource) cleanPath(path string) (string, string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", "", err
	}

	relPath := filepath.Clean(path)

	for strings.HasPrefix(relPath, "../") {
		relPath = strings.TrimPrefix(relPath, "../")
	}

	return absPath, relPath, nil
}
