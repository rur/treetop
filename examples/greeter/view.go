package greeter

import (
	"net/http"

	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/assets"
)

// GreetViewRoutes register routes for /view example endpoint
func GreetViewRoutes(mux *http.ServeMux) {
	page := treetop.NewView("base.html", treetop.Delegate("content"))
	_ = page.NewDefaultSubView("nav", "nav.html", treetop.Noop)
	content := page.NewSubView("content", "content.html", contentViewHandler)
	greetForm := content.NewSubView("message", "landing.html", treetop.Noop)
	greetMessage := content.NewSubView("message", "greeting.html", greetingViewHandler)

	exec := treetop.NewKeyedStringExecutor(map[string]string{
		"base.html":     assets.BaseHTML,
		"nav.html":      assets.NavHTML("View"),
		"content.html":  ContentHTML("View", "/view"),
		"landing.html":  LandingHTML,
		"greeting.html": GreetingHTML("/view"),
	})

	mux.Handle("/view", exec.NewViewHandler(greetForm))
	mux.Handle("/view/greet", exec.NewViewHandler(greetMessage))

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
