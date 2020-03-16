package ticket

import (
	"net/http"

	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/assets"
)

// Routes register routes for /view example endpoint
func Routes(mux *http.ServeMux) {
	page := treetop.NewView("local://base.html", treetop.Delegate("content"))
	_ = page.NewDefaultSubView("nav", "local://nav.html", treetop.Noop)
	content := page.NewSubView("content", "examples/ticket/templates/content.html.tmpl", ticketContentHandler)
	form := content.NewDefaultSubView("form", "examples/ticket/templates/form.html.tmpl", formHandler)

	var exec treetop.ViewExecutor
	exec = &treetop.FileExecutor{
		KeyedString: map[string]string{
			"local://base.html": assets.BaseHTML,
			"local://nav.html":  assets.NavHTML(assets.TicketsNav),
		},
	}

	exec = &treetop.DeveloperExecutor{ViewExecutor: exec}

	mux.Handle("/ticket/form", exec.NewViewHandler(form).FragmentOnly())
	mux.Handle("/ticket", exec.NewViewHandler(content).PageOnly())

	if errs := exec.FlushErrors(); len(errs) != 0 {
		panic(errs.Error())
	}
}

// ticketContentHandler
// extends: base.html{content}
func ticketContentHandler(rsp treetop.Response, req *http.Request) interface{} {
	data := struct {
		Form interface{}
	}{
		Form: rsp.HandleSubView("form", req),
	}
	return data
}

// formHandler
// extends: content.html{form}
func formHandler(rsp treetop.Response, req *http.Request) interface{} {
	// TODO: Implement this
	return nil
}
