package ticket

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/assets"
)

// Routes register routes for /view example endpoint
func Routes(mux *http.ServeMux) {
	page := treetop.NewView("local://base.html", treetop.Delegate("content"))
	_ = page.NewDefaultSubView("nav", "local://nav.html", treetop.Noop)
	content := page.NewSubView("content", "examples/ticket/templates/content.html.tmpl", ticketContentHandler)
	helpdesk := content.NewDefaultSubView("form", "examples/ticket/templates/helpdesk.html.tmpl", formHandler)
	software := content.NewSubView("form", "examples/ticket/templates/software.html.tmpl", formHandler)
	systems := content.NewSubView("form", "examples/ticket/templates/systems.html.tmpl", formHandler)

	var exec treetop.ViewExecutor
	exec = &treetop.FileExecutor{
		KeyedString: map[string]string{
			"local://base.html": assets.BaseHTML,
			"local://nav.html":  assets.NavHTML(assets.TicketsNav),
		},
	}

	exec = &treetop.DeveloperExecutor{ViewExecutor: exec}

	mux.Handle("/ticket", exec.NewViewHandler(content).PageOnly())
	mux.HandleFunc("/ticket/get-form", getFormHandler)
	mux.Handle("/ticket/helpdesk/new", exec.NewViewHandler(helpdesk))
	mux.Handle("/ticket/software/new", exec.NewViewHandler(software))
	mux.Handle("/ticket/systems/new", exec.NewViewHandler(systems))

	if errs := exec.FlushErrors(); len(errs) != 0 {
		panic(errs.Error())
	}
}

// ticketContentHandler
// extends: base.html{content}
func ticketContentHandler(rsp treetop.Response, req *http.Request) interface{} {
	data := struct {
		Form interface{}
	}{
		Form: rsp.HandleSubView("form", req),
	}
	return data
}

// formHandler
// extends: content.html{form}
func formHandler(rsp treetop.Response, req *http.Request) interface{} {
	if treetop.IsTemplateRequest(req) {
		// replace existing browser history entry with current URL
		rsp.ReplacePageURL(req.URL.String())
	}

	// TODO: Implement this
	return nil
}

func getFormHandler(w http.ResponseWriter, req *http.Request) {
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
		http.Error(w, fmt.Sprintf(`Unknown department: %s`, dpt), http.StatusBadRequest)
	}

	redirect.RawQuery = query.Encode()

	http.Redirect(w, req, redirect.String(), http.StatusSeeOther)
}

func mustParseURL(path string) *url.URL {
	u, err := url.Parse(path)
	if err != nil {
		panic(err)
	}
	return u
}
