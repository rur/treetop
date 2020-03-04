package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/rur/treetop/examples/shared"

	"github.com/rur/treetop/examples/view"
	"github.com/rur/treetop/examples/writer"
)

var (
	baseTmpl = template.Must(template.New("base").Parse(shared.BaseTemplate))
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/favicon.ico", func(_ http.ResponseWriter, _ *http.Request) { /* noop */ })
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		baseTmpl.Execute(w, nil)
	})
	mux.HandleFunc("/treetop.js", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(shared.TreetopJS)))
		io.WriteString(w, shared.TreetopJS)
	})
	view.SetupGreeter(mux)
	writer.SetupGreeter(mux)
	fmt.Println("serving on http://0.0.0.0:3000/")
	log.Fatal(http.ListenAndServe(":3000", mux))
}
