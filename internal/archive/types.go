package archive

import (
	"io"

	"github.com/Wa4h1h/terraform-provider-tools/internal/httpclient"
)

type Archiver interface {
	ZipLocal(r io.Reader, dst string) ([]byte, error)
	ZipRemote(remote string, dst string) ([]byte, error)
}

type Archive struct {
	httpclient.HTTPRunner
}

type ArchiverOpt func(*Archive)
