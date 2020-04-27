package handlers

import (
	"net/http"
	"net/url"

	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/ticket/inputs"
)

// newSystemsTicket (partial)
// Extends: form
// Method: GET
// Doc: Form designed for creating systems tickets
func NewSystemsTicketHandler(rsp treetop.Response, req *http.Request) interface{} {
	if treetop.IsTemplateRequest(req) {
		// replace existing browser history entry with current URL
		rsp.ReplacePageURL(req.URL.String())
	}
	query := req.URL.Query()
	data := struct {
		AttachmentList interface{}
		ComponentTags  interface{}
		FormMessage    interface{}
		Notes          interface{}
		Description    string
	}{
		ComponentTags:  rsp.HandleSubView("component-tags", req),
		AttachmentList: rsp.HandleSubView("attachment-list", req),
		FormMessage:    rsp.HandleSubView("form-message", req),
		Description:    query.Get("description"),
		Notes:          rsp.HandleSubView("notes", req),
	}
	return data
}

// SubmitSystemsTicket (partial)
// Extends: formMessage
// Method: POST
// Doc: process creation of a new help desk ticket
func SubmitSystemsTicketHandler(rsp treetop.Response, req *http.Request) interface{} {
	var redirected bool
	defer func() {
		if !redirected && treetop.IsTemplateRequest(req) && len(req.PostForm) > 0 {
			// quality-of-life improvement, replace browser URL to include latest
			// form state so that a refresh will preserve inputs
			newURL, _ := url.Parse("/ticket/systems/new")
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
		Level   int
		Message string
		Title   string
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
	ticket := inputs.SystemsTicketFromQuery(req.PostForm)

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
		data.Message = "This has not been implemented just yet!"
		return data
	}

	// ticket is valid redirect to preview endpoint
	previewURL := url.URL{
		Path:     "/ticket/systems/preview",
		RawQuery: ticket.RawQuery(),
	}
	treetop.Redirect(rsp, req, previewURL.String(), http.StatusSeeOther)
	redirected = true
	return nil
}

// PreviewSystemsTicketHandler (partial)
// Extends: content
// Method: GET
// Doc: Show preview of systems ticket, no database so take details form query params
func PreviewSystemsTicketHandler(rsp treetop.Response, req *http.Request) interface{} {
	data := struct {
		EditLink string
		Ticket   *inputs.SystemsTicket
	}{
		// generally this would be loaded from a database but for the demo
		// we are only previewing from URL parameters
		Ticket: inputs.SystemsTicketFromQuery(req.URL.Query()),
	}
	formURL := url.URL{}
	formURL.Path = "/ticket/systems/new"
	formURL.RawQuery = req.URL.Query().Encode()
	data.EditLink = formURL.String()
	return data
}

// SystemsComponentTagsInputGroup (fragment)
// Extends: componentTags
// Method: GET
// Doc: Load form input group for the component tags selector
func SystemsComponentTagsInputGroupHandler(rsp treetop.Response, req *http.Request) interface{} {
	query := req.URL.Query()
	data := struct {
		Tags      []string
		TagSearch interface{}
		AutoFocus bool
	}{
		TagSearch: rsp.HandleSubView("tag-search", req),
		Tags:      query["tags"],
	}

	if add := query.Get("add-tag"); add != "" {
		data.AutoFocus = true
		for _, t := range data.Tags {
			if t == add {
				goto Next
			}
		}
		data.Tags = append(data.Tags, add)
	}
Next:

	return data
}

// SystemsComponentTagSearch (fragment)
// Extends: tagSearch
// Method: GET
// Doc: fuzzy match query to available systems component tags
func SystemsComponentTagSearchHandler(rsp treetop.Response, req *http.Request) interface{} {
	query := req.URL.Query()
	existingTags := make(map[string]struct{})
	for _, t := range query["tags"] {
		existingTags[t] = struct{}{}
	}
	data := struct {
		Query   string
		Results []string
	}{
		Query: query.Get("tag-query"),
	}
	placeHolder := []string{
		"Example Tag A",
		"Example Tag B",
		"Example Tag C",
		"Example Tag D",
	}
	for _, p := range placeHolder {
		if _, ok := existingTags[p]; !ok {
			data.Results = append(data.Results, p)
		}
	}
	return data
}
