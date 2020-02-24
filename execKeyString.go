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
	parsed map[string]*parse.Tree
	errors []*ExecutorTemplateError
}

// ExecutorTemplateError is created within the executor when a template cannot be created
// for a view. Call exec.FlushErrors() to obtain a list of the template errors that occurred.
type ExecutorTemplateError struct {
	View *View
	Err  error
}

// Error implement the error interface
func (te *ExecutorTemplateError) Error() string {
	if te == nil {
		return "nil"
	}
	return fmt.Sprintln(SprintViewInfo(te.View), ":", te.Err)
}

// NewKeyedStringExecutor will parse the templates supplied and
// return an Executor capable of building TemplateHander instances
// using the parsed trees to construct templates.
// If parsing fails for any of the templates the error will be returned
func NewKeyedStringExecutor(templates map[string]string) (*KeyedStringExecutor, error) {
	exec := KeyedStringExecutor{
		parsed: make(map[string]*parse.Tree),
	}

	for key, str := range templates {
		t, err := template.New("temp").Parse(str)
		if err != nil {
			return nil, err
		}
		exec.parsed[key] = t.Tree
	}

	return &exec, nil
}

// NewViewHandler implements the Executor interface capable of converting view definitions
// into a TemplateHandler
func (kse *KeyedStringExecutor) NewViewHandler(view *View, includes ...*View) ViewHandler {
	page, part, incls := CompileViews(view, includes...)

	handler := &TemplateHandler{
		Page:             page,
		Partial:          part,
		Includes:         incls,
		IncludeTemplates: make([]*template.Template, len(incls)),
	}

	if t, err := kse.constructTemplate(page); err != nil {
		kse.errors = append(kse.errors, &ExecutorTemplateError{
			View: page,
			Err:  err,
		})
		// this handler will not accept page requests
		handler.Page = nil
	} else {
		handler.PageTemplate = t
	}

	if t, err := kse.constructTemplate(part); err != nil {
		kse.errors = append(kse.errors, &ExecutorTemplateError{
			View: part,
			Err:  err,
		})
		// error has been captured, disable partial handling
		handler.Partial = nil
	} else {
		handler.PartialTemplate = t
	}

	for i, inc := range incls {
		if t, err := kse.constructTemplate(inc); err != nil {
			kse.errors = append(kse.errors, &ExecutorTemplateError{
				View: inc,
				Err:  err,
			})
			// error has been captured, disable partial handing
			handler.Partial = nil
		} else {
			handler.IncludeTemplates[i] = t
		}
	}

	// normalize output
	var viewHandler ViewHandler
	viewHandler = handler
	if handler.Partial == nil {
		viewHandler = handler.PageOnly()
	}
	if handler.Page == nil {
		viewHandler = handler.FragmentOnly()
	}
	return viewHandler
}

// FlushErrors will return a list of all template generation errors that occurred
// while TemplateHandlers were being created
func (kse *KeyedStringExecutor) FlushErrors() []*ExecutorTemplateError {
	errors := kse.errors
	kse.errors = nil
	if len(errors) == 0 {
		return nil
	}
	return errors
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
