package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/rur/treetop"

	"github.com/rur/treetop/examples/assets"
	"github.com/rur/treetop/examples/greeter"
	"github.com/rur/treetop/examples/inline"
	"github.com/rur/treetop/examples/ticket"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "https://golang.org/favicon.ico", http.StatusSeeOther)
	})
	mux.HandleFunc("/treetop.js", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(assets.TreetopJS)))
		io.WriteString(w, assets.TreetopJS)
	})

	// Register routes for example apps
	greeter.Routes(mux)
	inline.Routes(mux)
	ticket.Routes(mux)

	// define handler for home page
	exec := treetop.NewKeyedStringExecutor(map[string]string{
		"local://base.html": assets.BaseHTML,
		"local://nav.html":  assets.NavHTML(assets.IntroNav),
	})

	home := treetop.
		NewView("local://base.html", treetop.Noop).
		NewSubView("nav", "local://nav.html", treetop.Noop)

	mux.Handle("/", exec.NewViewHandler(home).PageOnly())

	if errs := exec.FlushErrors(); len(errs) > 0 {
		log.Fatalf("Error(s) loading example templates:\n%s", errs)
	}

	fmt.Println("serving on http://0.0.0.0:3000/")
	log.Fatal(http.ListenAndServe(":3000", mux))
}
