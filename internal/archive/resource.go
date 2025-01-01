package archive

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"out_mode": schema.StringAttribute{
				Optional:    true,
				Description: "archive file mode: default is 666",
			},
			"resolve_symlink": schema.BoolAttribute{
				Optional:    true,
				Description: "resolve symbolic link: default is false",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"exclude_list": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "list of paths to exclude from the produced archive",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"size": schema.Int64Attribute{
				Computed:    true,
				Description: "Output file size",
			},
			"md5": schema.StringAttribute{
				Computed:    true,
				Description: "Output file computed MD5",
			},
			"sha256": schema.StringAttribute{
				Computed:    true,
				Description: "Output file computed SHA256",
			},
			"abs_path": schema.StringAttribute{
				Computed:    true,
				Description: "Output archive absolute path",
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
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
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
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
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
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (a *archiveResource) ValidateConfig(ctx context.Context,
	req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse,
) {
	tflog.Debug(ctx, "validating resource config...")

	var plan Model

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
	var plan Model

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
		mode    = DefaultArchiveMode
		symLink = false
		err     error
	)

	if !plan.OutMode.IsNull() {
		m, err := strconv.ParseInt(plan.OutMode.ValueString(), 8, 32)
		if err != nil {
			resp.Diagnostics.AddWarning("set archive permission",
				fmt.Sprintf("can not set archive file perimssions: %s", err))
		}

		mode = os.FileMode(m)
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
				plan.Name.ValueString(), err))

		return
	}

	err = archiver.Open(archName,
		WithFileMode(mode),
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
		tflog.Warn(ctx, "failed to add files to archive")
	}

	a.appendFiles(ctx, archiver, files...)

	dirs := make([]Dir, 0, len(plan.DirBlocks.Elements()))
	resp.Diagnostics.Append(plan.DirBlocks.ElementsAs(ctx, &dirs, false)...)
	if resp.Diagnostics.HasError() {
		tflog.Warn(ctx, "failed to add dirs to archive")
	}

	a.appendDirs(ctx, archiver, dirs...)

	contents := make([]Content, 0, len(plan.ContentBlocks.Elements()))
	resp.Diagnostics.Append(plan.ContentBlocks.ElementsAs(ctx, &contents, false)...)
	if resp.Diagnostics.HasError() {
		tflog.Warn(ctx, "failed to add contents to archive")
	}

	a.appendContents(ctx, archiver, contents...)

	err = archiver.Close()
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("failed to close %s", plan.Name.ValueString()),
			err.Error())

		return
	}

	var (
		md5    string
		sha256 string
		size   int64
	)

	md5, sha256, err = a.checksums(archName)
	if err != nil {
		resp.Diagnostics.AddWarning("computing md5 and sha2256",
			fmt.Sprintf("could not compute md5 and sha256 outputs: %s", err))
	}

	plan.MD5 = types.StringValue(md5)
	plan.SHA256 = types.StringValue(sha256)

	size, err = Size(archName)
	if err != nil {
		resp.Diagnostics.AddWarning("compute file size",
			fmt.Sprintf("can not compute file size for %s: %s",
				archName, err))
	}

	plan.Size = types.Int64Value(size)
	plan.AbsPath = types.StringValue(archName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (a *archiveResource) Read(ctx context.Context,
	req resource.ReadRequest, resp *resource.ReadResponse,
) {
	tflog.Debug(ctx, "refreshing state....")

	var state Model

	d := req.State.Get(ctx, &state)

	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	archName := state.AbsPath.ValueString()

	info, err := os.Stat(archName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			resp.State.RemoveResource(ctx)

			state.MD5 = types.StringNull()
			state.SHA256 = types.StringNull()
			state.Size = types.Int64Null()

			return
		}

		resp.Diagnostics.AddWarning(fmt.Sprintf("load %s info", archName),
			fmt.Sprintf("load %s info: %s", archName, err))
	}

	state.Size = types.Int64Value(info.Size())

	md5, sha256, err := a.checksums(archName)
	if err != nil {
		resp.Diagnostics.AddWarning("computing md5 and sha2256",
			fmt.Sprintf("could not refresh md5 and sha256 outputs: %s", err))
	} else {
		state.SHA256 = types.StringValue(sha256)
		state.MD5 = types.StringValue(md5)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (a *archiveResource) Update(ctx context.Context,
	req resource.UpdateRequest, resp *resource.UpdateResponse,
) {
	tflog.Debug(ctx, "updating archive....")
	var (
		plan  Model
		state Model
		err   error
	)

	d := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(d...)

	if resp.Diagnostics.HasError() {
		return
	}

	d = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(d...)

	if resp.Diagnostics.HasError() {
		return
	}

	nameFromState := state.Name.ValueString()

	if !plan.Name.IsNull() {
		newName := plan.Name.ValueString()

		if newName != nameFromState {
			nameFromState, err = filepath.Abs(nameFromState)
			if err != nil {
				resp.Diagnostics.AddWarning("resolve old abs path",
					fmt.Sprintf("could not resolve old abs path %s: %s", nameFromState, err))
			}

			newName, err = filepath.Abs(newName)
			if err != nil {
				resp.Diagnostics.AddWarning("resolve new abs path",
					fmt.Sprintf("could not resolve new abs path %s: %s", newName, err))
			}

			if err == nil {
				err = os.Rename(nameFromState, newName)
				if err != nil {
					resp.Diagnostics.AddWarning(fmt.Sprintf("rename archive file %s", nameFromState),
						fmt.Sprintf("can not rename to %s: %s", newName, err))
				}

				plan.AbsPath = types.StringValue(newName)
			}
		}
	}

	if !plan.OutMode.IsNull() {
		newMode, err := strconv.ParseInt(plan.OutMode.ValueString(), 8, 32)
		if err != nil {
			resp.Diagnostics.AddWarning("change archive permission",
				fmt.Sprintf("can not cahnge archive file perimssions: %s", err))
		} else {
			err = os.Chmod(nameFromState, os.FileMode(newMode))
			if err != nil {
				resp.Diagnostics.AddWarning(fmt.Sprintf("change %s mode", nameFromState),
					fmt.Sprintf("could not change mode to %d: %s", newMode, err))
			}

			plan.OutMode = types.StringValue(fmt.Sprintf("%d", newMode))
		}
	}

	plan.MD5 = state.MD5
	plan.SHA256 = state.SHA256
	plan.Size = state.Size

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (a *archiveResource) Delete(ctx context.Context,
	req resource.DeleteRequest, resp *resource.DeleteResponse,
) {
	var (
		archive string
		err     error
	)

	d := req.State.GetAttribute(ctx, path.Root("abs_path"), &archive)
	resp.Diagnostics.Append(d...)

	if resp.Diagnostics.HasError() {
		return
	}

	err = os.Remove(archive)
	if err != nil {
		resp.Diagnostics.AddError(
			"can not delete archive",
			fmt.Sprintf("can not delete %s: %s",
				archive, err))
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

func (a *archiveResource) checksums(name string) (string, string, error) {
	b, err := os.ReadFile(name)
	if err != nil {
		return "", "", err
	}

	sha256, err := SHA256(b)
	if err != nil {
		return "", "", err
	}

	md5 := MD5(b)

	return md5, sha256, nil
}

func (a *archiveResource) appendFiles(ctx context.Context,
	archiver Archiver, files ...File,
) {
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
}

func (a *archiveResource) appendDirs(ctx context.Context,
	archiver Archiver, dirs ...Dir,
) {
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
			tflog.Error(ctx, "can not add dir to archive",
				map[string]interface{}{
					"path": orgPath,
					"err":  err,
				})
		}
	}
}

func (a *archiveResource) appendContents(ctx context.Context,
	archiver Archiver, contents ...Content,
) {
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
}
