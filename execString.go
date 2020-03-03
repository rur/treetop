package treetop

import (
	"html/template"
)

// StringExecutor loads view templates as an inline template string.
//
// Example:
//
// 		exec := StringExecutor{}
// 		v := treetop.NewView("<p>Hello {{ . }}!</p>", Constant("world"))
// 		mux.Handle("/hello", exec.NewViewHandler(v))
//
type StringExecutor struct {
	exec Executor
}

// NewViewHandler creates a ViewHandler from a View endpoint definition treating
// view template strings as keys into the string template dictionary.
func (se *StringExecutor) NewViewHandler(view *View, includes ...*View) ViewHandler {
	se.exec.NewTemplate = se.constructTemplate
	return se.exec.NewViewHandler(view, includes...)
}

// FlushErrors will return a list of all template generation errors that occurred
// while ViewHandlers were being created by this executor
func (se *StringExecutor) FlushErrors() ExecutorErrors {
	return se.exec.FlushErrors()
}

// constructTempalate for StringExecutor will treat the template string of each view
// as a template in of itself.
func (se *StringExecutor) constructTemplate(view *View) (*template.Template, error) {
	if view == nil {
		return nil, nil
	}
	var out *template.Template

	queue := viewQueue{}
	queue.add(view)

	for !queue.empty() {
		v, _ := queue.next()
		var t *template.Template
		if out == nil {
			out = template.New(v.Defines)
			t = out
		} else {
			t = out.New(v.Defines)
		}
		if _, err := t.Parse(v.Template); err != nil {
			return nil, err
		}
		for _, sub := range v.SubViews {
			if sub != nil {
				queue.add(sub)
			}
		}
	}
	return out, nil
}
