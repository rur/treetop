package inline

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/rur/treetop"
)

// profileContentHandler
// extends: content
func profileContentHandler(rsp treetop.Response, req *http.Request) interface{} {
	switch req.Method {
	case "GET", "HEAD":
		return struct {
			ViewTree    string
			FirstName   interface{}
			Surname     interface{}
			Email       interface{}
			Country     interface{}
			Description interface{}
		}{
			ViewTree:    viewDebug,
			FirstName:   rsp.HandleSubView("first-name", req),
			Surname:     rsp.HandleSubView("surname", req),
			Email:       rsp.HandleSubView("email", req),
			Country:     rsp.HandleSubView("country", req),
			Description: rsp.HandleSubView("description", req),
		}

	default:
		pageErrorMessage(rsp, req, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return nil
	}
}

// getFormFieldHandler will create a request handler for an editable
// field of the FormData object
func getFormFieldHandler(field string) formDataHandlerFunc {
	return func(form *FormData, rsp treetop.Response, req *http.Request) interface{} {
		// data structure to be passed to the template
		data := struct {
			Field        string
			Value        string
			ErrorMessage string
			Editing      bool
			ElementID    string
			Title        string
			Options      []string
			Type         string
		}{
			Field: field,
		}
		var processInput func(url.Values, string) (string, string)
		switch data.Field {
		case "firstName":
			data.Title = "First Name"
			data.Value = form.FirstName
			processInput = processInputName
		case "surname":
			data.Title = "Last Name"
			data.Value = form.LastName
			processInput = processInputName
		case "email":
			data.Title = "Email"
			data.Value = form.Email
			data.Type = "email"
			processInput = processInputEmail
		case "country":
			data.Title = "Country"
			data.Value = form.Country
			data.Options = CountryOptions
			processInput = processInputContry
		case "description":
			data.Title = "Description"
			data.Value = form.Description
			processInput = processInputDescription
		default:
			data.ErrorMessage = fmt.Sprintf("Unknown field '%s'", field)
			return data
		}

		switch req.Method {
		case "GET", "HEAD":
			data.Editing = req.URL.Query().Get("edit") == "true"
			return data

		case "POST":
			if err := req.ParseForm(); err != nil {
				pageErrorMessage(rsp, req, "Failed to parse form", http.StatusBadRequest)
				return nil
			}

			data.Value, data.ErrorMessage = processInput(req.PostForm, field)
			form.SetField(field, data.Value)

			if data.ErrorMessage != "" {
				rsp.Status(http.StatusBadRequest)
				data.Editing = true
			} else {
				// commit the updated form data
				setCookieFormData(rsp, form)
				data.Editing = false
			}
			return data
		default:
			pageErrorMessage(rsp, req, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return nil
		}
	}
}
