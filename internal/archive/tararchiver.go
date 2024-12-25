package archive

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// writeToTar create a new file dst inside the tarball
// copies src content to the newly created dst file
func (t *TarArchiver) writeToTar(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error writeToTar: open %s: %w", src, err)
	}

	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("error writeToTar: get info %s: %w", f.Name(), err)
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return fmt.Errorf("error writeToTar: set header infor: %w", err)
	}

	header.Name = dst

	err = t.tarWriter.WriteHeader(header)
	if err != nil {
		return fmt.Errorf("error writeToTar: write header: %w", err)
	}

	if _, err := io.Copy(t.tarWriter, f); err != nil {
		return fmt.Errorf("error writeToTar: write to tar: %w", err)
	}

	return nil
}

// ArchiveFile accepts an absolute path src  and any other path dst
// every symbolic link is evaluated if SymLink is set to true
// call writeToTar, to write src content to dst.
func (t *TarArchiver) ArchiveFile(src, dst string) error {
	var err error

	if slices.Contains(t.settings.ExcludeList, src) {
		return nil
	}

	if t.settings.SymLink {
		src, err = evaluateSymLink(src)
		if err != nil {
			return err
		}
	}

	if err := t.writeToTar(src, dst); err != nil {
		return err
	}

	return nil
}

// ArchiveDir accepts an absolute path src  and any other path dst
// loops recursively through src path and calls ArchiveFile on each encountered file
// to add it to the tarball
// every symbolic link is evaluated if SymLink is set to true.
func (t *TarArchiver) ArchiveDir(src, dst string) error {
	var err error

	if t.settings.SymLink {
		src, err = evaluateSymLink(src)
		if err != nil {
			return err
		}
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("error ArchiveDir: read dirs under %s: %w", src, err)
	}

	for _, entry := range entries {
		tmpPath := filepath.Join(src, entry.Name())

		if !entry.IsDir() {
			relPathIndex := strings.Index(tmpPath, dst)
			fPath := tmpPath

			if relPathIndex != -1 {
				fPath = tmpPath[relPathIndex:]
			}

			if err := t.ArchiveFile(tmpPath, fPath); err != nil {
				log.Printf("error ArchiveDir: write to zip %s: %s", tmpPath, err)
			}
		} else {
			if err := t.ArchiveDir(tmpPath, dst); err != nil {
				log.Printf("error ArchiveDir: write to zip %s: %s", tmpPath, err)
			}
		}
	}

	return nil
}

func (t *TarArchiver) ArchiveContent(src []byte, dst string) error {
	err := t.tarWriter.WriteHeader(&tar.Header{
		Name:     dst,
		Size:     int64(len(src)),
		Mode:     770,
		ModTime:  time.Now(),
		Typeflag: tar.TypeReg,
	})
	if err != nil {
		return fmt.Errorf("error ArchiveContent: append file %s to zip: %w",
			dst, err)
	}

	if _, err := t.tarWriter.Write(src); err != nil {
		return fmt.Errorf("error ArchiveContent: write to zip: %w", err)
	}

	return nil
}

func (t *TarArchiver) Open(tarName string, opts ...Options) error {
	archiveSettings := &ArchiveSettings{
		FileMode: DefaultArchiveMode,
	}

	for _, opt := range opts {
		opt(archiveSettings)
	}

	f, err := os.OpenFile(tarName, os.O_TRUNC|os.O_CREATE|os.O_RDWR, archiveSettings.FileMode)
	if err != nil {
		return fmt.Errorf("error: Create zip file %s: %w", tarName, err)
	}

	t.tarFile = f
	t.fileName = tarName
	t.gzipWriter = gzip.NewWriter(f)
	t.tarWriter = tar.NewWriter(t.gzipWriter)
	t.settings = archiveSettings

	if t.settings.ExcludeList != nil {
		t.settings.ExcludeList, err = resolveExcludeList(t.settings.ExcludeList)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *TarArchiver) Close() error {
	err := errors.Join(t.tarWriter.Close(),
		t.gzipWriter.Close(),
		t.tarFile.Close())
	if err != nil {
		return fmt.Errorf("error Close: %w", err)
	}

	return nil
}
