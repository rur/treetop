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

func Delegate(blockname string) HandlerFunc {
	return func(dw DataWriter, req *http.Request) {
		data, ok := dw.BlockData(blockname, req)
		if ok {
			dw.Data(data)
		}
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
