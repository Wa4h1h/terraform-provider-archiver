package archive

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/Wa4h1h/terraform-provider-tools/internal/httpclient"
)

// make sure we conform to Archiver
var _ Archiver = &Archive{}

func NewArchiver(opts ...ArchiverOpt) Archiver {
	a := new(Archive)

	for _, opt := range opts {
		opt(a)
	}

	return a
}

func WithHTTPRunner(httpRunner httpclient.HTTPRunner) ArchiverOpt {
	return func(archive *Archive) {
		archive.HTTPRunner = httpRunner
	}
}

func (a *Archive) writeToZip(w *zip.Writer, src string, flatten bool) error {
	dst := src

	if flatten {
		dst = src[strings.LastIndex(src, "/")+1:]
	}

	fw, err := w.Create(dst)
	if err != nil {
		return fmt.Errorf("error ZipLocal: zipping %s: %w", src, err)
	}

	fr, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error ZipLocal: open zip file %s: %w", src, err)
	}

	b, err := io.ReadAll(fr)
	if err != nil {
		return fmt.Errorf("error ZipLocal: read zip file %s: %w", src, err)
	}

	if _, err := fw.Write(b); err != nil {
		return fmt.Errorf("error ZipLocal: write to zip file %s: %w", src, err)
	}

	return nil
}

func (a *Archive) walkPath(w *zip.Writer, src string, flatten bool) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("error ZipLocal: get file %s info: %w", src, err)
	}

	if !info.IsDir() {
		return a.writeToZip(w, src, flatten)
	} else {
		entries, err := os.ReadDir(src)
		if err != nil {
			return fmt.Errorf("error ZipLocal: read src dir %s: %w", src, err)
		}

		for _, val := range entries {
			if val.IsDir() {
				if err := a.walkPath(w, fmt.Sprintf("%s/%s", src, val.Name()), flatten); err != nil {
					slog.Error(err.Error())

					continue
				}
			} else {
				if err := a.writeToZip(w, fmt.Sprintf("%s/%s", src, val.Name()), flatten); err != nil {
					slog.Error(err.Error())

					continue
				}
			}
		}
	}

	return nil
}

// ZipLocal takes in a src path, dst where to create the zip file and flatten parameter
// flatten when enabled, the zip file will not preserve the directory structure,
// otherwise the entire directory structure will be preserved when compressing
func (a *Archive) ZipLocal(src, dst string, flatten bool) (*os.File, error) {
	f, err := os.Create(dst)
	if err != nil {
		return nil, fmt.Errorf("error ZipLocal: create zip: %w", err)
	}

	w := zip.NewWriter(f)

	if err := a.walkPath(w, src, flatten); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("error ZipLocal: finish writing to %s: %w", f.Name(), err)
	}

	return f, nil
}

func (a *Archive) ZipRemote(ctx context.Context, remote, dst string) (*os.File, error) {
	return nil, nil
}
