package treetop

import "net/http"

func Noop(_ DataWriter, _ *http.Request) {}

func Constant(data interface{}) HandlerFunc {
	return func(dw DataWriter, _ *http.Request) {
		dw.Data(data)
	}
}
