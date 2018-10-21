package treetop

import (
	"html/template"
	"io"
	"path/filepath"
)

// implements TemplateExec
func DefaultTemplateExec(w io.Writer, templates []string, data interface{}) error {
	t, err := template.New("__init__").ParseFiles(templates...)
	if err != nil {
		return err
	}
	if err := t.ExecuteTemplate(w, filepath.Base(templates[0]), data); err != nil {
		return err
	}
	return nil
}
