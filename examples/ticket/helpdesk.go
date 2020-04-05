package ticket

import (
	"log"
	"net/http"
	"net/url"

	"github.com/rur/treetop"
)

const (
	formMessageInfo = iota
	formMessageWarning
	formMessageError
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
		FormMessage    interface{}
		Description    string
	}{
		ReportedBy:     rsp.HandleSubView("reported-by", req),
		UploadFileList: rsp.HandleSubView("upload-file-list", req),
		FormMessage:    rsp.HandleSubView("form-message", req),
		Description:    req.URL.Query().Get("description"),
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
		data.ReportedByUser = query.Get("reported-by-user")
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

// Doc: Default helpdesk attachment file list template handler,
//      parse file info from query string
func helpdeskAttachmentFileListHandler(rsp treetop.Response, req *http.Request) interface{} {
	// load file info from query
	query := req.URL.Query()
	data := struct {
		Files []*FileInfo
	}{}

	for _, enc := range query["attachment"] {
		info := &FileInfo{}
		if err := info.UnmarshalBase64([]byte(enc)); err != nil {
			// skip it
			log.Println(err)
		} else {
			data.Files = append(data.Files, info)
		}
	}
	return data
}

// submitHelpDeskTicket (partial)
// Extends: formMessage
// Method: POST
// Doc: process creation of a new help desk ticket
func submitHelpDeskTicketHandler(rsp treetop.Response, req *http.Request) interface{} {
	// If all inputs are valid this handler will redirect the web browser
	// either to the newly created ticket or to a blank form.
	//
	// If creation cannot proceed for any reason, this endpoint will render
	// a form message HTML framgent with an alert level: info, warning or error
	data := struct {
		Level   int
		Message string
	}{
		Level: formMessageInfo,
	}

	if err := req.ParseForm(); err != nil {
		rsp.Status(http.StatusBadRequest)
		data.Level = formMessageError
		data.Message = "Failed to read form data, try again or contact support."
		return data
	}

	if treetop.IsTemplateRequest(req) {
		newURL, _ := url.Parse("/ticket/helpdesk/new")
		q := req.PostForm
		q.Del("file-upload")
		newURL.RawQuery = q.Encode()
		newURL.Fragment = "form-message"
		// replace existing browser history entry with current URL
		rsp.DesignatePageURL(newURL.String())
	}

	return data
}
