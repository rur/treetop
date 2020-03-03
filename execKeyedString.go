package treetop

import (
	"fmt"
	"html/template"
	"text/template/parse"
)

// KeyedStringExecutor builds handlers templates from a map
// of available templates. The view templates are treated as
// keys into the map for the purpose of build handlers.
type KeyedStringExecutor struct {
	exec   Executor
	parsed map[string]*parse.Tree
}

// NewKeyedStringExecutor will parse the templates supplied and
// return an Executor capable of building TemplateHander instances
// using the parsed trees to construct templates.
// If parsing fails for any of the templates the error will be returned
func NewKeyedStringExecutor(templates map[string]string) (*KeyedStringExecutor, error) {
	exec := &KeyedStringExecutor{
		parsed: make(map[string]*parse.Tree),
	}

	for key, str := range templates {
		t, err := template.New(key).Parse(str)
		if err != nil {
			return nil, err
		}
		exec.parsed[key] = t.Tree
	}

	return exec, nil
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
	out := template.New(view.Defines)

	queue := viewQueue{}
	queue.add(view)

	for !queue.empty() {
		v, err := queue.next()
		if err != nil {
			return nil, err
		}
		tree, ok := kse.parsed[v.Template]
		if !ok {
			return nil, fmt.Errorf(
				"KeyedStringExecutor: no template found for key '%s'",
				v.Template)
		}
		_, err = out.AddParseTree(v.Defines, tree)
		if err != nil {
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
