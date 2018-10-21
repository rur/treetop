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
func nextResponseId() uint32 {
	return atomic.AddUint32(&token, 1)
}

type Partial struct {
	Extends     string
	Template    string
	HandlerFunc HandlerFunc
	Blocks      []Partial
}

type Handler struct {
	// partial request template+handler dependency tree
	Partial *Partial
	// full page request template+handler dependency tree
	Page *Partial
	// Handlers that will be appended to response *only* for a partial request
	Postscript []Partial
	// Function that will be responsible for executing template contents against
	// data yielded from handlers
	Renderer TemplateExec
}

// implement http.Handler interface, see https://golang.org/pkg/net/http/?#Handler
func (h *Handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var part *Partial
	var contentType string
	if IsTreetopRequest(req) {
		part = h.Partial
		if h.Partial == nil {
			// this is a page only handler, do not accept partial requests
			http.Error(resp, http.StatusText(http.StatusNotAcceptable), http.StatusNotAcceptable)
			return
		} else if h.Page == nil {
			contentType = FragmentContentType
		} else {
			contentType = PartialContentType
		}
	} else if h.Page == nil {
		// TODO: Consider allowing a '303 See Other' redirect to be configured
		http.Error(resp, http.StatusText(http.StatusNotAcceptable), http.StatusNotAcceptable)
		return
	} else {
		part = h.Page
		contentType = "text/html"
	}

	dw := &dataWriter{
		writer:     resp,
		responseId: nextResponseId(),
		partial:    part,
	}

	// Topo-sort of templates connected via blocks. The order is important for how template inheritance is resolved.
	// TODO: The result should not change between requests so cache it when the handler instance is created.
	templates, err := part.TemplateList()
	if err != nil {
		log.Printf(err.Error())
		http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// executes data handlers
	part.HandlerFunc(dw, req)
	if dw.responseWritten {
		// response headers were already sent by one of the handlers, nothing left to do
		return
	}

	// TODO: use buffer pool
	var buf bytes.Buffer
	if tplErr := h.Renderer(&buf, templates, dw.data); tplErr != nil {
		log.Printf(tplErr.Error())
		http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

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

func (h *Handler) Include(defs ...PartialDef) *Handler {
	// Create a new handler which incorporates the templates from the supplied partial definition
	newHandler := Handler{
		h.Partial,
		h.Page,
		h.Postscript,
		h.Renderer,
	}
	for _, def := range defs {
		defH := def.FragmentHandler()
		if newPartial := insertPartial(newHandler.Partial, defH.Partial); newPartial != nil {
			newHandler.Partial = newPartial
		} else {
			// add it to postscript
			newHandler.Postscript = append(newHandler.Postscript, *defH.Partial)
		}
		if newPage := insertPartial(newHandler.Page, defH.Partial); newPage != nil {
			newHandler.Page = newPage
		}
	}
	return &newHandler
}

// obtain a list of all partial templates dependent through block associations, sorted topologically
func (p *Partial) TemplateList() ([]string, error) {
	tpls, err := aggregateTemplates(p.Blocks, p.Extends)
	if err != nil {
		return nil, err
	}
	tpls = append([]string{p.Template}, tpls...)

	return tpls, nil
}

// ---------
// Internal
// ---------

func aggregateTemplates(partials []Partial, seen ...string) ([]string, error) {
	var these []string
	var next []string
	for i := 0; i < len(partials); i++ {
		if contains(seen, partials[i].Extends) {
			return nil, fmt.Errorf(
				"aggregateTemplates: Encountered naming cycle within nested blocks:\n* %s",
				strings.Join(append(seen, partials[i].Extends), " -> "),
			)
		}
		agg, err := aggregateTemplates(partials[i].Blocks, append(seen, partials[i].Extends)...)
		if err != nil {
			return agg, err
		}
		if partials[i].Template != "" {
			these = append(these, partials[i].Template)
		}
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

func insertPartial(parent, child *Partial, seen ...string) *Partial {
	// Create a copy of the parent with the child partial incorporated into the template hierarchy, if possible.
	// If the child partial does not match any blocks in the hierarchy, a nil pointer will be returned.
	copy := Partial{
		parent.Extends,
		parent.Template,
		parent.HandlerFunc,
		make([]Partial, len(parent.Blocks)),
	}
	found := false
	for i := 0; i < len(parent.Blocks); i++ {
		sub := parent.Blocks[i]
		if contains(seen, sub.Extends) {
			// block naming cycle encountered, a combined partial cannot be produced.
			return nil
		}
		if sub.Extends == child.Extends {
			found = true
			copy.Blocks[i] = *child
		} else if updated := insertPartial(&sub, child, append(seen, sub.Extends)...); updated != nil {
			found = true
			copy.Blocks[i] = *updated
		} else {
			copy.Blocks[i] = sub
		}
	}
	if found {
		return &copy
	} else {
		return nil
	}
}
