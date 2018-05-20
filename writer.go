package treetop

import (
	"net/http"
	"net/url"
	"strings"
)

type TreetopWriter interface {
	Write([]byte) (int, error)
	Status(int)
	ResponseURI(string)
}

type treetopWriter struct {
	responseWriter http.ResponseWriter
	status         int
	responseURI    string
	contentType    string
}

func (t *treetopWriter) Status(code int) {
	t.status = code
}

func (t *treetopWriter) ResponseURI(uri string) {
	t.responseURI = uri
}

func (tw *treetopWriter) Write(p []byte) (n int, err error) {
	respURI, err := url.Parse(tw.responseURI)
	if err != nil {
		return n, err
	}
	tw.responseWriter.Header().Set("X-Response-Url", respURI.EscapedPath())
	tw.responseWriter.Header().Set("Content-Type", tw.contentType)
	if tw.status > 100 {
		tw.responseWriter.WriteHeader(tw.status)
	}
	return tw.responseWriter.Write(p)
}

func PartialWriter(w http.ResponseWriter, req *http.Request) (TreetopWriter, bool) {
	var isPartial bool
	for _, accept := range strings.Split(req.Header.Get("Accept"), ",") {
		if strings.TrimSpace(accept) == PartialContentType {
			isPartial = true
			break
		}
	}
	if !isPartial {
		return nil, false
	}
	return &treetopWriter{w, 200, req.URL.RequestURI(), PartialContentType}, true
}

func FragmentWriter(w http.ResponseWriter, req *http.Request) (TreetopWriter, bool) {
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
	return &treetopWriter{w, 200, req.URL.RequestURI(), FragmentContentType}, true
}
