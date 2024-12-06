package archive

import (
	"io"

	"github.com/Wa4h1h/terraform-provider-tools/internal/httpclient"
)

// make sure we conform to Archiver
var _ Archiver = &Archive{}

func NewArchiver(opts ...ArchiverOpt) Archiver {
	return &Archive{}
}

func WithHTTPRunner(httpRunner httpclient.HTTPRunner) ArchiverOpt {
	return func(archive *Archive) {
		archive.HTTPRunner = httpRunner
	}
}

func (a *Archive) ZipLocal(r io.Reader, dst string) ([]byte, error) {
	return nil, nil
}

func (a *Archive) ZipRemote(remote string, dst string) ([]byte, error) {
	return nil, nil
}
