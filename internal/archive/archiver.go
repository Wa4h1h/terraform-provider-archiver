// Copyright (c) HashiCorp, Inc.

package archive

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

// make sure we conform to Archiver.
var (
	_ Archiver = &ZipArchiver{}
	_ Archiver = &TarArchiver{}
)

func GetArchiver(archType string) Archiver {
	arch, ok := archivers[archType]
	if !ok {
		return nil
	}

	return arch
}

func WithExcludeList(list []string) Options {
	return func(settings *ArchiveSettings) {
		settings.ExcludeList = list
	}
}

func WithFileMode(mod os.FileMode) Options {
	return func(settings *ArchiveSettings) {
		settings.FileMode = mod
	}
}

func WithSymLink(link bool) Options {
	return func(settings *ArchiveSettings) {
		settings.SymLink = link
	}
}

// evaluateSymLink takes in an absolute path link
// evaluates the symbolic link and returns the underlying absolute path.
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

// resolveExcludeList takes a list of absolute/relative paths
// returns a list of absolute paths.
func resolveExcludeList(list []string) ([]string, error) {
	newExcludeList := make([]string, 0, len(list))

	for _, excludePath := range list {
		excludePath, err := filepath.Abs(excludePath)
		if err != nil {
			return nil, fmt.Errorf("error Open: set abs path for %s: %w", excludePath, err)
		}

		newExcludeList = append(newExcludeList, excludePath)
	}

	return newExcludeList, nil
}

func MD5(b []byte) string {
	return fmt.Sprintf("%x", md5.Sum(b))
}

func SHA256(b []byte) (string, error) {
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
