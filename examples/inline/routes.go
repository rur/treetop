package inline

import (
	"net/http"

	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/assets"
)

func Routes(mux *http.ServeMux) {
	page := treetop.NewView("local://base.html", treetop.Delegate("content"))
	_ = page.NewDefaultSubView("nav", "local://nav.html", treetop.Noop)
	content := page.NewSubView("content", "examples/inline/templates/content.html.tmpl", ticketContentHandler)

	exec := &treetop.DeveloperExecutor{
		ViewExecutor: &treetop.FileExecutor{
			KeyedString: map[string]string{
				"local://base.html": assets.BaseHTML,
				"local://nav.html":  assets.NavHTML(assets.InlineNav),
			},
		},
	}

	mux.Handle("/inline", exec.NewViewHandler(content).PageOnly())

	if errs := exec.FlushErrors(); len(errs) != 0 {
		panic(errs.Error())
	}
}

// ticketContentHandler
// extends: base.html{content}
func ticketContentHandler(rsp treetop.Response, req *http.Request) interface{} {
	// TODO: implement
	return nil
}
