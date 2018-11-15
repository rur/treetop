package treetop

import (
	"net/http"
	"strings"
)

func Noop(_ Response, _ *http.Request) {}

func Constant(data interface{}) HandlerFunc {
	return func(rsp Response, _ *http.Request) {
		rsp.Data(data)
	}
}

func Delegate(blockname string) HandlerFunc {
	return func(rsp Response, req *http.Request) {
		data, ok := rsp.Delegate(blockname, req)
		if ok {
			rsp.Data(data)
		}
	}
}

func RequestHandler(f func(*http.Request) interface{}) HandlerFunc {
	return func(rsp Response, req *http.Request) {
		rsp.Data(f(req))
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
