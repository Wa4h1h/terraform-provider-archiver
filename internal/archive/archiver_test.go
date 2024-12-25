package archive

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZipArchive_ArchiveFile(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			t.Cleanup(func() {
				os.Remove("test.zip")
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
