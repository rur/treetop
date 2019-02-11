package treetop

import (
	"net/http"
	"net/url"
	"strings"
)

const (
	PartialContentType  = "application/x.treetop-html-partial+xml"
	FragmentContentType = "application/x.treetop-html-fragment+xml"
)

type TreetopWriter interface {
	Write([]byte) (int, error)
	Status(int)
	ResponseURI(string)
}

type treetopWriter struct {
	responseWriter http.ResponseWriter
	status         int
	responseURL    string
	contentType    string
	written        bool
}

func (t *treetopWriter) Status(code int) {
	t.status = code
}

func (t *treetopWriter) ResponseURI(uri string) {
	t.responseURL = uri
}

func (tw *treetopWriter) Write(p []byte) (n int, err error) {
	respURI, err := url.Parse(tw.responseURL)
	if err != nil {
		return n, err
	}
	if !tw.written {
		tw.responseWriter.Header().Set("X-Response-Url", respURI.EscapedPath())
		tw.responseWriter.Header().Set("Content-Type", tw.contentType)
		if tw.status > 100 {
			tw.responseWriter.WriteHeader(tw.status)
		}
		tw.written = true
	}
	return tw.responseWriter.Write(p)
}

func Writer(w http.ResponseWriter, req *http.Request, isPartial bool) (TreetopWriter, bool) {
	var contentType string
	if isPartial {
		contentType = PartialContentType
	} else {
		contentType = FragmentContentType
	}
	var writer *treetopWriter
	for _, accept := range strings.Split(req.Header.Get("Accept"), ",") {
		if strings.TrimSpace(accept) == contentType {
			writer = &treetopWriter{
				responseWriter: w,
				responseURL:    req.URL.RequestURI(),
				contentType:    contentType,
			}
			break
		}
	}
	return writer, (writer != nil)
}
