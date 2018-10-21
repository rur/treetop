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
	renderer := treetop.NewRenderer(treetop.StringTemplateExec)
	page := renderer.Define(base, baseHandler)
	messsage := page.Block("message")
	greetForm := messsage.Extend(landing, treetop.Noop)
	greetMessage := messsage.Extend(greeting, greetingHandler)

	http.Handle("/", greetForm.PartialHandler())
	http.Handle("/greet", greetMessage.PartialHandler())
	fmt.Println("serving on http://0.0.0.0:3000/")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func baseHandler(w treetop.DataWriter, req *http.Request) {
	msg, _ := w.BlockData("message", req)
	w.Data(struct {
		Message interface{}
	}{
		Message: msg,
	})
}

func greetingHandler(w treetop.DataWriter, req *http.Request) {
	w.Data(req.URL.Query().Get("name"))
}
