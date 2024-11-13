package template

// make sure we conform to TemplateRunner
var _ TemplateRunner = &Template{}

type TemplateRunner interface{}

type Template struct{}

func NewTemplateRunner() *Template {
	return &Template{}
}
