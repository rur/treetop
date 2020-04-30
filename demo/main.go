package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"

	"github.com/rur/treetop"

	"os"

	"github.com/rur/treetop/demo/assets"
	"github.com/rur/treetop/demo/greeter"
	"github.com/rur/treetop/demo/inline"
	"github.com/rur/treetop/demo/ticket"
)

var portReg = regexp.MustCompile(`^\d+$`)

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

	devMode := true
	port := "3000"
	for i, arg := range os.Args[1:] {
		if arg == "--prod" {
			devMode = false
		} else if portReg.MatchString(arg) {
			port = arg
		} else {
			log.Fatalf("Invalid argument [%d] '%s'", i, arg)
		}
	}

	// Register routes for demo apps
	greeter.Setup(mux, devMode)
	inline.Setup(mux, devMode)
	ticket.Setup(mux, devMode)

	infoSetup(mux, devMode)

	fmt.Printf("serving on http://0.0.0.0:%s/\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

// infoSetup will create template, handlers and bind to routes for the demo landing page
func infoSetup(mux *http.ServeMux, devMode bool) {
	// define handler for home page
	var exec treetop.ViewExecutor = &treetop.FileExecutor{
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
	}
	if devMode {
		exec = &treetop.DeveloperExecutor{
			ViewExecutor: exec,
		}
	}

	base := treetop.
		NewView("local://base.html", treetop.Noop)
	base.NewDefaultSubView("nav", "local://nav.html", treetop.Noop)

	home := base.NewSubView("content", "demo/intro.html", treetop.Noop)
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
		log.Fatalf("Error(s) loading demo templates:\n%s", errs)
	}
}
