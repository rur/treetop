package writer

import (
	"net/http"
	"text/template"

	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/shared"
)

var (
	base    = template.Must(template.New("base").Parse(shared.BaseTemplate))
	content = template.Must(template.New("content").Parse(`
<div id="content" style="text-align: center;">
	<h1>Treetop Writer Greeter</h1>
	<div>
		<form action="/writer/greet" treetop>
			<span>Greet, </span><input placeholder="Name" type="text" name="name">
		</form>
	</div>
	{{ template "message" .Message}}
</div>
	`))
	landing = template.Must(template.New("message").Parse(`
	<p id="message"><i>Give me someone to say hello to!</i></p>
	`))
	greeting = template.Must(template.New("message").Parse(`
	<div id="message">
		<h2>Hello, {{ . }}!</h2>
		<p><a href="/writer" treetop>Clear</a></p>
	</div>
	`))
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
		landing.ExecuteTemplate(pw, "landing", nil)
		return
	}
	t := template.New("base")
	t.AddParseTree("base", base.Tree)
	t.AddParseTree("content", content.Tree)
	t.AddParseTree("message", landing.Tree)
	err = t.ExecuteTemplate(w, "base", nil)
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
		greeting.ExecuteTemplate(pw, "message", who)
		return
	}

	// return full page instead
	t := template.New("base")
	t.AddParseTree("base", base.Tree)
	t.AddParseTree("content", content.Tree)
	t.AddParseTree("message", greeting.Tree)
	err = t.ExecuteTemplate(w, "base", struct{ Message string }{who})
}
