package treetop

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
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

// CaptureErrors is a base type for implementing concreate view executors
type CaptureErrors struct {
	Errors ExecutorErrors
}

// FlushErrors will return the list of template creation errors that occurred
// while ViewHandlers were begin created, since the last time it was called.
func (ce *CaptureErrors) FlushErrors() ExecutorErrors {
	errors := ce.Errors
	ce.Errors = nil
	if len(errors) == 0 {
		return nil
	}
	return errors
}

// AddErrors will store a list of errors to flushe later
func (ce *CaptureErrors) AddErrors(errs ExecutorErrors) {
	ce.Errors = append(ce.Errors, errs...)
}

// StringExecutor loads view templates as an inline template string.
//
// Example:
//
// 		exec := StringExecutor{}
// 		v := treetop.NewView("<p>Hello {{ . }}!</p>", Constant("world"))
// 		mux.Handle("/hello", exec.NewViewHandler(v))
//
type StringExecutor struct {
	CaptureErrors
	Funcs template.FuncMap
}

// NewViewHandler creates a ViewHandler from a View endpoint definition treating
// view template strings as keys into the string template dictionary.
func (se *StringExecutor) NewViewHandler(view *View, includes ...*View) ViewHandler {
	loader := NewTemplateLoader(se.Funcs, func(tmpl string) (string, error) {
		return tmpl, nil
	})
	handler, errs := NewTemplateHandler(view, includes, loader)
	se.AddErrors(errs)
	return handler
}

// KeyedStringExecutor builds handlers templates from a map
// of available templates. The view templates are treated as
// keys into the map for the purpose of build handlers.
type KeyedStringExecutor struct {
	CaptureErrors
	Templates map[string]string
	Funcs     template.FuncMap
}

// NewKeyedStringExecutor is a deprecated method for constructing an
// instance from a template map
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
func (ks *KeyedStringExecutor) NewViewHandler(view *View, includes ...*View) ViewHandler {
	loader := NewTemplateLoader(ks.Funcs, func(key string) (string, error) {
		tmpl, ok := ks.Templates[key]
		if !ok {
			return "", fmt.Errorf("no key found for template '%s'", key)
		}
		return tmpl, nil
	})
	handler, errs := NewTemplateHandler(view, includes, loader)
	ks.AddErrors(errs)
	return handler
}

// FileExecutor loads view templates as a path from a template file.
type FileExecutor struct {
	CaptureErrors
	Funcs       template.FuncMap
	KeyedString map[string]string
}

// NewViewHandler creates a ViewHandler from a View endpoint definition treating
// view template strings as a file path using os.Open.
func (fe *FileExecutor) NewViewHandler(view *View, includes ...*View) ViewHandler {
	loader := NewTemplateLoader(fe.Funcs, func(name string) (string, error) {
		if len(fe.KeyedString) > 0 {
			tmpl, ok := fe.KeyedString[name]
			if ok {
				return tmpl, nil
			}
		}
		file, err := os.Open(name)
		if err != nil {
			return "", err
		}
		defer file.Close()
		tpl, err := ioutil.ReadAll(file)
		if err != nil {
			return "", err
		}
		return string(tpl), nil
	})
	handler, errs := NewTemplateHandler(view, includes, loader)
	fe.AddErrors(errs)
	return handler
}

// FileSystemExecutor loads view templates as a path from a Go HTML template file.
// The underlying file system is abstracted through the http.FileSystem interface to allow for
// in-memory use.
//
// The optional KeyedString map will be checked before the loader attempts to use the FS
// instance when obtain a template string
type FileSystemExecutor struct {
	CaptureErrors
	FS          http.FileSystem
	Funcs       template.FuncMap
	KeyedString map[string]string
}

// NewViewHandler creates a ViewHandler from a View endpoint definition treating
// view template strings as keys into the string template dictionary.
func (fse *FileSystemExecutor) NewViewHandler(view *View, includes ...*View) ViewHandler {
	loader := NewTemplateLoader(fse.Funcs, func(name string) (string, error) {
		if len(fse.KeyedString) > 0 {
			tmpl, ok := fse.KeyedString[name]
			if ok {
				return tmpl, nil
			}
		}
		file, err := fse.FS.Open(name)
		if err != nil {
			return "", fmt.Errorf(
				"failed to open template file '%s', error %s",
				name, err.Error(),
			)
		}
		defer file.Close()
		tpl, err := ioutil.ReadAll(file)
		if err != nil {
			return "", fmt.Errorf(
				"failed to open template file '%s', error %s",
				name, err.Error(),
			)
		}
		return string(tpl), nil
	})
	handler, errs := NewTemplateHandler(view, includes, loader)
	fse.AddErrors(errs)
	return handler
}
