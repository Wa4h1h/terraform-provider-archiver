// Copyright (c) HashiCorp, Inc.

package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func getZipContentFullPaths(src string) ([]string, error) {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return nil, fmt.Errorf("error getZipContentFullPaths: open zip reader: %w", err)
	}

	defer reader.Close()

	files := make([]string, 0, len(reader.File))

	for _, file := range reader.File {
		files = append(files, file.Name)
	}

	return files, nil
}

func getTarContentFullPaths(src string) ([]string, error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("error getTarContentFullPaths: %w", err)
	}

	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("error getTarContentFullPaths: open gzip reader: %w", err)
	}

	r := tar.NewReader(gr)
	files := make([]string, 0)

	for {
		h, err := r.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, fmt.Errorf("error getTarContentFullPaths: read header: %w", err)
		}

		files = append(files, h.Name)
	}

	return files, nil
}

func getFilePathFromDir(src string) ([]string, error) {
	files := make([]string, 0)

	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error getFilePathFromDir: walk %s: %w", src, err)
	}

	return files, nil
}
