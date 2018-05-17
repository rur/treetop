package treetop

import (
	"net/http"
	"strings"
)

type PartialWriter interface {
	Write([]byte) (int, error)
	Status(int)
}

type treetopWriter struct {
	responseWriter http.ResponseWriter
	status         int
}

func (t *treetopWriter) Status(code int) {
	t.status = code
}

func (tw *treetopWriter) Write(p []byte) (n int, err error) {
	tw.responseWriter.Header().Set("Content-Type", FragmentContentType)
	if tw.status > 100 {
		tw.responseWriter.WriteHeader(tw.status)
	}
	return tw.responseWriter.Write(p)
}

func Writer(w http.ResponseWriter, req *http.Request) (PartialWriter, bool) {
	var isPartial bool
	for _, accept := range strings.Split(req.Header.Get("Accept"), ",") {
		if strings.TrimSpace(accept) == FragmentContentType {
			isPartial = true
			break
		}
	}
	if !isPartial {
		return nil, false
	}
	return &treetopWriter{w, 200}, true
}
