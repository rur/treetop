package handlers

import (
	"net/http"

	"github.com/rur/treetop"
)

// newSoftwareTicket (partial)
// Extends: form
// Method: GET
// Doc: Form designed for creating software tickets
func NewSoftwareTicketHandler(rsp treetop.Response, req *http.Request) interface{} {
	if treetop.IsTemplateRequest(req) {
		// replace existing browser history entry with current URL
		rsp.ReplacePageURL(req.URL.String())
	}
	// TODO: implement this form
	return nil
}
