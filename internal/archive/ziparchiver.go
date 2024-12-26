package archive

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// writeToZip create a new file dst inside the zip file
// copies src content to the newly created dst file.
func (z *ZipArchiver) writeToZip(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error writeToZip: open %s: %w", src, err)
	}

	defer f.Close()

	w, err := z.zipWriter.Create(dst)
	if err != nil {
		return fmt.Errorf("error writeToZip: create %s writer: %w", dst, err)
	}

	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("error writeToZip: write to zip: %w", err)
	}

	return nil
}

// ArchiveFile accepts an absolute path src  and any other path dst
// every symbolic link is evaluated if SymLink is set to true
// call writeToZip, to write src content to dst.
func (z *ZipArchiver) ArchiveFile(src, dst string) error {
	var err error

	if slices.Contains(z.settings.ExcludeList, src) {
		return nil
	}

	if z.settings.SymLink {
		src, err = evaluateSymLink(src)
		if err != nil {
			return err
		}
	}

	if err := z.writeToZip(src, dst); err != nil {
		return err
	}

	return nil
}

// ArchiveDir accepts an absolute path src  and any other path dst
// loops recursively through src path and calls ArchiveFile on each encountered file
// to add it  to zip file
// every symbolic link is evaluated if SymLink is set to true.
func (z *ZipArchiver) ArchiveDir(src, dst string) error {
	var err error

	if z.settings.SymLink {
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

			if err := z.ArchiveFile(tmpPath, fPath); err != nil {
				log.Printf("error ArchiveDir: write to zip %s: %s", tmpPath, err)
			}
		} else {
			if err := z.ArchiveDir(tmpPath, dst); err != nil {
				log.Printf("error ArchiveDir: write to zip %s: %s", tmpPath, err)
			}
		}
	}

	return nil
}

// ArchiveContent accepts a slice of bytes and dst path
// it creates a new dst file within the zip and write they bytes into it.
func (z *ZipArchiver) ArchiveContent(src []byte, dst string) error {
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

func (z *ZipArchiver) Open(zipName string, opts ...Options) error {
	archiveSettings := &ArchiveSettings{
		FileMode: DefaultArchiveMode,
	}

	for _, opt := range opts {
		opt(archiveSettings)
	}

	f, err := os.OpenFile(zipName, os.O_TRUNC|os.O_CREATE|os.O_RDWR, archiveSettings.FileMode)
	if err != nil {
		return fmt.Errorf("error: Create zip file %s: %w", zipName, err)
	}

	z.zipFile = f
	z.fileName = zipName
	z.zipWriter = zip.NewWriter(f)
	z.settings = archiveSettings

	if z.settings.ExcludeList != nil {
		z.settings.ExcludeList, err = resolveExcludeList(z.settings.ExcludeList)
		if err != nil {
			return err
		}
	}

	return nil
}

func (z *ZipArchiver) Close() error {
	err := errors.Join(z.zipWriter.Close(), z.zipFile.Close())
	if err != nil {
		return fmt.Errorf("error Close: %w", err)
	}

	return nil
}
