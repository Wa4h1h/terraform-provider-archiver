package archive

import (
	"archive/zip"
	"fmt"
	"io/fs"
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
