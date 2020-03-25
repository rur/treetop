package ticket

import (
	"net/http"

	"github.com/rur/treetop"
)

// formHandler
// doc: basis for all forms, using a generic one during development
// extends: content.html{form}
func formHandler(rsp treetop.Response, req *http.Request) interface{} {
	if treetop.IsTemplateRequest(req) {
		// replace existing browser history entry with current URL
		rsp.ReplacePageURL(req.URL.String())
	}
	data := struct {
		// TODO: add components here
		AssigneeData interface{}
	}{
		AssigneeData: rsp.HandleSubView("assignee", req),
	}
	return data
}
