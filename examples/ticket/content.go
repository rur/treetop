package ticket

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/rur/treetop"
)

// -------------------------
// ticket Block Handlers
// -------------------------

// ticketFormContentHandler (default partial)
// Extends: content
// Method: GET
// Doc: Landing page for ticket wizard
func ticketFormContentHandler(rsp treetop.Response, req *http.Request) interface{} {
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

var (
	wsRegex = regexp.MustCompile(`\s+`)
)

// remove redundant whitespace from a string that is to be used as a visual summary
func sanitizeSummary(s string) string {
	return strings.TrimSpace(wsRegex.ReplaceAllString(s, " "))
}
