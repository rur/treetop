package treetop

import (
	"net/http"
	"strings"
)

// Noop treetop handler helper is useful when a treetop.HandlerFunc instance is needed
// but you dont want it to do anything. Function returns `nil`.
func Noop(_ Response, _ *http.Request) interface{} { return nil }

// Constant treetop handler helper is used to generate a treetop.HandlerFunc that always
// returns the same value.
func Constant(data interface{}) HandlerFunc {
	return func(rsp Response, _ *http.Request) interface{} {
		return data
	}
}

// Delegate handler helper will delegate partial handling to a named block of that
// partial. The designated block data will be adopted as the partial template data
// and no other block handler will be executed.
func Delegate(blockname string) HandlerFunc {
	return func(rsp Response, req *http.Request) interface{} {
		return rsp.HandlePartial(blockname, req)
	}
}

// RequestHandler handler helper is used where only the http.Request instance is needed
// to resolve the template data so the treetop.Response isnt part of the actual handler function.
func RequestHandler(f func(*http.Request) interface{}) HandlerFunc {
	return func(_ Response, req *http.Request) interface{} {
		return f(req)
	}
}

// IsTreetopRequest is a predicate function which will check the headers of a given request
// and return true if a partial (or fragment) response is accepted by the client.
func IsTreetopRequest(req *http.Request) bool {
	for _, accept := range strings.Split(req.Header.Get("Accept"), ",") {
		accept = strings.ToLower(strings.TrimSpace(accept))
		if accept == FragmentContentType {
			return true
		}
		if accept == PartialContentType {
			return true
		}
	}
	return false
}
