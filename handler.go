package treetop

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

var token uint32

// Generate a token which can be used to identify treetop
// responses *locally*. The only uniqueness requirement
// is that concurrent active requests must not possess the same value.
func nextLocalToken() uint32 {
	return atomic.AddUint32(&token, 1)
}

type Partial struct {
	Extends     string
	Template    string
	HandlerFunc HandlerFunc
	Root        *Partial
	Blocks      []*Partial
}

type Handler struct {
	// Pointer within a hierarchy of templates connected via 'blocks'
	Partial *Partial
	// Handlers that will be appended to response *only* for a partial request
	Postscript []*Partial
	// Only accept request for treetop fragment content type (end-point will not serve a full page)
	FragmentOnly bool
	// Function that will be responsible for executing template contents against
	// data yielded from handlers
	Renderer TemplateExec
}

// implement http.Handler interface, see https://golang.org/pkg/net/http/?#Handler
func (h *Handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	dw := &dataWriter{
		writer:     resp,
		localToken: nextLocalToken(),
		partial:    h.Partial,
	}

	var part *Partial
	var contentType string
	if IsTreetopRequest(req) {
		part = h.Partial
		if h.FragmentOnly {
			contentType = FragmentContentType
		} else {
			contentType = PartialContentType
		}
	} else if h.FragmentOnly {
		// TODO: Consider allowing a '303 See Other' redirect to be configured
		http.Error(resp, http.StatusText(http.StatusNotAcceptable), http.StatusNotAcceptable)
		return
	} else {
		if h.Partial.Root != nil {
			part = h.Partial.Root
		} else {
			part = h.Partial
		}
		contentType = "text/html"
	}

	// Topo-sort of templates connected via blocks. The order is important for how template inheritance is resolved.
	// TODO: The result should not change been requests so cache it when the handler instance is created.
	tmpls, err := part.TemplateList()
	if err != nil {
		log.Printf(err.Error())
		http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// executes data handlers
	part.HandlerFunc(dw, req)

	// TODO: use buffer pool
	var buf bytes.Buffer
	h.Renderer(&buf, tmpls, dw.data)

	resp.Header().Set("Content-Type", contentType)

	if contentType == PartialContentType || contentType == FragmentContentType {
		// this is useful for XHR requests because if a redirect occurred
		// the final response URL is not necessarily available to the client
		resp.Header().Set("X-Response-Url", req.URL.RequestURI())
	}

	// Since we are modulating the representation based upon a header value, it is
	// necessary to inform the caches. See https://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html#sec13.6
	resp.Header().Set("Vary", "Accept")

	// if a status code was specified, use it. Otherwise fallback to the net/http default.
	if dw.status > 0 {
		resp.WriteHeader(dw.status)
	}
	buf.WriteTo(resp)
}

// obtain a list of all partial templates dependent through block associations, sorted topologically
func (p *Partial) TemplateList() ([]string, error) {
	tmpls, err := aggregateTemplates([]string{p.Extends}, p.Blocks)
	if err != nil {
		return nil, err
	}
	tmpls = append([]string{p.Template}, tmpls...)

	return tmpls, nil
}

func aggregateTemplates(seen []string, partials []*Partial) ([]string, error) {
	var these []string
	var next []string
	for i := 0; i < len(partials); i++ {
		if contains(seen, partials[i].Extends) {
			return nil, fmt.Errorf(
				"aggregateTemplates: Encountered naming cycle within nested blocks:\n* %s",
				strings.Join(append(seen, partials[i].Extends), " -> "),
			)
		} else {
			seen = append(seen, partials[i].Extends)
		}
		agg, err := aggregateTemplates(seen, partials[i].Blocks)
		if err != nil {
			return agg, err
		}
		these = append(these, partials[i].Template)
		next = append(next, agg...)
	}
	return append(these, next...), nil
}

func contains(values []string, query string) bool {
	for i := 0; i < len(values); i++ {
		if values[i] == query {
			return true
		}
	}
	return false
}
