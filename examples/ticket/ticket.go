package ticket

import (
	"net/http"
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
func ticketHandler(rsp treetop.Response, req *http.Request) interface{} {
	query := req.URL.Query()
	data := struct {
		Summary string
		Dept    string
		Form    interface{}
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
		formDepartmentRedirectHandler(rsp, req)
		return nil
	}
	return data
}

// Method: GET
// Doc: search the database for a list of users matching a query string
func findTeamMemberHandler(rsp treetop.Response, req *http.Request) interface{} {
	query := req.URL.Query()

	data := struct {
		Results     []string
		QueryString string
	}{
		QueryString: query.Get("search-query"),
	}

	// For demo purposes, filter out any characters not in the latin alphabet.
	// All other characters must be in an allowlist, otherwise the result set will be empty
	filteredQuery := make([]byte, 0, len(data.QueryString))
FILTER:
	for _, codePoint := range data.QueryString {
		if (codePoint >= 64 && codePoint <= 90) || (codePoint >= 97 && codePoint <= 122) {
			filteredQuery = append(filteredQuery, byte(codePoint))
			continue
		}
		switch codePoint {
		case ' ', '-', '_', '.', '\t':
			// allowed non latin alphabet character, skip for filter
			continue
		default:
			filteredQuery = nil
			break FILTER
		}
	}
	if len(filteredQuery) == 0 {
		return data
	}

	// For example purposes, vary number of results based
	// on the number of characters in the input query.
	for i := len(filteredQuery) - 1; i < 26; i++ {
		data.Results = append(data.Results, "Example User "+string(i+65))
		if len(data.Results) == 5 {
			break
		}
	}

	return data
}
