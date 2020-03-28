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
		FindAssignee: rsp.HandleSubView("find-assignee", req),
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
	// for example purposes, vary number of results based solely
	// on the number of alpha characters in the input query
	data := struct {
		QueryResultTarget string
		Results           []string
	}{
		QueryResultTarget: query.Get("query-result-target"),
		Results:           []string{"User A", "User B", "User C"},
	}
	return data
}
