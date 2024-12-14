package archive

import (
	"archive/zip"
	"crypto/md5"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// make sure we conform to Archiver
var (
	_ Archiver = &ZipArchive{}
)

func NewArchiver(archType ArchiverType) Archiver {
	return &ZipArchive{}
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

func (z *ZipArchive) writeToZip(file string, flatten bool) error {
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("error writeToZip: open %s: %w", file, err)
	}

	defer f.Close()

	// check flatten
	if flatten {
		file = file[strings.LastIndex(file, "/")+1:]
	}

	w, err := z.zipWriter.Create(file)
	if err != nil {
		return fmt.Errorf("error writeToZip: create %s writer: %w", file, err)
	}

	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("error ArchiveFile: write to zip: %w", err)
	}

	return nil
}

func (z *ZipArchive) ArchiveFile(file string, flatten bool,
) error {
	var (
		path = file
		err  error
	)

	if z.settings.SymLink {
		path, err = evaluateSymLink(file)
		if err != nil {
			return err
		}
	}

	if err := z.writeToZip(path, flatten); err != nil {
		return err
	}

	return nil
}

func (z *ZipArchive) ArchiveDir(src string, flatten bool) error {
	var (
		path = src
		err  error
	)

	if z.settings.SymLink {
		path, err = evaluateSymLink(src)
		if err != nil {
			return err
		}
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("error ArchiveDir: read dirs under %s: %w", path, err)
	}

	for _, entry := range entries {
		tmpPath := fmt.Sprintf("%s/%s", src, entry.Name())

		if !entry.IsDir() {
			if err := z.writeToZip(tmpPath, flatten); err != nil {
				log.Printf("error ArchiveDir: write to zip %s: %s", tmpPath, err)
			}
		} else {
			if err := z.ArchiveDir(tmpPath, flatten); err != nil {
				log.Printf("error ArchiveDir: write to zip %s: %s", tmpPath, err)
			}
		}
	}

	return nil
}

func (z *ZipArchive) ArchiveContent(src []byte, dst string) error {
	w, err := z.zipWriter.Create(dst)
	if err != nil {
		return fmt.Errorf("error ArchiveContent: append file %s to zip: %w",
			dst, err)
	}

	if _, err := w.Write(src); err != nil {
		return fmt.Errorf("error ArchiveContent: write to zip: %w", err)
	}

	return nil
}

func (z *ZipArchive) Open(zipName string, archiveSettings *ArchiveSettings) error {
	f, err := os.OpenFile(zipName, os.O_TRUNC|os.O_CREATE|os.O_RDWR, archiveSettings.FileMode)
	if err != nil {
		return fmt.Errorf("error: Create zip file %s: %w", zipName, err)
	}

	z.zipFile = f
	z.fileName = zipName
	z.zipWriter = zip.NewWriter(f)
	z.settings = archiveSettings

	return nil
}

func (z *ZipArchive) Close() error {
	var err error

	if err = z.zipWriter.Close(); err != nil {
		err = errors.Join(fmt.Errorf("error Close: zip %s: %w", z.fileName, err))
	}

	if err = z.zipFile.Close(); err != nil {
		err = errors.Join(fmt.Errorf("error Close: zip writer: %w", err))
	}

	return err
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
