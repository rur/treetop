package greeter

import (
	"net/http"

	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/assets"
)

// Routes register routes for /greeter example endpoint
func Routes(mux *http.ServeMux) {
	page := treetop.NewView("local://base.html", treetop.Delegate("content"))
	_ = page.NewDefaultSubView("nav", "local://nav.html", treetop.Noop)
	content := page.NewSubView("content", "examples/greeter/templates/content.html", contentViewHandler)
	greetForm := content.NewSubView("message", "local://landing.html", treetop.Noop)
	greetMessage := content.NewSubView("message", "examples/greeter/templates/greeting.html", greetingViewHandler)

	exec := &treetop.DeveloperExecutor{
		ViewExecutor: &treetop.FileExecutor{
			KeyedString: map[string]string{
				"local://base.html":    assets.BaseHTML,
				"local://nav.html":     assets.NavHTML(assets.GreeterNav),
				"local://landing.html": `<p id="message"><i>Give me someone to say hello to!</i></p>`,
			},
		},
	}

	mux.Handle("/greeter", exec.NewViewHandler(greetForm))
	mux.Handle("/greeter/greet", exec.NewViewHandler(greetMessage))

	if errs := exec.FlushErrors(); len(errs) != 0 {
		panic(errs.Error())
	}
}

// contentViewHandler loads data for the content template
func contentViewHandler(rsp treetop.Response, req *http.Request) interface{} {
	return struct {
		Value   string
		Message interface{}
	}{
		// initialize form text input (relying on auto-sanitization)
		Value:   req.URL.Query().Get("name"),
		Message: rsp.HandleSubView("message", req),
	}
}

// greetingViewHandler obtains the name for the greeting from the request query
func greetingViewHandler(_ treetop.Response, req *http.Request) interface{} {
	return getGreetingQuery(req)
}
