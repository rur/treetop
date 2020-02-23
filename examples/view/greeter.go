package view

import (
	"net/http"

	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/shared"
)

var (
	content = `
{{ block "content" . }}
<div id="content" style="text-align: center;">
	<h1>Treetop View Greeter</h1>
	<div>
		<form action="/view/greet" treetop>
			<span>Greet, </span><input placeholder="Name" type="text" name="name">
		</form>
	</div>
	{{ template "message" .Message}}
</div>
{{ end }}
	`
	landing = `
{{ block "message" . }}
	<p id="message"><i>Give me someone to say hello to!</i></p>
{{ end }}
	`
	greeting = `
{{ block "message" . }}
	<div id="message">
		<h2>Hello, {{ . }}!</h2>
		<p><a href="/view" treetop>Clear</a></p>
	</div>
{{ end }}
	`
)

// SetupGreeter add routes for /view example endpoint
func SetupGreeter(mux *http.ServeMux) {
	page := treetop.NewView(
		treetop.StringTemplateExec,
		shared.BaseTemplate,
		treetop.Delegate("content"), // adopt "content" handler as own template data
	)
	content := page.NewSubView("content", content, contentHandler)

	greetForm := content.NewSubView("message", landing, treetop.Noop)
	greetMessage := content.NewSubView("message", greeting, greetingHandler)

	mux.Handle("/view", treetop.ViewHandler(greetForm))
	mux.Handle("/view/greet", treetop.ViewHandler(greetMessage))
}

// contentHandler loads data for the content template
func contentHandler(rsp treetop.Response, req *http.Request) interface{} {
	return struct {
		// In the content template the .Message field is passed explicitly to the
		// "message" template block.
		Message interface{}
	}{
		// HandleSubView triggers the sub handler call and returns the sub template data.
		// `who` in the case of the greetingHandler, or nil in the case of Noop
		Message: rsp.HandleSubView("message", req),
	}
}

// greetingHandler obtains the name for the greeting from the request query
func greetingHandler(_ treetop.Response, req *http.Request) interface{} {
	who := req.URL.Query().Get("name")
	if who == "" {
		return "World"
	}
	// NOTE: relying on Go package html/template for input escaping
	return who
}
