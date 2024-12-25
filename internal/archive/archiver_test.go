package archive

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZipArchive_ArchiveFile(t *testing.T) {
	testCases := []struct {
		name    string
		routine func(*testing.T) (string, string)
	}{
		{
			name: "SrcAbsDstRel_CreateZip",
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
			name: "SrcAbsDstAbs_CreateZip",
			routine: func(t *testing.T) (string, string) {
				var err error
				path := "../../internal/random/types.go"

				path, err = filepath.Abs(path)

				require.Nil(t, err)

				return path, path
			},
		},
		{
			name: "SrcAbsSymLinkDstRel_CreateZip",
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

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Cleanup(func() {
				os.Remove("test.zip")
				os.Remove("symlink-path")
			})

			a := NewArchiver(Zip)

			src, dst := testCase.routine(t)

			err := a.Open("test.zip", &ArchiveSettings{
				FileMode: 0o666,
				SymLink:  true,
			})

			require.Nil(t, err)

			err = errors.Join(a.ArchiveFile(src, dst), a.Close())

			require.Nil(t, err)

			paths, err := getZipContentFullPaths("test.zip")

			require.Nil(t, err)

			assert.Equal(t, 1, len(paths))
			assert.Equal(t, dst, paths[0])
		})
	}
}

func TestZipArchive_ArchiveDir(t *testing.T) {
	testCases := []struct {
		name    string
		routine func(*testing.T) (string, string)
	}{
		{
			name: "SrcAbsDstRel_CreateZip",
			routine: func(t *testing.T) (string, string) {
				relPath := "../../internal/random"

				path, err := filepath.Abs(relPath)

				require.Nil(t, err)

				for strings.HasPrefix(relPath, "../") {
					relPath = strings.TrimPrefix(relPath, "../")
				}

				return path, relPath
			},
		},
		{
			name: "SrcAbsDstAbs_CreateZip",
			routine: func(t *testing.T) (string, string) {
				var err error
				path := "../../internal/random/"

				path, err = filepath.Abs(path)

				require.Nil(t, err)

				return path, path
			},
		},
		{
			name: "SrcAbsSymLinkDstRel_CreateZip",
			routine: func(t *testing.T) (string, string) {
				relPath := "../../internal/random"

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

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Cleanup(func() {
				os.Remove("test.zip")
				os.Remove("symlink-path")
			})

			a := NewArchiver(Zip)

			src, dst := testCase.routine(t)

			err := a.Open("test.zip", &ArchiveSettings{
				FileMode: 0o666,
				SymLink:  true,
			})

			require.Nil(t, err)

			err = errors.Join(a.ArchiveDir(src, dst), a.Close())

			require.Nil(t, err)

			paths, err := getZipContentFullPaths("test.zip")

			require.Nil(t, err)

			src, err = evaluateSymLink(src)

			require.Nil(t, err)

			readPaths, err := getFilePathFromDir(src)

			require.Nil(t, err)

			require.Equal(t, len(readPaths), len(paths))

			for i, path := range readPaths {
				assert.True(t, strings.HasSuffix(path, paths[i]))
			}
		})
	}
}

func TestZipArchive_ArchiveContent(t *testing.T) {
	byteInput := []byte("test input")

	testCases := []struct {
		name    string
		routine func(*testing.T) ([]byte, string)
	}{
		{
			name: "SrcBytesDstRel_CreateZip",
			routine: func(t *testing.T) ([]byte, string) {
				return byteInput, "./test.txt"
			},
		},
		{
			name: "SrcBytesDstAbs_CreateZip",
			routine: func(t *testing.T) ([]byte, string) {
				path, err := os.Getwd()

				require.Nil(t, err)

				filepath.Join(path, "test.txt")

				return byteInput, path
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Cleanup(func() {
				os.Remove("test.zip")
			})

			a := NewArchiver(Zip)

			b, dst := testCase.routine(t)

			err := a.Open("test.zip", &ArchiveSettings{
				FileMode: 0o666,
				SymLink:  true,
			})

			require.Nil(t, err)

			err = errors.Join(a.ArchiveContent(b, dst), a.Close())

			require.Nil(t, err)

			paths, err := getZipContentFullPaths("test.zip")

			require.Nil(t, err)

			assert.Equal(t, 1, len(paths))
			assert.Equal(t, dst, paths[0])

			reader, err := zip.OpenReader("test.zip")

			require.Nil(t, err)

			require.Equal(t, 1, len(reader.File))

			f := reader.File[0]

			rc, err := f.Open()
			require.Nil(t, err)

			by, err := io.ReadAll(rc)
			require.Nil(t, err)

			assert.Equal(t, b, by)
		})
	}
}
