package treetop

import (
	"html/template"
	"net/http"
)

// DeveloperExecutor wraps another executor, it will re-generate a new view handler for
// every request. This can be used to live-reload templates during development.
//
// Note: this is for development use only, it is not suitable for production systems
type DeveloperExecutor struct {
	ViewExecutor
}

// NewViewHandler will create a special handler that will reload the templates
// for ever request. Any template errors that occur will be rendered to the client.
//
// Note: this is for development use only, it is not suitable for production systems
func (de *DeveloperExecutor) NewViewHandler(view *View, includes ...*View) ViewHandler {
	// dry run to capture errors up front
	_ = de.ViewExecutor.NewViewHandler(view, includes...)
	return &devHandler{
		view: view,
		incl: includes,
		exec: de.ViewExecutor,
	}
}

// devHandler will load and execute the view template for every request.
type devHandler struct {
	pageOnly     bool
	templateOnly bool
	view         *View
	incl         []*View
	exec         ViewExecutor
}

// FragmentOnly creates a new Handler that only responds to fragment requests
func (h *devHandler) FragmentOnly() ViewHandler {
	return &devHandler{
		templateOnly: true,
		pageOnly:     h.pageOnly,
		view:         h.view,
		incl:         h.incl,
		exec:         h.exec,
	}
}

// PageOnly create a new handler that will only respond to non-fragment (full page) requests
func (h *devHandler) PageOnly() ViewHandler {
	return &devHandler{
		pageOnly:     true,
		templateOnly: h.templateOnly,
		view:         h.view,
		incl:         h.incl,
		exec:         h.exec,
	}
}

// ServeHTTP for the development handler will generate a new ViewHandler from the
// executor and the view definitions for each request.
// If an error occurs a HTML page will be rendered with the details for debug purposes.
//
// NOTE: This is intended for development, it is not suitable for production use.
func (h *devHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler := h.exec.NewViewHandler(h.view, h.incl...)
	errs := h.exec.FlushErrors()
	if len(errs) > 0 {
		if err := writeDebugErrorPage(w, handler, errs); err != nil {
			panic(err)
		}
		return
	}
	if h.pageOnly {
		handler = handler.PageOnly()
	}
	if h.templateOnly {
		handler = handler.FragmentOnly()
	}

	if th, ok := handler.(*TemplateHandler); ok && th.ServeTemplateError == nil {
		th.ServeTemplateError = func(err error, resp Response, req *http.Request) {
			if err := writeDebugErrorPage(w, handler, err); err != nil {
				panic(err)
			}
		}
	}
	handler.ServeHTTP(w, req)
}

// devErrorTemplate is used to draw debug pages when a template error is
// encountered in a devHandler endpoint
var devErrorTemplate *template.Template

func init() {
	devErrorTemplate = template.Must(template.New("dev_error").Parse(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
		<title>Treetop Template Error</title>
		<style type="text/css" media="screen">
		
			body {
				line-height: 140%;
				margin: 50px;
				width: 650px;
			}
			code {font-size: 120%;}
			
			
			pre code {
				background-color: #eee;
				border: 1px solid #999;
				display: block;
				padding: 20px;
			}
			
		</style>
	</head>
	<body>
		<h1>Treetop Endpoint Error</h1>
		<h3>Errors:</h3>
		<pre>
			<code>
				{{ .Output }}
			</code>
		</pre>
		
		{{ if .PageView }}
		<h3>Page View:</h3>
		<pre>
			<code>
				{{ .PageView }}
			</code>
		</pre>
		{{ end }}
	
		{{ if .TemplateView }}
		<h3>Template View:</h3>
		<pre>
			<code>
				{{ .TemplateView }}
			</code>
		</pre>
		{{ end }}
	
		{{ range $index, $ps := .Includes }}
		<h3>Postscript[{{ $index }}]:</h3>
		<pre>
			<code>
				{{ $ps }}
			</code>
		</pre>
		{{ end }}
	</body>
	</html>
	`))
}

// writeDebugErrorPage will write a HTML page will information about the error
// and whatever it can get about the handler endpoint
func writeDebugErrorPage(w http.ResponseWriter, handler ViewHandler, err error) error {
	errData := struct {
		Output       string
		PageView     string
		TemplateView string
		Includes     []string
	}{
		Output: err.Error(),
	}
	if th, ok := handler.(*TemplateHandler); ok {
		errData.PageView = SprintViewTree(th.Page)
		errData.TemplateView = SprintViewTree(th.Partial)
		for _, incl := range th.Includes {
			errData.Includes = append(errData.Includes, SprintViewTree(incl))
		}
	}
	w.WriteHeader(http.StatusInternalServerError)
	return devErrorTemplate.Execute(w, errData)
}
