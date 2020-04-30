package ticket

import (
	"net/http"

	"github.com/rur/treetop"
	"github.com/rur/treetop/demo/assets"
	"github.com/rur/treetop/demo/ticket/handlers"
)

// Setup register routes for /view demo endpoint
func Setup(mux *http.ServeMux, devMode bool) {
	// developer executor will force templates to be reloaded from disk for
	// every request. Also template errors will be rendered in the browser
	var exec treetop.ViewExecutor = &treetop.FileExecutor{
		KeyedString: map[string]string{
			"local://base.html": assets.BaseHTML,
			"local://nav.html":  assets.NavHTML(assets.TicketsNav),
		},
	}
	if devMode {
		// Use developer executor to permit template file editing
		exec = &treetop.DeveloperExecutor{
			ViewExecutor: exec,
		}
	}

	// construct views and register routes given an executor
	Routes(Mux{ServeMux: mux}, exec)

	// add endpoint for regular http request handler
	mux.HandleFunc("/ticket/get-form", handlers.FormDepartmentRedirectHandler)

	if errs := exec.FlushErrors(); len(errs) != 0 {
		panic(errs.Error())
	}
}

// Mux type adds a HandleGET and HandlePOST method to the standard http mux
type Mux struct {
	*http.ServeMux
}

// HandleGET register a handler tha twill accept on GET (or HEAD) requests
// or write a 501 NotImplemented error
func (m *Mux) HandleGET(path string, handler http.Handler) {
	m.ServeMux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "GET", "HEAD":
			handler.ServeHTTP(w, req)
		default:
			http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		}
	})
}

// HandlePOST register a handler tha twill accept on POST requests
// or write a 501 NotImplemented error
func (m *Mux) HandlePOST(path string, handler http.Handler) {
	m.ServeMux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "POST":
			handler.ServeHTTP(w, req)
		default:
			http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		}
	})
}
