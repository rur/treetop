package treetop

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

// FileSystemExecutor loads view templates as a path from a Go HTML template file.
// The underlying file system is abstracted through the http.FileSystem interface to allow for
// in-memory use.
type FileSystemExecutor struct {
	FS http.FileSystem
}

// NewViewHandler will create a Handler instance capable of serving treetop requests
// for the supplied view configuration
//
// TODO: Implement this
func (ft *FileSystemExecutor) NewViewHandler(view *View, includes ...*View) ViewHandler {
	page, part, incls := CompileViews(view, includes...)
	handler := &TemplateHandler{
		Page:            page,
		PageTemplate:    ft.MustParseTemplateFiles(page),
		Partial:         part,
		PartialTemplate: ft.MustParseTemplateFiles(part),
		Includes:        incls,
	}
	for _, inc := range incls {
		handler.IncludeTemplates = append(handler.IncludeTemplates, ft.MustParseTemplateFiles(inc))
	}
	return handler
}

// MustParseTemplateFiles will load template files and parse contents into a HTML template instance
// it will use the supplied http.FileSystem for loading the template files
//
// TODO: Implement this
func (ft *FileSystemExecutor) MustParseTemplateFiles(view *View) *template.Template {
	var t *template.Template
	// snippet based upon https://golang.org/pkg/html/template/#ParseFiles implementation
	for _, filename := range GetTemplateList(view) {
		buffer := new(bytes.Buffer)
		file, err := ft.FS.Open(filename)
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
