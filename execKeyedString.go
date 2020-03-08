package treetop

import (
	"fmt"
	"html/template"
)

// KeyedStringExecutor builds handlers templates from a map
// of available templates. The view templates are treated as
// keys into the map for the purpose of build handlers.
type KeyedStringExecutor struct {
	Templates map[string]string
	Funcs     template.FuncMap
	exec      Executor
}

// NewKeyedStringExecutor will parse the templates supplied and
// return an Executor capable of building TemplateHander instances
// using the parsed trees to construct templates.
// If parsing fails for any of the templates the error will be returned
func NewKeyedStringExecutor(templates map[string]string) *KeyedStringExecutor {
	exec := &KeyedStringExecutor{
		Templates: make(map[string]string),
	}
	for key, str := range templates {
		exec.Templates[key] = str
	}
	return exec
}

// NewViewHandler creates a ViewHandler from a View endpoint definition treating
// view template strings as keys into the string template dictionary.
func (kse *KeyedStringExecutor) NewViewHandler(view *View, includes ...*View) ViewHandler {
	kse.exec.NewTemplate = kse.constructTemplate
	return kse.exec.NewViewHandler(view, includes...)
}

// FlushErrors will return a list of all template generation errors that occurred
// while ViewHandlers were being created by this executor
func (kse *KeyedStringExecutor) FlushErrors() ExecutorErrors {
	return kse.exec.FlushErrors()
}

// constructTemplate will traverse the supplied view hierarchy in breath first order and
// use items from the parse tree reference to construct a template namespace
func (kse *KeyedStringExecutor) constructTemplate(view *View) (*template.Template, error) {
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
			out = template.New(v.Defines).Funcs(kse.Funcs)
			t = out
		} else {
			t = out.New(v.Defines)
		}
		templateStr, ok := kse.Templates[v.Template]
		if !ok {
			return nil, fmt.Errorf(
				"KeyedStringExecutor: no template found for key '%s'",
				v.Template)
		}
		if _, err := t.Parse(templateStr); err != nil {
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
