package treetop

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func Noop(_ DataWriter, _ *http.Request) {}

func Constant(data interface{}) HandlerFunc {
	return func(dw DataWriter, _ *http.Request) {
		dw.Data(data)
	}
}

func RequestHandler(f func(*http.Request) interface{}) HandlerFunc {
	return func(dw DataWriter, req *http.Request) {
		dw.Data(f(req))
	}
}

func IsTreetopRequest(req *http.Request) bool {
	for _, accept := range strings.Split(req.Header.Get("Accept"), ",") {
		if strings.TrimSpace(accept) == FragmentContentType {
			return true
		}
		if strings.TrimSpace(accept) == PartialContentType {
			return true
		}
	}
	return false
}

func FullRedirect(w http.ResponseWriter, req *http.Request, uri string, statusCode int) {
	if IsTreetopRequest(req) {
		// This is an XMLHttpRequest, force redirect
		redir, err := url.Parse(uri)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("")
			return
		}

		w.Header().Set("X-Treetop-Redirect", redir.EscapedPath())
		w.WriteHeader(http.StatusNoContent)
		if req.Method == "GET" {
			note := "<a href=\"" + html.EscapeString(redir.EscapedPath()) + "\">Redirect</a>.\n"
			fmt.Fprintln(w, note)
		}
	} else {
		http.Redirect(w, req, uri, statusCode)
	}
}

func Delegate(block string) HandlerFunc {
	return func(dw DataWriter, req *http.Request) {
		data, _ := dw.PartialData(block, req)
		dw.Data(data)
	}
}
