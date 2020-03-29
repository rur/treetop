package ticket

import (
	"net/http"

	"github.com/rur/treetop"
)

// newHelpdeskTicket (partial)
// Extends: form
// Method: *
// Doc: Form designed for creating helpdesk tickets
func newHelpdeskTicketHandler(rsp treetop.Response, req *http.Request) interface{} {
	if treetop.IsTemplateRequest(req) {
		// replace existing browser history entry with current URL
		rsp.ReplacePageURL(req.URL.String())
	}
	data := struct {
		Assignee interface{}
	}{
		Assignee: rsp.HandleSubView("assignee", req),
	}
	return data
}

// helpdeskAssignee (partial)
// Extends: assignee
// Method: *
// Doc: Options for assigning someone to the help desk item, different options for admins
func helpdeskAssigneeHandler(rsp treetop.Response, req *http.Request) interface{} {
	query := req.URL.Query()
	data := struct {
		Assignee     string
		AssignedUser string
		FindAssignee interface{}
		IsAdmin      bool
	}{
		Assignee:     query.Get("assignee"),
		FindAssignee: rsp.HandleSubView("findassignee", req),
		IsAdmin:      true,
	}
	// allow-list approach to input validation with a default fallback
	switch data.Assignee {
	case "myself", "unassigned":
		// ok!
	case "user-name":
		// ok!
		// Now parse extra input for this setting
		data.AssignedUser = query.Get("assigned-user")
	default:
		data.Assignee = "unassigned"
	}
	return data
}

// findHelpdeskAssignee (partial)
// Extends: findAssignee
// Method: *
// Doc: user supplied query string to select a user to assign to helpdesk ticket
func findHelpdeskAssigneeHandler(rsp treetop.Response, req *http.Request) interface{} {
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
