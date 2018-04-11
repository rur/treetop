package treetop

import (
	"net/http"
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

func ForcedRedirect(dw DataWriter, req *http.Request, url string, statusCode int) {
	var isXHR bool
	for _, accept := range strings.Split(req.Header.Get("Accept"), ",") {
		if strings.Trim(accept, " ") == FragmentContentType || strings.Trim(accept, " ") == PartialContentType {
			isXHR = true
			break
		}
	}
	if isXHR {
		// This is an XMLHttpRequest, force redirect
		dw.Header().Set("X-Response-Url", url)
		dw.WriteHeader(http.StatusNoContent)
	} else {
		// normal http redirect
		http.Redirect(dw, req, url, statusCode)
	}
}
