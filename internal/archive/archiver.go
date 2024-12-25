package archive

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// make sure we conform to Archiver
var (
	_ Archiver = &ZipArchive{}
	_ Archiver = &TarArchiver{}
)

func NewArchiver(archType ArchiverType) Archiver {
	switch archType {
	case Zip:
		return &ZipArchive{}
	case Tar:
		return &TarArchiver{}
	default:
		panic("must not happen: wrong archiver type")
	}
}

// evaluateSymLink takes in an absolute path link
// evaluates the symbolic link and returns the underlying absolute path
func evaluateSymLink(link string) (string, error) {
	absPath := link

	fInfo, err := os.Lstat(link)
	if err != nil {
		return "", fmt.Errorf("error evaluateSymLink: get %s info: %w", link, err)
	}

	// check symlink
	if fInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		orgPath, err := os.Readlink(link)
		if err != nil {
			return "", fmt.Errorf("error evaluateSymLink: evaluate symlink %s: %w",
				orgPath, err)
		}

		absPath, err = filepath.Abs(orgPath)
		if err != nil {
			return "", fmt.Errorf("error evaluateSymLink: get absolute path for %s: %w",
				orgPath, err)
		}
	}

	return absPath, nil
}

func MD5(f *os.File) (string, error) {
	b, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("error MD5: read bytes: %w", err)
	}

	return fmt.Sprintf("%x", md5.Sum(b)), nil
}

func SHA256(f *os.File) (string, error) {
	b, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("error SHA256: read bytes: %w", err)
	}

	hash := sha256.New()

	if _, err := hash.Write(b); err != nil {
		return "", fmt.Errorf("error SHA256: write sha256: %w", err)
	}

	hb := hash.Sum(nil)

	return fmt.Sprintf("%x", hb), nil
}

func Size(file string) (int64, error) {
	stats, err := os.Stat(file)
	if err != nil {
		return 0, fmt.Errorf("error Size: get stats for %s: %w", file, err)
	}

	return stats.Size(), nil
}
