package template

import (
	"context"
)

// make sure we conform to TemplateRunner
var _ TemplateRunner = &Template{}

func NewTemplateRunner() TemplateRunner {
	return &Template{}
}

func (t *Template) Run(ctx context.Context, src string, data any) error {
	return nil
}
