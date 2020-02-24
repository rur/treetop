package treetop

import (
	"fmt"
	"html/template"
)

// StringExecutor loads view templates as an inline template string.
// TODO: Implement this
type StringExecutor struct{}

// NewViewHandler will create a Handler instance capable of serving treetop requests
// for the supplied view configuration
//
// TODO: Implement this
func (se *StringExecutor) NewViewHandler(view *View, includes ...*View) ViewHandler {
	page, part, incls := CompileViews(view, includes...)
	handler := TemplateHandler{
		Page:            page,
		PageTemplate:    se.MustParseTemplateFiles(page),
		Partial:         part,
		PartialTemplate: se.MustParseTemplateFiles(part),
		Includes:        incls,
	}
	for _, inc := range incls {
		handler.IncludeTemplates = append(handler.IncludeTemplates, se.MustParseTemplateFiles(inc))
	}
	return handler
}

// MustParseTemplateFiles will load template files and parse contents into a HTML template instance
// it will use the supplied http.FileSystem for loading the template files
//
// TODO: Implement this
func (se *StringExecutor) MustParseTemplateFiles(view *View) *template.Template {
	var out *template.Template
	// snippet based upon https://golang.org/pkg/html/template/#ParseFiles implementation
	for i, s := range GetTemplateList(view) {
		name := fmt.Sprintf("Template[%v]", i)
		var t *template.Template
		if out == nil {
			// first file in the list is used as the root template
			out = template.New(name)
			t = out
		} else {
			t = out.New(name)
		}
		_, err := t.Parse(s)
		if err != nil {
			panic(fmt.Sprintf("Error parsing template %s, error %s", name, err.Error()))
		}
	}
	return out
}
