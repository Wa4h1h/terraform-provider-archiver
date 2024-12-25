package archive

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZipArchive_ArchiveFile(t *testing.T) {
	for _, testCase := range fileTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Cleanup(func() {
				os.Remove("test.zip")
				os.Remove("symlink-path")
			})

			a := NewArchiver(Zip)

			src, dst := testCase.routine(t)

			err := a.Open("test.zip",
				WithFileMode(0o666),
				WithSymLink(true))

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
	for _, testCase := range dirTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Cleanup(func() {
				os.Remove("test.zip")
				os.Remove("symlink-path")
			})

			a := NewArchiver(Zip)

			src, dst := testCase.routine(t)

			err := a.Open("test.zip",
				WithFileMode(0o666),
				WithSymLink(true))

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
	for _, testCase := range bytesTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Cleanup(func() {
				os.Remove("test.zip")
			})

			a := NewArchiver(Zip)

			b, dst := testCase.routine(t)

			err := a.Open("test.zip",
				WithFileMode(0o666),
				WithSymLink(true))

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

func TestTarArchiver_ArchiveFile(t *testing.T) {
	for _, testCase := range fileTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Cleanup(func() {
				os.Remove("test.tar.gz")
				os.Remove("symlink-path")
			})

			a := NewArchiver(Tar)

			src, dst := testCase.routine(t)

			err := a.Open("test.tar.gz")

			require.Nil(t, err)

			err = errors.Join(a.ArchiveFile(src, dst), a.Close())

			require.Nil(t, err)

			paths, err := getTarContentFullPaths("test.tar.gz")

			require.Nil(t, err)

			assert.Equal(t, 1, len(paths))
			assert.Equal(t, dst, paths[0])
		})
	}
}

func TestTarArchiver_ArchiveDir(t *testing.T) {
	for _, testCase := range dirTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Cleanup(func() {
				os.Remove("test.tar.gz")
				os.Remove("symlink-path")
			})

			a := NewArchiver(Tar)

			src, dst := testCase.routine(t)

			err := a.Open("test.tar.gz", WithSymLink(true))

			require.Nil(t, err)

			err = errors.Join(a.ArchiveDir(src, dst), a.Close())

			require.Nil(t, err)

			paths, err := getTarContentFullPaths("test.tar.gz")

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

func TestTarArchiver_ArchiveContent(t *testing.T) {
	for _, testCase := range bytesTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Cleanup(func() {
				os.Remove("test.tar.gz")
			})

			a := NewArchiver(Tar)

			b, dst := testCase.routine(t)

			err := a.Open("test.tar.gz")

			require.Nil(t, err)

			err = errors.Join(a.ArchiveContent(b, dst), a.Close())

			require.Nil(t, err)

			f, err := os.Open("test.tar.gz")

			require.Nil(t, err)

			t.Cleanup(func() {
				f.Close()
			})

			gr, err := gzip.NewReader(f)

			require.Nil(t, err)

			r := tar.NewReader(gr)

			header, err := r.Next()

			require.Nil(t, err)

			assert.Equal(t, dst, header.Name)

			buff := new(bytes.Buffer)

			_, err = io.Copy(buff, r)

			require.Nil(t, err)

			assert.Equal(t, byteInput, buff.Bytes())
		})
	}
}
