package handlers

import (
	"net/http"
	"net/url"

	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/ticket/inputs"
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
	query := req.URL.Query()
	data := struct {
		AttachmentList interface{}
		FormMessage    interface{}
		Notes          interface{}
		Assignee       interface{}
		Description    string
		IssueType      string
	}{
		AttachmentList: rsp.HandleSubView("attachment-list", req),
		FormMessage:    rsp.HandleSubView("form-message", req),
		Description:    query.Get("description"),
		IssueType:      query.Get("issue-type"),
		Notes:          rsp.HandleSubView("notes", req),
		Assignee:       rsp.HandleSubView("assignee", req),
	}
	return data
}

// SubmitSoftwareTicket (partial)
// Extends: formMessage
// Method: POST
// Doc: process creation of a new help desk ticket
func SubmitSoftwareTicketHandler(rsp treetop.Response, req *http.Request) interface{} {
	var redirected bool
	defer func() {
		if !redirected && treetop.IsTemplateRequest(req) && len(req.PostForm) > 0 {
			// quality-of-life improvement, replace browser URL to include latest
			// form state so that a refresh will preserve inputs
			newURL, _ := url.Parse("/ticket/software/new")
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
	ticket := inputs.SoftwareTicketFromQuery(req.PostForm)

	// validation rules for creating a new Help Desk ticket
	// NOTE, Do not take client-side validation for granted
	if ticket.Summary == "" {
		data.Level = formMessageWarning
		data.Title = "Missing input"
		data.Message = "Ticket title is required"
		return data
	}

	if true {
		data.Level = formMessageWarning
		data.Title = "Not Implemented"
		data.Message = "Software form has not yet been implemented!"
		return data
	}

	// ticket is valid redirect to preview endpoint
	previewURL := url.URL{
		Path:     "/ticket/software/preview",
		RawQuery: ticket.RawQuery(),
	}
	treetop.Redirect(rsp, req, previewURL.String(), http.StatusSeeOther)
	redirected = true
	return nil
}

// SoftwareAssigneeHandler (fragment)
// Extends: reported-by
// Method: GET
// Doc: Options for notifying help desk of who reported the issue
func SoftwareAssigneeHandler(rsp treetop.Response, req *http.Request) interface{} {
	query := req.URL.Query()
	data := struct {
		Assignees []string
		FindUser  interface{}
	}{
		Assignees: query["assignees"],
	}
	if len(data.Assignees) >= 10 {
		return data
	}
	if addAssignee := query.Get("add-assignee"); addAssignee != "" {
		for _, user := range data.Assignees {
			if addAssignee == user {
				// already added, nothing to add
				goto CheckRemove
			}
		}
		data.Assignees = append(data.Assignees, addAssignee)
	}
CheckRemove:
	if removeAssignee := query.Get("remove-assignee"); removeAssignee != "" {
		for i, user := range data.Assignees {
			if removeAssignee == user {
				data.Assignees = append(data.Assignees[0:i], data.Assignees[i+1:]...)
				// no need to keep looking
				break
			}
		}
	}
	return data
}

// SoftwareFineAssigneeHandler (fragment)
// Extends: reported-by
// Method: GET
// Doc: Options for notifying help desk of who reported the issue
func SoftwareFindAssigneeHandler(rsp treetop.Response, req *http.Request) interface{} {
	query := req.URL.Query()

	type SearchResult struct {
		Name     string
		Selected bool
	}

	data := struct {
		Results     []SearchResult
		QueryString string
	}{
		QueryString: query.Get("search-query"),
	}

	selected := make(map[string]struct{})
	for _, user := range query["assignees"] {
		selected[user] = struct{}{}
	}

	for _, result := range inputs.SearchForUser(data.QueryString) {
		_, ok := selected[result]
		data.Results = append(data.Results, SearchResult{
			Name:     result,
			Selected: ok,
		})
	}

	return data
}
