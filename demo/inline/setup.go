package inline

import (
	"net/http"

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
