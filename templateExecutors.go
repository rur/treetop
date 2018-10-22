package treetop

import (
	"fmt"
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

func StringTemplateExec(w io.Writer, templates []string, data interface{}) error {
	var t *template.Template
	// snippet based upon https://golang.org/pkg/html/template/#ParseFiles implementation
	for i, s := range templates {
		name := fmt.Sprintf("Template[%v]", i)
		var tpls *template.Template
		if t == nil {
			// first file in the list is used as the root template
			t = template.New(name)
			tpls = t
		} else {
			tpls = t.New(name)
		}
		_, err := tpls.Parse(s)
		if err != nil {
			return fmt.Errorf("Error parsing template %s, error %s", name, err.Error())
		}
	}

	if err := t.ExecuteTemplate(w, t.Name(), data); err != nil {
		return err
	}
	return nil
}
