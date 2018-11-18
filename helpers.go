package treetop

import (
	"net/http"
	"strings"
)

func Noop(_ Response, _ *http.Request) interface{} { return nil }

func Constant(data interface{}) HandlerFunc {
	return func(rsp Response, _ *http.Request) interface{} {
		return data
	}
}

func Delegate(blockname string) HandlerFunc {
	return func(rsp Response, req *http.Request) interface{} {
		return rsp.HandlePartial(blockname, req)
	}
}

func RequestHandler(f func(*http.Request) interface{}) HandlerFunc {
	return func(_ Response, req *http.Request) interface{} {
		return f(req)
	}
}

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
