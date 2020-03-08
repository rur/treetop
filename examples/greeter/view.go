package greeter

import (
	"net/http"

	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/assets"
)

// GreetViewRoutes register routes for /view example endpoint
func GreetViewRoutes(mux *http.ServeMux) {
	page := treetop.NewView(
		assets.BaseHTML,
		treetop.Delegate("content"), // adopt "content" handler as own template data
	)
	page.NewDefaultSubView("nav", assets.NavHTML("View"), treetop.Noop)
	content := page.NewSubView("content", ContentHTML("View", "/view"), contentViewHandler)

	greetForm := content.NewSubView("message", LandingHTML, treetop.Noop)
	greetMessage := content.NewSubView("message", GreetingHTML("/view"), greetingViewHandler)

	exec := treetop.StringExecutor{}
	mux.Handle("/view", exec.NewViewHandler(greetForm))
	mux.Handle("/view/greet", exec.NewViewHandler(greetMessage))
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
