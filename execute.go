package treetop

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

// DefaultExecutor loads view templates as a path from a Go HTML template file.
// The underlying file system is abstracted through the http.FileSystem interface to allow for
// in-memory use.
type DefaultExecutor struct{}

// NewViewHandler will create a Handler instance capable of serving treetop requests
// for the supplied view configuration
// TODO: Implement this
func (de *DefaultExecutor) NewViewHandler(view *View, includes ...*View) ViewHandler {
	page, part, incls := CompileViews(view, includes...)
	handler := TemplateHandler{
		Page:            page,
		PageTemplate:    de.MustParseTemplateFiles(page),
		Partial:         part,
		PartialTemplate: de.MustParseTemplateFiles(part),
		Includes:        incls,
	}
	for _, inc := range incls {
		handler.IncludeTemplates = append(handler.IncludeTemplates, de.MustParseTemplateFiles(inc))
	}
	return handler
}

// MustParseTemplateFiles will load template files and parse contents into a HTML template instance
// this is similar to html/template ParseFiles function
// TODO: Implement this
func (de *DefaultExecutor) MustParseTemplateFiles(view *View) *template.Template {
	var t *template.Template
	// snippet based upon https://golang.org/pkg/html/template/#ParseFiles implementation
	for _, filename := range GetTemplateList(view) {
		buffer := new(bytes.Buffer)
		file, err := os.Open(filename)
		if err != nil {
			panic(fmt.Sprintf("Failed to open template file '%s', error %s", filename, err.Error()))
		}
		_, err = buffer.ReadFrom(file)
		if err != nil {
			panic(fmt.Sprintf("Failed to read contents of template file '%s', error %s", filename, err.Error()))
		}
		s := buffer.String()
		name := filepath.Base(filename)
		var tmpl *template.Template
		if t == nil {
			// first file in the list is used as the root template
			t = template.New(name)
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		_, err = tmpl.Parse(s)
		if err != nil {
			panic(fmt.Sprintf("Error parsing template %s, error %s", filename, err.Error()))
		}
	}
	return t
}

// GetTemplateList returns all templates in view hierarchy in a BFS order list
// TODO: Implement this
func GetTemplateList(view *View) []string {
	return []string{view.Template}
}
