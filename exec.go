package treetop

import (
	"html/template"
)

// Executor encapsulates a procedure for converting an endpoint definition
// into a request handler using golang HTML templates.
//
// It should be extended, so to speak, with an implementation for
// constructing the actual template instances, given a View, potentially with a hierarchy
// of template handlers.
type Executor struct {
	NewTemplate func(*View) (*template.Template, error)
	Errors      []*ExecutorError
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

// FlushErrors will return a list of all template generation errors that occurred
// while TemplateHandlers were being created
func (ex *Executor) FlushErrors() []*ExecutorError {
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
		IncludeTemplates: make([]*template.Template, len(incls)),
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
