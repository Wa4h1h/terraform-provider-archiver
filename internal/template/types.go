package template

import (
	"context"
)

type TemplateRunner interface {
	Run(ctx context.Context, src string, data any) error
}

type Template struct{}
