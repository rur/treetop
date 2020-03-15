package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/rur/treetop/examples/assets"
	"github.com/rur/treetop/examples/greeter"
	"github.com/rur/treetop/examples/ticket"
	"github.com/rur/treetop/examples/turing"
	"github.com/rur/treetop/examples/writer"
)

var (
	home = template.Must(
		template.Must(
			template.New("base").
				Parse(assets.BaseHTML),
		).New("nav").
			Parse(assets.NavHTML(assets.HomeNav)))
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/favicon.ico", func(_ http.ResponseWriter, _ *http.Request) { /* noop */ })
	mux.HandleFunc("/treetop.js", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(assets.TreetopJS)))
		io.WriteString(w, assets.TreetopJS)
	})
	greeter.Routes(mux)
	writer.Routes(mux)
	turing.Routes(mux)
	ticket.Routes(mux)
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		home.ExecuteTemplate(w, "base", nil)
	})
	fmt.Println("serving on http://0.0.0.0:3000/")
	log.Fatal(http.ListenAndServe(":3000", mux))
}
