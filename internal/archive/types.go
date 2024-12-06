package archive

import (
	"context"
	"os"

	"github.com/Wa4h1h/terraform-provider-tools/internal/httpclient"
)

type Archiver interface {
	ZipLocal(src, dst string, flatter bool) (*os.File, error)
	ZipRemote(ctx context.Context, remote string, dst string) (*os.File, error)
}

type Archive struct {
	httpclient.HTTPRunner
}

type ArchiverOpt func(*Archive)
