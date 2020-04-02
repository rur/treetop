package ticket

import (
	"net/http"

	"github.com/rur/treetop"
)

// newHelpdeskTicket (partial)
// Extends: form
// Method: GET
// Doc: Form designed for creating helpdesk tickets
func newHelpdeskTicketHandler(rsp treetop.Response, req *http.Request) interface{} {
	if treetop.IsTemplateRequest(req) {
		// replace existing browser history entry with current URL
		rsp.ReplacePageURL(req.URL.String())
	}
	data := struct {
		ReportedBy     interface{}
		UploadFileList interface{}
	}{
		ReportedBy:     rsp.HandleSubView("reported-by", req),
		UploadFileList: rsp.HandleSubView("upload-file-list", req),
	}
	return data
}

// helpdeskReportedBy (fragment)
// Extends: reportedBy
// Method: GET
// Doc: Options for notifying help desk of who reported the issue
func helpdeskReportedByHandler(rsp treetop.Response, req *http.Request) interface{} {
	query := req.URL.Query()
	data := struct {
		ReportedBy         string
		ReportedByUser     string
		ReportedByCustomer string
		CustomerList       []string
		CustomerContact    string
	}{
		ReportedBy: query.Get("reported-by"),
	}
	// Use allow-list for input validation when possible
	switch data.ReportedBy {
	case "user-name":
		// Now parse extra input for this setting
		data.ReportedByUser = query.Get("resported-by-user")
	case "customer":
		// Would otherwise be loaded from a customer database
		data.CustomerList = []string{
			"Example Customer 0",
			"Example Customer 1",
			"Example Customer 2",
			"Example Customer 3",
			"Example Customer A",
			"Example Customer B",
			"Example Customer C",
			"Example Customer D",
			"Example Customer E",
		}
		if rBC := query.Get("reported-by-customer"); rBC != "" {
			for _, cst := range data.CustomerList {
				if cst == rBC {
					// accept the input when it matches a known customer
					data.ReportedByCustomer = rBC
					break
				}
			}
		}
		data.CustomerContact = query.Get("customer-contact")
	case "myself":
		// no additional information required
	default:
		// default for empty or unrecognized input
		data.ReportedBy = ""
	}
	return data
}

// uploadedHelpdeskFiles (fragment)
// Extends: uploadFileList
// Method: POST
// Doc: Load a list of uploaded files, save to storage and return metadata to the form
func uploadedHelpdeskFilesHandler(rsp treetop.Response, req *http.Request) interface{} {
	data := struct {
		HandlerInfo string
	}{
		HandlerInfo: "uploadedHelpdeskFiles",
	}
	return data
}
