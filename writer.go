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
	DesignatePageURL(string)
	ReplacePageURL(string)
}

type treetopWriter struct {
	responseWriter  http.ResponseWriter
	status          int
	responseURL     string
	replaceURLState bool
	contentType     string
	written         bool
}

func (t *treetopWriter) Status(code int) {
	t.status = code
}

func (t *treetopWriter) DesignatePageURL(uri string) {
	t.responseURL = uri
	t.replaceURLState = false
}

func (t *treetopWriter) ReplacePageURL(uri string) {
	t.responseURL = uri
	t.replaceURLState = true
}

func (tw *treetopWriter) Write(p []byte) (n int, err error) {
	respURI, err := url.Parse(tw.responseURL)
	if err != nil {
		return n, err
	}
	if !tw.written {
		tw.responseWriter.Header().Set("X-Response-Url", respURI.EscapedPath())
		if tw.replaceURLState {
			tw.responseWriter.Header().Set("X-Response-History", "replace")
		}
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
	accept := strings.Split(req.Header.Get("Accept"), ";")[0]
	for _, accept := range strings.Split(accept, ",") {
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
