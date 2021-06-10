package treetop

import (
	"errors"
	"fmt"
	"html/template"
	"sort"
	"strconv"
	"strings"
	"text/template/parse"
)

// ViewExecutor is an interface for objects that implement transforming a View definition
// into a ViewHandler that supports full page and template requests.
type ViewExecutor interface {
	NewViewHandler(view *View, includes ...*View) ViewHandler
	FlushErrors() ExecutorErrors
}

// ExecutorErrors is a list zero or more template errors created when parsing
// templates
type ExecutorErrors []*ExecutorError

// Errors implements error interface
func (ee ExecutorErrors) Error() string {
	var output string
	for i := range ee {
		output += ee[i].Error() + "\n"
	}
	return output
}

// ExecutorError is created within the executor when a template cannot be created
// for a view. Call exec.FlushErrors() to obtain a list of the template errors that occurred.
type ExecutorError struct {
	View *View
	Err  error
}

// Error implement the error interface
func (te *ExecutorError) Error() string {
	if te == nil {
		return "nil"
	}
	return te.Err.Error()
}

// Executor implements a procedure for converting a view endpoint definition
// into a request handler using HTML templates.
//
// It is designed to be extended with different means for creating template instances
// given a View instance.
//
// Example:
//
// 		exec := Executor{
//			NewTemplate: func(v *View) (*template.Template, error){
//				// always hello
//				return template.Must(template.New(v.Defines).Parse("Hello!"))
// 			},
// 		}
// 		mux.Handle("/hello", exec.NewViewHandler(v))
//
type Executor struct {
	NewTemplate func(*View) (*template.Template, error)
	Errors      ExecutorErrors
}

// FlushErrors will return the list of template creation errors that occurred
// while ViewHandlers were begin created, since the last time it was called.
func (ex *Executor) FlushErrors() ExecutorErrors {
	errors := ex.Errors
	ex.Errors = nil
	if len(errors) == 0 {
		return nil
	}
	return errors
}

// NewViewHandler implements the Executor interface capable of creating a ViewHandler
// from a View endpoint definition.
func (ex *Executor) NewViewHandler(view *View, includes ...*View) ViewHandler {
	page, part, incls := CompileViews(view, includes...)

	handler := &TemplateHandler{
		Page:             page,
		Partial:          part,
		Includes:         incls,
		IncludeTemplates: make([]Template, len(incls)),
	}

	if t, err := ex.NewTemplate(page); err != nil {
		ex.Errors = append(ex.Errors, &ExecutorError{
			View: page,
			Err:  err,
		})
		// this handler will not accept page requests
		handler.Page = nil
	} else {
		handler.PageTemplate = t
	}

	if t, err := ex.NewTemplate(part); err != nil {
		ex.Errors = append(ex.Errors, &ExecutorError{
			View: part,
			Err:  err,
		})
		// error has been captured, disable partial handling
		handler.Partial = nil
	} else {
		handler.PartialTemplate = t
	}

	for i, inc := range incls {
		if t, err := ex.NewTemplate(inc); err != nil {
			ex.Errors = append(ex.Errors, &ExecutorError{
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

// utilities ---

var errEmptyViewQueue = errors.New("empty view queue")

// viewQueue simple queue implementation used for breath first traversal
//
// NB: this is only suitable for localized short-lived queues since the underlying
// array will not deallocate pointers
type viewQueue struct {
	offset int
	items  []*View
}

func (q *viewQueue) add(v *View) {
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

// checkTemplateForBlockNames will scan the parsed templates for blocks/template slots
// that match the declared block names. If a block naming is not present, return an error
func checkTemplateForBlockNames(tmpl *template.Template, v *View) error {
	parsedBlocks := make(map[string]bool)
	for _, tplName := range listTemplateNodeName(tmpl.Tree.Root) {
		parsedBlocks[tplName] = true
	}

	var missing []string
	for blockName := range v.SubViews {
		if _, ok := parsedBlocks[blockName]; !ok {
			missing = append(missing, strconv.Quote(blockName))
		}
	}
	if len(missing) == 0 {
		return nil
	}
	sort.Strings(missing)
	return fmt.Errorf("%s is missing template declaration(s) for sub view blocks: %s", v.Template, strings.Join(missing, ", "))
}

// listTemplateNodeName will scan a parsed template tree for template nodes
// and list all template names found
func listTemplateNodeName(list *parse.ListNode) (names []string) {
	if list == nil {
		return
	}
	for _, node := range list.Nodes {
		switch n := node.(type) {
		case *parse.TemplateNode:
			names = append(names, n.Name)
		case *parse.IfNode:
			names = append(names, listTemplateNodeName(n.List)...)
			names = append(names, listTemplateNodeName(n.ElseList)...)
		case *parse.RangeNode:
			names = append(names, listTemplateNodeName(n.List)...)
			names = append(names, listTemplateNodeName(n.ElseList)...)
		case *parse.WithNode:
			names = append(names, listTemplateNodeName(n.List)...)
			names = append(names, listTemplateNodeName(n.ElseList)...)
		case *parse.ListNode:
			names = append(names, listTemplateNodeName(n)...)
		}
	}
	return
}
