package ticket

import (
	"net/http"

	"github.com/rur/treetop"
)

func assigneeHandler(rsp treetop.Response, req *http.Request) interface{} {
	query := req.URL.Query()
	data := struct {
		Assignee     string
		AssignedUser string
	}{
		Assignee: query.Get("assignee"),
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
