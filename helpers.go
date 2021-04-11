package treetop

import (
	"net/http"
	"strings"
)

// Noop treetop handler helper is useful when a treetop.HandlerFunc instance is needed
// but you don't want it to do anything. Function returns `nil`.
func Noop(_ Response, _ *http.Request) interface{} { return nil }

// Constant treetop handler helper is used to generate a treetop.HandlerFunc that always
// returns the same value.
func Constant(data interface{}) ViewHandlerFunc {
	return func(rsp Response, _ *http.Request) interface{} {
		return data
	}
}

// Delegate handler helper will delegate partial handling to a named block of that
// partial. The designated block data will be adopted as the partial template data
// and no other block handler will be executed.
func Delegate(blockname string) ViewHandlerFunc {
	return func(rsp Response, req *http.Request) interface{} {
		return rsp.HandleSubView(blockname, req)
	}
}

// RequestHandler handler helper is used where only the http.Request instance is needed
// to resolve the template data so the treetop.Response isn't part of the actual handler function.
func RequestHandler(f func(*http.Request) interface{}) ViewHandlerFunc {
	return func(_ Response, req *http.Request) interface{} {
		return f(req)
	}
}

// IsTemplateRequest is a predicate function which will check the headers of a given request
// and return true if a template response is supported by the client.
func IsTemplateRequest(req *http.Request) bool {
	for _, accept := range strings.Split(req.Header.Get("Accept"), ";") {
		if strings.ToLower(strings.TrimSpace(accept)) == TemplateContentType {
			return true
		}
	}
	return false
}

// Redirect is a helper that will instruct the Treetop client library to direct the web browser
// to a new URL. If the request is not from a Treetop client, the 3xx redirect method is used.
//
// This is necessary because 3xx HTTP redirects are opaque to XHR, when a full browser redirect
// is needed a 'X-Treetop-Redirect' header is used.
//
// Example:
// 		treetop.Redirect(w, req, "/some/other/path", http.StatusSeeOther)
//
func Redirect(w http.ResponseWriter, req *http.Request, location string, status int) {
	if IsTemplateRequest(req) {
		w.Header().Add("X-Treetop-Redirect", "SeeOther")
		http.Redirect(w, req, location, 200) // must be 200 because XHR cannot intercept a 3xx redirect
	} else {
		http.Redirect(w, req, location, status)
	}
}
