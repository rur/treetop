package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rur/treetop"
)

var (
	base = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Treetop Greeter</title>
</head>
<body style="text-align: center;">
	<h1>Treetop Greeter</h1>
	<div>
		<form action="/greet" treetop onsubmit="setTimeout(this.reset.bind(this), 10)">
			<span>Greet, </span><input placeholder="Name" type="text" name="name">
		</form>
	</div>
{{ block "message" .Message}}{{ end }}
<script>TREETOP_CONFIG={/*defaults*/}</script>
<script src="https://rawgit.com/rur/treetop-client/master/treetop.js" async></script>
</body>
</html>
	`
	landing = `
{{ block "message" .}}
	<p id="message"><i>Give me someone to say hello to!</i></p>
{{ end }}
	`
	greeting = `
{{ block "message" .}}
	<div id="message">
		<h2>Hello, {{ . }}!</h2>
		<p><a href="/" treetop>Clear</a></p>
	</div>
{{ end }}
	`
)

func main() {
	page := treetop.NewView(
		treetop.StringTemplateExec,
		base,
		baseHandler,
	)
	greetForm := page.SubView("message", landing, treetop.Noop)
	greetMessage := page.SubView("message", greeting, greetingHandler)

	http.Handle("/", treetop.ViewHandler(greetForm))
	http.Handle("/greet", treetop.ViewHandler(greetMessage))
	fmt.Println("serving on http://0.0.0.0:3000/")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func baseHandler(rsp treetop.Response, req *http.Request) interface{} {
	return struct {
		Message interface{}
	}{
		Message: rsp.HandlePartial("message", req),
	}
}

func greetingHandler(_ treetop.Response, req *http.Request) interface{} {
	return req.URL.Query().Get("name")
}
