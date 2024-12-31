package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ArchiverResult struct {
	Sha256 string
	MD5    string
	Size   int64
}

type ArchiveSettings struct {
	// files/dirs to exclude during archiving
	ExcludeList []string
	// octal file mode of the created archive
	FileMode os.FileMode
	// include symbolic links
	SymLink bool
}

type Options func(*ArchiveSettings)

type Archiver interface {
	ArchiveFile(src, dst string) error
	ArchiveDir(src, dst string) error
	ArchiveContent(src []byte, dst string) error
	Open(zipName string, opts ...Options) error
	Close() error
}

type ZipArchiver struct {
	zipFile   *os.File
	zipWriter *zip.Writer
	settings  *ArchiveSettings
	fileName  string
}

type TarArchiver struct {
	tarFile    *os.File
	gzipWriter *gzip.Writer
	tarWriter  *tar.Writer
	settings   *ArchiveSettings
	fileName   string
}

var archivers = map[string]Archiver{
	"zip":    &ZipArchiver{},
	"tar.gz": &TarArchiver{},
}

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

type Model struct {
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	OutMode        types.String `tfsdk:"out_mode"`
	MD5            types.String `tfsdk:"md5"`
	SHA256         types.String `tfsdk:"sha256"`
	AbsPath        types.String `tfsdk:"abs_path"`
	ExcludeList    types.List   `tfsdk:"exclude_list"`
	ResolveSymLink types.Bool   `tfsdk:"resolve_symlink"`
	FileBlocks     types.Set    `tfsdk:"file"`
	DirBlocks      types.Set    `tfsdk:"dir"`
	ContentBlocks  types.Set    `tfsdk:"content"`
	Size           types.Int64  `tfsdk:"size"`
}
