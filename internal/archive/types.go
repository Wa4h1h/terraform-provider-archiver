package archive

import (
	"archive/zip"
	"os"
)

// zip content
// zip one file
// zip list of files
// zip directory
// zip list of directories

// options:
// exclude/resolve symlink
// exclude list of files,dirs
// output file mode
// flatten

type ArchiverType string

const (
	Zip ArchiverType = "zip"
	Tar ArchiverType = "tar"
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

type Archiver interface {
	ArchiveFile(src, dst string) error
	ArchiveDir(src, dst string) error
	ArchiveContent(src []byte, dst string) error
	Open(zipName string, archiveSettings *ArchiveSettings) error
	Close() error
}

type ZipArchive struct {
	zipFile   *os.File
	zipWriter *zip.Writer
	settings  *ArchiveSettings
	fileName  string
}

type TarArchiver struct{}
