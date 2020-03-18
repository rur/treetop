package inline

import (
	"fmt"
	"net/http"
	"time"

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

const formDataCookieName = "inline-form-data"

// ticketContentHandler
// extends: base.html{content}
func ticketContentHandler(rsp treetop.Response, req *http.Request) interface{} {
	data := struct {
		FormData *FormData
	}{
		FormData: &FormData{},
	}
	// get or set an encoded cookie to simulate persistance
	if cookie, err := req.Cookie(formDataCookieName); err == nil {
		err := data.FormData.UnmarshalBase64([]byte(cookie.Value))
		if err != nil {
			http.Error(
				rsp,
				fmt.Sprintf("Error unmarshalling cookie base64 data, %s", err),
				http.StatusInternalServerError,
			)
			return nil
		}
	} else {
		// create default form state for the demo and
		// set an encoded cookie value
		data.FormData = getDefaultFormData()
		b64, err := data.FormData.MarshalBase64()
		if err != nil {
			http.Error(
				rsp,
				fmt.Sprintf("Error unmarshalling cookie base64 data, %s", err),
				http.StatusInternalServerError,
			)
			return nil
		}
		cookie := http.Cookie{
			Name:     formDataCookieName,
			Path:     "/inline",
			Value:    string(b64),
			Expires:  time.Now().Add(1 * time.Hour),
			HttpOnly: true,
		}
		http.SetCookie(rsp, &cookie)
	}

	return data
}
