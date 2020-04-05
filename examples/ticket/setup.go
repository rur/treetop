package ticket

import (
	"net/http"
	"net/url"

	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/assets"
)

// Setup register routes for /view example endpoint
func Setup(mux *http.ServeMux) {
	// developer executor will force templates to be reloaded from disk for
	// every request. Also template errors will be rendered in the browser
	exec := &treetop.DeveloperExecutor{
		ViewExecutor: &treetop.FileExecutor{
			KeyedString: map[string]string{
				"local://base.html": assets.BaseHTML,
				"local://nav.html":  assets.NavHTML(assets.TicketsNav),
			},
		},
	}

	// construct views and register routes given an executor
	Routes(Mux{ServeMux: mux}, exec)

	// add endpoint for regular http request handler
	mux.HandleFunc("/ticket/get-form", formDepartmentRedirectHandler)

	if errs := exec.FlushErrors(); len(errs) != 0 {
		panic(errs.Error())
	}
}

// formDepartmentRedirectHandler will issue a redirect to the correct form path based upon the value
// of the department query parameter. If not recognized it directs browser to ticket landing page.
func formDepartmentRedirectHandler(w http.ResponseWriter, req *http.Request) {
	var (
		redirect *url.URL
		query    = req.URL.Query()
	)
	switch dpt := query.Get("department"); dpt {
	case "helpdesk":
		redirect = mustParseURL("/ticket/helpdesk/new")

	case "software":
		redirect = mustParseURL("/ticket/software/new")

	case "systems":
		redirect = mustParseURL("/ticket/systems/new")

	default:
		query.Del("department")
		redirect = mustParseURL("/ticket")
	}

	redirect.RawQuery = query.Encode()

	http.Redirect(w, req, redirect.String(), http.StatusSeeOther)
}

// for use with hard coded urls
func mustParseURL(path string) *url.URL {
	u, err := url.Parse(path)
	if err != nil {
		panic(err)
	}
	return u
}

// Mux type adds a HandleGET and HandlePOST method to the standard http mux
type Mux struct {
	*http.ServeMux
}

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
