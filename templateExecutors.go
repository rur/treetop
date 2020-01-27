package treetop

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
)

// DefaultTemplateExec implements TemplateExec as a thin wrapper around the
// ParseFiles method in html/template package
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

// StringTemplateExec will parse template strings as the template body, not the path
// to a template file on disk
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

// TemplateFileSystem is similar to default executor except template files will be loading using an interface
// see https://golang.org/pkg/net/http/#FileSystem
func TemplateFileSystem(fs http.FileSystem) TemplateExec {
	return func(w io.Writer, templates []string, data interface{}) error {
		var t *template.Template
		// snippet based upon https://golang.org/pkg/html/template/#ParseFiles implementation
		for _, filename := range templates {
			buffer := new(bytes.Buffer)
			file, err := fs.Open(filename)
			if err != nil {
				return fmt.Errorf("Failed to open template file '%s', error %s", filename, err.Error())
			}
			_, err = buffer.ReadFrom(file)
			if err != nil {
				return fmt.Errorf("Failed to read contents of template file '%s', error %s", filename, err.Error())
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
				return fmt.Errorf("Error parsing template %s, error %s", filename, err.Error())
			}
		}

		if err := t.ExecuteTemplate(w, t.Name(), data); err != nil {
			return err
		}
		return nil
	}
}
