package archive

import (
	"archive/zip"
	"fmt"
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
