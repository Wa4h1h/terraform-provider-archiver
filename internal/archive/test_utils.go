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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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

var fileTestCases = []struct {
	name    string
	routine func(*testing.T) (string, string)
}{
	{
		name: "SrcAbsDstRel_CreateArchive",
		routine: func(t *testing.T) (string, string) {
			relPath := "../../internal/random/types.go"

			path, err := filepath.Abs(relPath)

			require.Nil(t, err)

			for strings.HasPrefix(relPath, "../") {
				relPath = strings.TrimPrefix(relPath, "../")
			}

			return path, relPath
		},
	},
	{
		name: "SrcAbsDstAbs_CreateArchive",
		routine: func(t *testing.T) (string, string) {
			var err error
			path := "../../internal/random/types.go"

			path, err = filepath.Abs(path)

			require.Nil(t, err)

			return path, path
		},
	},
	{
		name: "SrcAbsSymLinkDstRel_CreateArchive",
		routine: func(t *testing.T) (string, string) {
			relPath := "../../internal/random/types.go"

			path, err := filepath.Abs(relPath)

			require.Nil(t, err)

			for strings.HasPrefix(relPath, "../") {
				relPath = strings.TrimPrefix(relPath, "../")
			}

			symLink := "symlink-path"

			err = os.Symlink(path, symLink)

			require.Nil(t, err)

			path, err = filepath.Abs(symLink)

			require.Nil(t, err)

			return path, relPath
		},
	},
}
