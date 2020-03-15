package greeter

import (
	"fmt"
	"net/http"

	"github.com/rur/treetop"
)

// GreetTemplateData is created from the request query params
// and passed to the template
type GreetTemplateData struct {
	Who        string
	Notes      string
	IsFullPage bool
}

// GetGreetingQuery interprets the greeting from the request URL query
func GetGreetingQuery(req *http.Request) GreetTemplateData {
	query := req.URL.Query()
	data := GreetTemplateData{
		Who: query.Get("name"),
	}
	if !treetop.IsTemplateRequest(req) {
		data.Notes = "Full page request!"
		data.IsFullPage = true
	} else if submitter := query.Get("submitter"); submitter != "" {
		data.Notes = fmt.Sprintf("XHR form submit with the '%s' button submitter!", submitter)
	} else {
		data.Notes = "XHR form submit, notice that the text input cursor is preserved."
	}
	if data.Who == "" {
		data.Who = "World"
	}
	return data
}
