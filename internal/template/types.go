package template

import (
	"context"
	"io"
)

type TemplateRunner interface {
	Run(ctx context.Context, w io.Writer, data any) error
}

type Template struct{}
