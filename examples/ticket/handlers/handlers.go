package handlers

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/rur/treetop"
)

var (
	wsRegex = regexp.MustCompile(`\s+`)
)

// -------------------------
// ticket Block Handlers
// -------------------------

// ticketHandler (default partial)
// Extends: content
// Method: GET
// Doc: Landing page for ticket wizard
func TicketHandler(rsp treetop.Response, req *http.Request) interface{} {
	query := req.URL.Query()
	data := struct {
		Summary string
		Dept    string
		Form    interface{}
		Notes   interface{}
	}{
		Summary: strings.TrimSpace(wsRegex.ReplaceAllString(query.Get("summary"), " ")),
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
		FormDepartmentRedirectHandler(rsp, req)
		return nil
	}
	return data
}

// formDepartmentRedirectHandler will issue a redirect to the correct form path based upon the value
// of the department query parameter. If not recognized it directs browser to ticket landing page.
func FormDepartmentRedirectHandler(w http.ResponseWriter, req *http.Request) {
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

// issuePreview (partial)
// Extends: content
// Method: GET
// Doc: Content wrapper around preview for different ticket type
func IssuePreviewHandler(rsp treetop.Response, req *http.Request) interface{} {
	data := struct {
		Preview interface{}
		Notes   interface{}
	}{
		Preview: rsp.HandleSubView("preview", req),
		Notes:   rsp.HandleSubView("notes", req),
	}
	return data
}
