package inline

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/rur/treetop"
	"github.com/rur/treetop/demo/assets"
)

var viewDebug string

// Setup will construct a view hierarchy for this form and bind
// handlers to the supplied HTTP request router.
func Setup(mux *http.ServeMux, devMode bool) {
	srv := newCookieServer()

	base := treetop.NewView("local://base.html", treetop.Delegate("content"))
	_ = base.NewDefaultSubView("nav", "local://nav.html", treetop.Noop)
	content := base.NewSubView("content",
		"demo/inline/templates/content.html.tmpl",
		profileContentHandler)

	firstName := content.NewDefaultSubView("first-name",
		"demo/inline/templates/input.html.tmpl",
		srv.bind(getFormFieldHandler("firstName")))

	surname := content.NewDefaultSubView("surname",
		"demo/inline/templates/input.html.tmpl",
		srv.bind(getFormFieldHandler("surname")))

	email := content.NewDefaultSubView("email",
		"demo/inline/templates/email.html.tmpl",
		srv.bind(getFormFieldHandler("email")))

	country := content.NewDefaultSubView("country",
		"demo/inline/templates/select.html.tmpl",
		srv.bind(getFormFieldHandler("country")))

	description := content.NewDefaultSubView("description",
		"demo/inline/templates/textarea.html.tmpl",
		srv.bind(getFormFieldHandler("description")))

	var exec treetop.ViewExecutor = &treetop.FileExecutor{
		KeyedString: map[string]string{
			"local://base.html": assets.BaseHTML,
			"local://nav.html":  assets.NavHTML(assets.InlineNav),
		},
	}
	if devMode {
		// Use developer executor to permit template file editing
		exec = &treetop.DeveloperExecutor{
			ViewExecutor: exec,
		}
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

	// get debug string print for this page
	page, _, _ := treetop.CompileViews(content)
	viewDebug = treetop.SprintViewTree(page)
}

// profileContentHandler
// extends: base.html{content}
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
// extends: content.html{?field?}
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
