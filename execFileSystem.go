package treetop

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"text/template/parse"
)

// FileSystemExecutor loads view templates as a path from a Go HTML template file.
// The underlying file system is abstracted through the http.FileSystem interface to allow for
// in-memory use.
type FileSystemExecutor struct {
	FS    http.FileSystem
	Funcs template.FuncMap
	exec  Executor
}

// NewViewHandler creates a ViewHandler from a View endpoint definition treating
// view template strings as keys into the string template dictionary.
func (fe *FileSystemExecutor) NewViewHandler(view *View, includes ...*View) ViewHandler {
	fe.exec.NewTemplate = fe.constructTemplate
	return fe.exec.NewViewHandler(view, includes...)
}

// FlushErrors will return a list of all template generation errors that occurred
// while ViewHandlers were being created by this executor
func (fe *FileSystemExecutor) FlushErrors() ExecutorErrors {
	return fe.exec.FlushErrors()
}

// constructTempalate for FileSystemExecutor will treat the template string of each view
// as a template in of itself.
func (fe *FileSystemExecutor) constructTemplate(view *View) (*template.Template, error) {
	if view == nil {
		return nil, nil
	}
	var out *template.Template
	cache := make(map[string]*parse.Tree)
	buffer := new(bytes.Buffer)

	queue := viewQueue{}
	queue.add(view)

	for !queue.empty() {
		v, _ := queue.next()
		var t *template.Template
		if out == nil {
			out = template.New(v.Defines).Funcs(fe.Funcs)
			t = out
		} else {
			t = out.New(v.Defines)
		}

		if tree, ok := cache[v.Template]; ok {
			t.AddParseTree(v.Defines, tree)
		} else {
			buffer.Reset()
			file, err := fe.FS.Open(v.Template)
			if err != nil {
				return nil, fmt.Errorf(
					"Failed to open template file '%s', error %s",
					v.Template, err.Error(),
				)
			}
			_, err = buffer.ReadFrom(file)
			if err != nil {
				return nil, fmt.Errorf(
					"Failed to read contents of template file '%s', error %s",
					v.Template, err.Error(),
				)
			}
			_, err = t.Parse(buffer.String())
			if err != nil {
				return nil, fmt.Errorf(
					"Failed to parse contents of template file '%s', error %s",
					v.Template, err.Error(),
				)
			}
		}
		for _, sub := range v.SubViews {
			if sub != nil {
				queue.add(sub)
			}
		}
	}
	return out, nil
}
