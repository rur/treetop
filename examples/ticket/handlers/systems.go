package handlers

import (
	"net/http"

	"github.com/rur/treetop"
)

// newSystemsTicket (partial)
// Extends: form
// Method: GET
// Doc: Form designed for creating systems tickets
func NewSystemsTicketHandler(rsp treetop.Response, req *http.Request) interface{} {
	data := struct {
		HandlerInfo string
	}{
		HandlerInfo: "newSystemsTicket",
	}
	return data
}
