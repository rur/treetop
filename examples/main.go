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
	greeter.Setup(mux)
	inline.Setup(mux)
	ticket.Setup(mux)

	infoSetup(mux)

	fmt.Println("serving on http://0.0.0.0:3000/")
	log.Fatal(http.ListenAndServe(":3000", mux))
}

// infoSetup will create template, handlers and bind to routes for the example landing page
func infoSetup(mux *http.ServeMux) {
	// define handler for home page
	exec := &treetop.DeveloperExecutor{
		ViewExecutor: &treetop.FileExecutor{
			KeyedString: map[string]string{
				"local://base.html":     assets.BaseHTML,
				"local://nav-home.html": assets.NavHTML(assets.IntroNav),
				"local://nav.html":      assets.NavHTML(assets.NoPage),
				"local://notfound.html": `
				<div class="text-center">
					<hr class="mb-3">
					<h3>404 Not Found</h3>
					<p>No page was found for this path.</p>
					<p>You can file a (real) issue here <a href="https://github.com/rur/treetop/issues">treetop/issues</a></p>
				</div>
				`,
			},
		},
	}

	base := treetop.
		NewView("local://base.html", treetop.Noop)
	base.NewDefaultSubView("nav", "local://nav.html", treetop.Noop)

	home := base.NewSubView("content", "examples/intro.html", treetop.Noop)
	navHome := base.NewSubView("nav", "local://nav-home.html", treetop.Noop)

	homeHandler := exec.NewViewHandler(home, navHome).PageOnly()
	notFoundHandler := exec.NewViewHandler(
		base.NewSubView("content", "local://notfound.html", treetop.Noop)).
		PageOnly()

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/", "":
			homeHandler.ServeHTTP(w, req)
		default:
			notFoundHandler.ServeHTTP(w, req)
		}
	})

	if errs := exec.FlushErrors(); len(errs) > 0 {
		log.Fatalf("Error(s) loading example templates:\n%s", errs)
	}
}
