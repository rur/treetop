package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/rur/treetop/examples/shared"

	"github.com/rur/treetop/examples/view"
	"github.com/rur/treetop/examples/writer"
)

var baseTmpl *template.Template

func init() {
	baseTmpl = template.Must(template.New("base").Parse(shared.BaseTemplate))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/favicon.ico", func(_ http.ResponseWriter, _ *http.Request) { /* noop */ })
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		baseTmpl.Execute(w, nil)
	})
	view.SetupGreeter(mux)
	writer.SetupGreeter(mux)
	fmt.Println("serving on http://0.0.0.0:3000/")
	log.Fatal(http.ListenAndServe(":3000", mux))
}
