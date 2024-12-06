package template

import (
	"context"
	"io"
)

// make sure we conform to TemplateRunner
var _ TemplateRunner = &Template{}

func NewTemplateRunner() TemplateRunner {
	return &Template{}
}

func (t *Template) Run(ctx context.Context, w io.Writer, data any) error {
	return nil
}
