package treetop

import (
	"net/http"
	"net/url"
	"strings"
)

const (
	// PartialContentType Content-Type header constant denoting a part of a full page,
	// the response URL can be used to load a full page
	PartialContentType = "application/x.treetop-html-partial+xml"
	// FragmentContentType Content-Type header constant denoting a fragment of a web page,
	// the response URL cannot necessarily be used to load a full page
	FragmentContentType = "application/x.treetop-html-fragment+xml"
)

// Writer is an interface for used ad-hoc Treetop response writing
// this can be used for a regular http handler function to send a fragment
// response based upon the request header
type Writer interface {
	Write([]byte) (int, error)
	Status(int)
	DesignatePageURL(string)
	ReplacePageURL(string)
}

// writer wraps a http.ResponseWriter instance, to set the appropriate
// headers based upon the treetop prototcol
type writer struct {
	responseWriter  http.ResponseWriter
	status          int
	responseURL     string
	replaceURLState bool
	contentType     string
	written         bool
}

func (tw *writer) Status(code int) {
	tw.status = code
}

func (tw *writer) DesignatePageURL(uri string) {
	tw.responseURL = uri
	tw.replaceURLState = false
}

func (tw *writer) ReplacePageURL(uri string) {
	tw.responseURL = uri
	tw.replaceURLState = true
}

func (tw *writer) Write(p []byte) (n int, err error) {
	respURI, err := url.Parse(tw.responseURL)
	if err != nil {
		return n, err
	}
	if !tw.written {
		tw.responseWriter.Header().Set("X-Response-Url", respURI.String())
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

// NewPartialWriter will check if the client accepts one of the Treetop content types,
// if so it will return a wrapped response writer for a Treetop html partial.
//
// Note that this will render a Treetop 'partial' content type, which by the protocol means that on subsequence requests
// the client can expect to be able to load a full HTML document by varying the accept header.
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
	var ttW *writer
	accept := strings.Split(req.Header.Get("Accept"), ";")[0]
	for _, accept := range strings.Split(accept, ",") {
		if strings.TrimSpace(accept) == PartialContentType {
			ttW = &writer{
				responseWriter: w,
				responseURL:    req.URL.RequestURI(),
				contentType:    PartialContentType,
			}
			break
		}
	}
	return ttW, (ttW != nil)
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
	accept := strings.Split(req.Header.Get("Accept"), ";")[0]
	for _, accept := range strings.Split(accept, ",") {
		if strings.TrimSpace(accept) == FragmentContentType {
			ttW = &writer{
				responseWriter: w,
				responseURL:    req.URL.RequestURI(),
				contentType:    FragmentContentType,
			}
			break
		}
	}
	return ttW, (ttW != nil)
}
