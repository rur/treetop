package handlers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/ticket/inputs"
)

// newHelpdeskTicket (partial)
// Extends: form
// Method: GET
// Doc: Form designed for creating helpdesk tickets
func NewHelpdeskTicketHandler(rsp treetop.Response, req *http.Request) interface{} {
	if treetop.IsTemplateRequest(req) {
		// replace existing browser history entry with current URL
		rsp.ReplacePageURL(req.URL.String())
	}
	query := req.URL.Query()
	data := struct {
		ReportedBy     interface{}
		AttachmentList interface{}
		FormMessage    interface{}
		Notes          interface{}
		Description    string
		Urgency        string
	}{
		ReportedBy:     rsp.HandleSubView("reported-by", req),
		AttachmentList: rsp.HandleSubView("attachment-list", req),
		FormMessage:    rsp.HandleSubView("form-message", req),
		Description:    query.Get("description"),
		Urgency:        query.Get("urgency"),
		Notes:          rsp.HandleSubView("notes", req),
	}
	return data
}

// helpdeskReportedBy (fragment)
// Extends: reported-by
// Method: GET
// Doc: Options for notifying help desk of who reported the issue
func HelpdeskReportedByHandler(rsp treetop.Response, req *http.Request) interface{} {
	query := req.URL.Query()
	data := struct {
		ReportedBy         string
		ReportedByUser     string
		ReportedByCustomer string
		FindUser           interface{}
		CustomerList       []string
		CustomerContact    string
	}{
		ReportedBy: query.Get("reported-by"),
	}
	// Use allow-list for input validation when possible
	switch data.ReportedBy {
	case "team-member":
		// Now parse extra input for this setting
		data.ReportedByUser = query.Get("reported-by-user")
		data.FindUser = rsp.HandleSubView("find-user", req)
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

// Method: GET
// Doc: search the database for a list of users matching a query string
func FindReportedByUserHandler(rsp treetop.Response, req *http.Request) interface{} {
	query := req.URL.Query()
	return struct {
		Results     []string
		QueryString string
	}{
		QueryString: query.Get("search-query"),
		Results:     inputs.SearchForUser(query.Get("search-query")),
	}
}

const (
	formMessageInfo = iota
	formMessageWarning
	formMessageError
)

// submitHelpDeskTicket (partial)
// Extends: formMessage
// Method: POST
// Doc: process creation of a new help desk ticket
func SubmitHelpDeskTicketHandler(rsp treetop.Response, req *http.Request) interface{} {
	var redirected bool
	defer func() {
		if !redirected && treetop.IsTemplateRequest(req) && len(req.PostForm) > 0 {
			// quality-of-life improvement, replace browser URL to include latest
			// form state so that a refresh will preserve inputs
			newURL, _ := url.Parse("/ticket/helpdesk/new")
			q := req.PostForm
			q.Del("file-upload")
			newURL.RawQuery = q.Encode()
			newURL.Fragment = "form-message"
			// replace existing browser history entry with current URL
			rsp.ReplacePageURL(newURL.String())
		}
	}()

	// If all inputs are valid this handler will redirect the web browser
	// either to the newly created ticket or to a blank form.
	//
	// If creation cannot proceed for any reason, this endpoint will render
	// a form message HTML framgent with an alert level: info, warning or error
	data := struct {
		Level           int
		Message         string
		Title           string
		ConfirmCritical bool
	}{
		Level: formMessageInfo,
	}

	if err := req.ParseForm(); err != nil {
		rsp.Status(http.StatusBadRequest)
		data.Level = formMessageError
		data.Title = "Request Error"
		data.Message = "Failed to read form data, try again or contact support."
		return data
	}
	ticket := inputs.HelpdeskTicketFromQuery(req.PostForm)

	// validation rules for creating a new Help Desk ticket
	// NOTE, Do not take client-side validation for granted
	if ticket.Summary == "" {
		data.Level = formMessageWarning
		data.Title = "Missing input"
		data.Message = "Ticket title is required"
		return data
	}
	switch ticket.ReportedBy {
	case "team-member":
		if ticket.ReportedByUser == "" {
			rsp.Status(http.StatusBadRequest)
			data.Level = formMessageWarning
			data.Title = "Missing input"
			data.Message = "Please sepecify which user reported the issue"
			return data
		}
	case "customer":
		if ticket.ReportedByCustomer == "" {
			rsp.Status(http.StatusBadRequest)
			data.Level = formMessageWarning
			data.Title = "Missing input"
			data.Message = "Please sepecify which customer reported the issue"
			return data
		}
	case "":
		rsp.Status(http.StatusBadRequest)
		data.Level = formMessageWarning
		data.Title = "Missing input"
		data.Message = "Please sepecify for whom this issue is being reported"
		return data
	}
	if ticket.Urgency == "" {
		rsp.Status(http.StatusBadRequest)
		data.Level = formMessageWarning
		data.Title = "Invalid input"
		data.Message = fmt.Sprintf("Invalid ticket urgency value '%s'",
			req.PostForm.Get("urgency"))
		return data
	}

	if ticket.Urgency == "critical" && req.PostForm.Get("confirm-critical") != "yes" {
		data.ConfirmCritical = true
		return data
	}

	// ticket is valid redirect to preview endpoint
	previewURL := url.URL{
		Path:     "/ticket/helpdesk/preview",
		RawQuery: ticket.RawQuery(),
	}
	treetop.Redirect(rsp, req, previewURL.String(), http.StatusSeeOther)
	redirected = true
	return nil
}

// previewHelpdeskTicket (partial)
// Extends: content
// Method: GET
// Doc: Show preview of help desk ticket, no database so take details form query params
func PreviewHelpdeskTicketHandler(rsp treetop.Response, req *http.Request) interface{} {
	data := struct {
		EditLink string
		Ticket   *inputs.HelpDeskTicket
	}{
		// generally this would be loaded from a database but for the demo
		// we are only previewing from URL parameters
		Ticket: inputs.HelpdeskTicketFromQuery(req.URL.Query()),
	}
	formURL := url.URL{}
	formURL.Path = "/ticket/helpdesk/new"
	formURL.RawQuery = req.URL.Query().Encode()
	data.EditLink = formURL.String()
	return data
}
