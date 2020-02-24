package writer

import (
	"net/http"

	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/shared"
)

var (
	content = `
{{ block "content" . }}
<div id="content" style="text-align: center;">
	<h1>Treetop Writer Greeter</h1>
	<div>
		<form action="/writer/greet" treetop>
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
		<p><a href="/writer" treetop>Clear</a></p>
	</div>
{{ end }}
	`
)

// SetupGreeter registers writer greeter endpoints
func SetupGreeter(mux *http.ServeMux) {
	mux.HandleFunc("/writer/greet", greetingHandler)
	mux.HandleFunc("/writer", landingHandler)
}

func landingHandler(w http.ResponseWriter, req *http.Request) {
	var err error
	defer func() {
		if err != nil {
			panic(err)
		}
	}()
	if pw, ok := treetop.NewPartialWriter(w, req); ok {
		err = treetop.StringTemplateExec(pw, []string{landing}, nil)
		return
	}
	err = treetop.StringTemplateExec(w, []string{
		shared.BaseTemplate,
		content,
		landing,
	}, nil)
}

func greetingHandler(w http.ResponseWriter, req *http.Request) {
	var err error
	defer func() {
		if err != nil {
			panic(err)
		}
	}()
	who := req.URL.Query().Get("name")
	if who == "" {
		who = "World"
	}
	if pw, ok := treetop.NewPartialWriter(w, req); ok {
		err = treetop.StringTemplateExec(pw, []string{greeting}, who)
		return
	}
	// return full page instead
	err = treetop.StringTemplateExec(w, []string{
		shared.BaseTemplate,
		content,
		greeting,
	}, struct{ Message string }{who})
}
