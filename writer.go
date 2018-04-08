package treetop

import (
	"io"
	"net/http"
	"strings"
)

type treetopWriter struct {
	responseWriter http.ResponseWriter
}

func (tw *treetopWriter) Write(p []byte) (n int, err error) {
	tw.responseWriter.Header().Set("Content-Type", FragmentContentType)
	return tw.responseWriter.Write(p)
}

func Writer(w http.ResponseWriter, req *http.Request) (io.Writer, bool) {
	var isPartial bool
	for _, accept := range strings.Split(req.Header.Get("Accept"), ",") {
		if strings.Trim(accept, " ") == FragmentContentType {
			isPartial = true
			break
		}
	}
	if !isPartial {
		return nil, false
	}
	return &treetopWriter{w}, true
}
