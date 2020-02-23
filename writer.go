package treetop

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	// TemplateContentType is used for content negotiation within template requests
	TemplateContentType = "application/x.treetop-html-template+xml"
)

// Writer is an interface for writing HTTP responses that conform to the Treetop protocol
type Writer interface {
	io.Writer
	Status(int)
	DesignatePageURL(string)
	ReplacePageURL(string)
}

// writer wraps a http.ResponseWriter instance, to set the appropriate
// headers based upon the treetop prototcol
type writer struct {
	responseWriter    http.ResponseWriter
	status            int
	responseURLExists bool
	responseURL       string
	replaceURLState   bool
	written           bool
}

// Status will make a record of what the HTTP status code should be when the response
// headers are written. If this is not set, the fallback will be the default
// http.ResponseWriter behavior.
//
// Note, calling this after a response headers have already been written will have
// no effect.
func (tw *writer) Status(code int) {
	tw.status = code
}

// DesignatePageURL specifies a URL that the client should use to add a
// navigation entry to the browser history
//
// Note, calling this after a response headers have already been written will have
// no effect.
func (tw *writer) DesignatePageURL(uri string) {
	tw.responseURL = uri
	tw.responseURLExists = true
	tw.replaceURLState = false
}

// ReplacePageURL specifies a URL that the client should use to 'replace' the current
// navigation history entry
//
// Note, calling this after a response headers have already been written will have
// no effect.
func (tw *writer) ReplacePageURL(uri string) {
	tw.responseURL = uri
	tw.responseURLExists = true
	tw.replaceURLState = true
}

// Write will add the necessary headers to the HTTP response
// and output the supplied bytes in the response body
func (tw *writer) Write(p []byte) (n int, err error) {
	if tw.written {
		return tw.responseWriter.Write(p)
	}
	if tw.responseURLExists {
		respURI, err := url.Parse(tw.responseURL)
		if err != nil {
			return 0, err
		}
		tw.responseWriter.Header().Set("X-Page-URL", respURI.RequestURI())
		if tw.replaceURLState {
			tw.responseWriter.Header().Set("X-Response-History", "replace")
		}
	}
	tw.responseWriter.Header().Set("Content-Type", TemplateContentType)
	if tw.status > 0 {
		tw.responseWriter.WriteHeader(tw.status)
	}
	tw.written = true
	return tw.responseWriter.Write(p)
}

// NewPartialWriter will check if the client accepts the template content type.
// If so it will return a wrapped response writer that will add the appropriate headers.
//
// The partial writer will include the 'X-Page-URL' response header with the URI of the request.
// By the protocol, this means that it must be possible for a subsequent request
// to load a full HTML document from that URL by varying the accept header.
//
// Example:
//
// 	func MyHandler(w http.ResponseWriter, req *http.Request) {
// 		if ttW, ok := treetop.NewPartialWriter(w, req); ok {
// 			// this is a treetop request, write a HTML fragment
// 			fmt.Fprintf(tw, `<h3 id="greeting">Hello, %s</h3>`, "Treetop")
// 			return
// 		}
//		/* otherwise render a full HTML page as normal */
// 	}
//
func NewPartialWriter(w http.ResponseWriter, req *http.Request) (Writer, bool) {
	ttW, ok := NewFragmentWriter(w, req)
	if ok {
		ttW.DesignatePageURL(req.URL.RequestURI())
	}
	return ttW, ok
}

// NewFragmentWriter will check if the client accepts one of the Treetop content types,
// if so it will return a wrapped response writer for a Treetop html fragment.
//
// Example:
//
// 	func MyHandler(w http.ResponseWriter, req *http.Request) {
// 		if ttW, ok := treetop.NewFragmentWriter(w, req); ok {
// 			// this is a treetop request, write a HTML fragment
// 			fmt.Fprintf(tw, `<h3 id="greeting">Hello, %s</h3>`, "Treetop")
// 			return
// 		}
//		/* otherwise handle request in a different way (unspecified) */
// 	}
//
func NewFragmentWriter(w http.ResponseWriter, req *http.Request) (Writer, bool) {
	var ttW *writer
	for _, accept := range strings.Split(req.Header.Get("Accept"), ";") {
		if strings.ToLower(strings.TrimSpace(accept)) == TemplateContentType {
			ttW = &writer{
				responseWriter: w,
			}
			break
		}
	}
	return ttW, (ttW != nil)
}
