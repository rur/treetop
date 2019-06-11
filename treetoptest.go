package treetop

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
)

// Helper for testing individual handler functions in isolation
// experimental feature, not documented or considered stable
func Test(handler HandlerFunc, render func(interface{}) []byte, req *http.Request) *httptest.ResponseRecorder {
	w := &httptest.ResponseRecorder{
		Body: new(bytes.Buffer),
	}
	rsp := responseImpl{
		ResponseWriter: w,
		responseID:     999,
		partial: &Partial{
			HandlerFunc: handler,
		},
	}
	var buf bytes.Buffer
	rsp.execute(&buf, func(w io.Writer, _ []string, data interface{}) error {
		w.Write(render(data))
		return nil
	}, req)

	if rsp.status > 0 {
		w.WriteHeader(rsp.status)
	}
	buf.WriteTo(w)
	return w
}
