package inline

import (
	"fmt"
	"net/http"

	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/assets"
)

func Routes(mux *http.ServeMux) {
	srv := newCookieServer()

	page := treetop.NewView("local://base.html", treetop.Delegate("content"))
	_ = page.NewDefaultSubView("nav", "local://nav.html", treetop.Noop)
	content := page.NewSubView("content",
		"examples/inline/templates/content.html.tmpl",
		srv.bind(ticketContentHandler))

	firstName := content.NewDefaultSubView("first-name",
		"examples/inline/templates/input.html.tmpl",
		srv.bind(getFormFieldHandler("firstName")))

	surname := content.NewDefaultSubView("surname",
		"examples/inline/templates/input.html.tmpl",
		srv.bind(getFormFieldHandler("surname")))

	email := content.NewDefaultSubView("email",
		"examples/inline/templates/input.html.tmpl",
		srv.bind(getFormFieldHandler("email")))

	country := content.NewDefaultSubView("country",
		"examples/inline/templates/input.html.tmpl",
		srv.bind(getFormFieldHandler("country")))

	description := content.NewDefaultSubView("description",
		"examples/inline/templates/input.html.tmpl",
		srv.bind(getFormFieldHandler("description")))

	exec := &treetop.DeveloperExecutor{
		ViewExecutor: &treetop.FileExecutor{
			KeyedString: map[string]string{
				"local://base.html": assets.BaseHTML,
				"local://nav.html":  assets.NavHTML(assets.InlineNav),
			},
		},
	}

	mux.Handle("/inline", exec.NewViewHandler(content))
	mux.Handle("/inline/firstName", exec.NewViewHandler(firstName).FragmentOnly())
	mux.Handle("/inline/surname", exec.NewViewHandler(surname).FragmentOnly())
	mux.Handle("/inline/email", exec.NewViewHandler(email).FragmentOnly())
	mux.Handle("/inline/country", exec.NewViewHandler(country).FragmentOnly())
	mux.Handle("/inline/description", exec.NewViewHandler(description).FragmentOnly())

	if errs := exec.FlushErrors(); len(errs) != 0 {
		panic(errs.Error())
	}
}

// ticketContentHandler
// extends: base.html{content}
func ticketContentHandler(form *FormData, rsp treetop.Response, req *http.Request) interface{} {
	switch req.Method {
	case "GET", "HEAD":
		return struct {
			FormData    *FormData
			FirstName   interface{}
			Surname     interface{}
			Email       interface{}
			Country     interface{}
			Description interface{}
		}{
			FormData:    form,
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
		data := struct {
			Field        string
			Value        string
			ErrorMessage string
			Editing      bool
			ElementID    string
			Title        string
		}{
			Field: field,
		}
		var validate func(string) string
		switch data.Field {
		case "firstName":
			data.Title = "First Name"
			data.Value = form.FirstName
			validate = assertValidName
		case "surname":
			data.Title = "Last Name"
			data.Value = form.LastName
			validate = assertValidName
		case "email":
			data.Title = "Email"
			data.Value = form.Email
			validate = assertValidEmail
		case "country":
			data.Title = "Country"
			data.Value = form.Country
			validate = assertValidContry
		case "description":
			data.Title = "Description"
			data.Value = form.Description
			validate = assertValidDescription
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

			value := req.PostFormValue(field)
			data.Value = value
			form.SetField(field, value)
			data.ErrorMessage = validate(value)

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
