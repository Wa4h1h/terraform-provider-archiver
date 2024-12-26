package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
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
