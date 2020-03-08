package greeter

import (
	"fmt"
	"net/http"

	"github.com/rur/treetop"
)

var LandingHTML = `
	<p id="message"><i>Give me someone to say hello to!</i></p>
	`

func ContentHTML(title, base string) string {
	return fmt.Sprintf(`
	<div id="content">
		<hr>
		<h3 class="mb-3">Treetop %s Greeter</h3>
		<form action="%s/greet" treetop>
			<div class="input-group mb-3">
				<input id="name"
					name="name"
					type="text"
					autofocus tabindex="0"
					class="form-control"
					aria-label="Name of the person who is to be greeted"
					placeholder="Name of person to greet"
					value="{{ .Value }}">
				<div class="input-group-append">
				<button
					treetop-submitter
					name="submitter"
					value="Greet Me"
					type="button"
					tabindex="1"
					class="btn btn-outline-secondary">Greet Me</button>
				</div>
			</div>
		</form>

		{{ template "message" .Message}}
	</div>
	`, title, base)
}

type GreetTemplateData struct {
	Who   string
	Notes string
}

func GreetingHTML(base string) string {
	return fmt.Sprintf(`
	<div id="message" class="mt-4 text-center">
		<h1>Hello, {{ .Who }}!</h1>
		<p><a href="%s" treetop>Clear</a></p>

		<div class="alert alert-info" role="alert">
			{{ .Notes }}
		</div>

		<div class="alert alert-secondary" role="alert">
			Try refreshing the page or using the back button
		</div>
	</div>

	`, base)
}

func getGreetingQuery(req *http.Request) GreetTemplateData {
	query := req.URL.Query()
	data := GreetTemplateData{
		Who: query.Get("name"),
	}
	if !treetop.IsTemplateRequest(req) {
		data.Notes = "Full page request!"
	} else if submitter := query.Get("submitter"); submitter != "" {
		data.Notes = fmt.Sprintf("XHR form submit with the '%s' button submitter!", submitter)
	} else {
		data.Notes = "XHR form submit, notice that the text input focus is preserved."
	}
	if data.Who == "" {
		data.Who = "World"
	}
	return data
}
