package treetop

import (
	"html/template"

	"github.com/rur/treetop/internal"
)

type StringExecutor struct {
	TemplateExecutor
	Funcs template.FuncMap
}

func (se *StringExecutor) NewViewHandler(view *View, includes ...*View) ViewHandler {
	se.TemplateExecutor.Funcs = se.Funcs
	se.TemplateExecutor.Loader = internal.LoadStringTemplate
	return se.TemplateExecutor.NewViewHandler(view, includes...)
}
