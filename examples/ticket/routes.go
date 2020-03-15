package ticket

import (
	"net/http"

	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/assets"
)

// Routes register routes for /view example endpoint
func Routes(mux *http.ServeMux) {
	page := treetop.NewView("base.html", treetop.Delegate("content"))
	_ = page.NewDefaultSubView("nav", "nav.html", treetop.Noop)
	content := page.NewSubView("content", "content.html", ticketContentHandler)

	exec := treetop.NewKeyedStringExecutor(map[string]string{
		"base.html":    assets.BaseHTML,
		"nav.html":     assets.NavHTML(assets.TicketsNav),
		"content.html": ContentHTML,
	})

	mux.Handle("/ticket", exec.NewViewHandler(content))

	if errs := exec.FlushErrors(); len(errs) != 0 {
		panic(errs.Error())
	}
}

// ticketContentHandler
// extends: base.html{content}
func ticketContentHandler(rsp treetop.Response, req *http.Request) interface{} {
	// TODO: Implement this
	return nil
}
