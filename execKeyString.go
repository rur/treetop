package treetop

import (
	"errors"
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
func (ex *KeyedStringExecutor) NewViewHandler(view *View, includes ...*View) ViewHandler {
	ex.exec.NewTemplate = ex.constructTemplate
	return ex.exec.NewViewHandler(view, includes...)
}

// FlushErrors will return a list of all template generation errors that occurred
// while ViewHandlers were being created by this executor
func (ex *KeyedStringExecutor) FlushErrors() []*ExecutorError {
	return ex.exec.FlushErrors()
}

var errEmptyViewQueue = errors.New("empty view queue")

// viewQueue simple queue implementation used for breath first traversal
type viewQueue struct {
	offset int
	items  []*View
}

func (q *viewQueue) join(v *View) {
	q.items = append(q.items, v)
}

func (q *viewQueue) next() (*View, error) {
	if q.empty() {
		return nil, errEmptyViewQueue
	}
	next := q.items[q.offset]
	q.offset++
	return next, nil
}

func (q *viewQueue) empty() bool {
	return q.offset >= len(q.items)
}

// constructTemplates will traverse the supplied view hierarchy in breath first order and
// use items from the parse tree reference to construct a template namespace
func (kse *KeyedStringExecutor) constructTemplate(view *View) (*template.Template, error) {
	if view == nil {
		return nil, nil
	}
	out := template.New(view.Defines)

	queue := viewQueue{}
	queue.join(view)

	for !queue.empty() {
		v, _ := queue.next()
		tree, ok := kse.parsed[v.Template]
		if !ok {
			return nil, fmt.Errorf(
				"KeyedStringExecutor: no template found for key '%s'",
				v.Template)
		}
		_, err := out.AddParseTree(v.Defines, tree)
		if err != nil {
			return nil, err
		}
		for _, sub := range v.SubViews {
			queue.join(sub)
		}
	}
	return out, nil
}
