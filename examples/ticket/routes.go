package ticket

import (
	"net/http"
	"net/url"

	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/assets"
)

// Routes register routes for /view example endpoint
func Routes(mux *http.ServeMux) {
	// components
	findUser := treetop.NewSubView("find-user", "examples/ticket/templates/components/findUser.html.tmpl", findUsersHandler)
	assignee := treetop.NewSubView("assignee", "examples/ticket/templates/components/assignee.html.tmpl", assigneeHandler)

	// page views
	page := treetop.NewView("local://base.html", treetop.Delegate("content"))
	_ = page.NewDefaultSubView("nav", "local://nav.html", treetop.Noop)
	content := page.NewSubView("content", "examples/ticket/templates/content.html.tmpl", ticketContentHandler)

	helpdesk := content.NewSubView("form", "examples/ticket/templates/helpdesk.html.tmpl", formHandler)
	helpdesk.SubViews["assignee"] = assignee // use assignee view for helpdesk form

	software := content.NewSubView("form", "examples/ticket/templates/software.html.tmpl", formHandler)
	software.SubViews["assignee"] = assignee // use assignee view for software form

	systems := content.NewSubView("form", "examples/ticket/templates/systems.html.tmpl", formHandler)
	systems.SubViews["assignee"] = assignee // use assignee view for systems form

	var exec treetop.ViewExecutor
	exec = &treetop.FileExecutor{
		KeyedString: map[string]string{
			"local://base.html": assets.BaseHTML,
			"local://nav.html":  assets.NavHTML(assets.TicketsNav),
		},
	}

	// developer executor will force templates to be reloaded from disk for
	// every request. Also template errors will be rendered in the browser
	exec = &treetop.DeveloperExecutor{ViewExecutor: exec}

	// demo page entry point
	mux.Handle("/ticket", exec.NewViewHandler(content).PageOnly())

	// forms entry point
	mux.Handle("/ticket/helpdesk/new", exec.NewViewHandler(helpdesk))
	mux.Handle("/ticket/software/new", exec.NewViewHandler(software))
	mux.Handle("/ticket/systems/new", exec.NewViewHandler(systems))

	// form update handlers
	mux.HandleFunc("/ticket/get-form", formDepartmentRedirectHandler)
	mux.Handle("/ticket/get-assignee", exec.NewViewHandler(assignee).FragmentOnly())
	mux.Handle("/ticket/find-user", exec.NewViewHandler(findUser).FragmentOnly())

	if errs := exec.FlushErrors(); len(errs) != 0 {
		panic(errs.Error())
	}
}

// ticketContentHandler
// extends: base.html{content}
func ticketContentHandler(rsp treetop.Response, req *http.Request) interface{} {
	query := req.URL.Query()
	data := struct {
		Summary string
		Dept    string
		Form    interface{}
	}{
		Summary: sanitizeSummary(query.Get("summary")),
		Form:    rsp.HandleSubView("form", req),
	}
	// validate department and redirect if necessary
	switch d := query.Get("department"); d {
	case "helpdesk", "software", "systems":
		// form redirect handler
		data.Dept = d
	}
	if (data.Dept == "" && req.URL.Path != "/ticket") || (data.Dept != "" && req.URL.Path == "/ticket") {
		// url does not match the department value, redirect
		formDepartmentRedirectHandler(rsp, req)
		return nil
	}
	return data
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
