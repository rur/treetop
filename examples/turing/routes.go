package turing

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/assets"
)

// Routes register routes for /view example endpoint
func Routes(mux *http.ServeMux) {
	page := treetop.NewView("base.html", treetop.Delegate("content"))
	_ = page.NewDefaultSubView("nav", "nav.html", treetop.Noop)
	content := page.NewSubView("content", "content.html", turingContentHandler)

	exec := treetop.NewKeyedStringExecutor(map[string]string{
		"base.html":    assets.BaseHTML,
		"nav.html":     assets.NavHTML(assets.TuringNav),
		"content.html": ContentHTML,
	})

	mux.Handle("/turing", exec.NewViewHandler(content))

	if errs := exec.FlushErrors(); len(errs) != 0 {
		panic(errs.Error())
	}
}

// turingContentHandler
// extends: base.html{"content"}
func turingContentHandler(rsp treetop.Response, req *http.Request) interface{} {
	data := struct {
		Data         string
		Compressed   string
		ErrorMessage string
	}{}

	msgs := Messages{
		Hello:  "World",
		Number: 49,
	}

	// generate some data to test compression
	for i := 2 << 10; i < 2<<11; i++ {
		msgs.Data = append(msgs.Data, float64(i))
	}

	if d, err := json.Marshal(msgs); err != nil {
		rsp.Status(http.StatusInternalServerError)
		data.ErrorMessage = fmt.Sprintf("Somthing went wrong: %s", err)
		return data
	} else {
		data.Data = string(d)
	}

	// now comress
	if b64, err := msgs.MarshalBase64(); err != nil {
		data.ErrorMessage = fmt.Sprintf("Something went wrong with the marshal: %s", err)
		return data
	} else {
		data.Compressed = fmt.Sprintf("Size bytes %d: %s", len(b64), b64)
	}

	return data
}
