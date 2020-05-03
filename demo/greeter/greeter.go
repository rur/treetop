package greeter

import (
	"net/http"

	"github.com/rur/treetop"
	"github.com/rur/treetop/demo/assets"
)

// Setup register routes for /greeter demo endpoint
func Setup(mux *http.ServeMux, devMode bool) {
	// base view
	page := treetop.NewView("local://base.html", treetop.Delegate("content"))
	_ = page.NewDefaultSubView("nav", "local://nav.html", treetop.Noop)

	// Greeter Content View
	content := page.NewSubView("content", "demo/greeter/templates/content.html", contentViewHandler)
	// content -> message
	greetForm := content.NewSubView("message", "local://landing.html", treetop.Noop)
	greetMessage := content.NewSubView("message", "demo/greeter/templates/greeting.html", greetingViewHandler)
	// content -> notes
	hideNotes := content.NewSubView("notes", "local://hide-notes.html", treetop.Noop)
	notes := content.NewSubView("notes", "demo/greeter/templates/notes.html", treetop.Noop)

	// Configure template executor with a hybrid of template files and inlined string templates
	var exec treetop.ViewExecutor = &treetop.FileExecutor{
		KeyedString: map[string]string{
			"local://base.html":       assets.BaseHTML,
			"local://nav.html":        assets.NavHTML(assets.GreeterNav),
			"local://landing.html":    `<p id="message"><i>Give me someone to say hello to!</i></p>`,
			"local://hide-notes.html": `<div id="notes" class="hide"></div>`,
		},
	}
	if devMode {
		// Use developer executor to permit template file editing
		exec = &treetop.DeveloperExecutor{
			ViewExecutor: exec,
		}
	}

	mux.Handle("/greeter", exec.NewViewHandler(
		greetForm,
		hideNotes))
	mux.Handle("/greeter/greet", exec.NewViewHandler(
		greetMessage,
		notes))

	if errs := exec.FlushErrors(); len(errs) != 0 {
		panic(errs.Error())
	}
}

// contentViewHandler loads data for the content template
func contentViewHandler(rsp treetop.Response, req *http.Request) interface{} {
	return struct {
		Value   string
		Message interface{}
		Notes   interface{}
	}{
		// initialize form text input (relying on auto-sanitization)
		Value:   req.URL.Query().Get("name"),
		Message: rsp.HandleSubView("message", req),
		Notes:   rsp.HandleSubView("notes", req),
	}
}

// greetingViewHandler obtains the name for the greeting from the request query
func greetingViewHandler(_ treetop.Response, req *http.Request) interface{} {
	query := req.URL.Query()
	who := query.Get("name")
	if who == "" {
		who = "World"
	}
	return who
}
