package treetop

import (
	"bytes"
	"context"
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
func nextResponseID() uint32 {
	return atomic.AddUint32(&token, 1)
}

func ViewHandler(view View, includes ...View) *Handler {
	// Create a new handler which incorporates the templates from the supplied partial definition
	newHandler := view.Handler()

	// Includes allows one or more unrelated View configurations to be combined with the primary
	// handler instance. Unrelated parts of the page can be rendered with the same handler
	// and existing blocks in the current Handler can be further shadowed.
	for _, include := range includes {
		iH := include.Handler()
		if newPartial := insertPartial(newHandler.Fragment, iH.Fragment); newPartial != nil {
			newHandler.Fragment = newPartial
		} else {
			// add it to postscript
			newHandler.Postscript = append(newHandler.Postscript, *iH.Fragment)
		}
		if newPage := insertPartial(newHandler.Page, iH.Fragment); newPage != nil {
			newHandler.Page = newPage
		}
	}
	return newHandler
}

// Partial represents a template partial in a treetop Handler
type Partial struct {
	Extends     string
	Template    string
	HandlerFunc HandlerFunc
	Blocks      []Partial
}

// Handler implements http.Handler interface. The struct contains all configuration
// necessary for a hierarchical style treetop endpoint.
type Handler struct {
	// partial request template+handler dependency tree
	Fragment *Partial
	// full page request template+handler dependency tree
	Page *Partial
	// Handlers that will be appended to response *only* for a partial request
	Postscript []Partial
	// Function that will be responsible for executing template contents against
	// data yielded from handlers
	Renderer TemplateExec
}

// Implementation of http.Handler interface, see https://golang.org/pkg/net/http/?#Handler
func (h *Handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	select {
	case <-req.Context().Done():
		// request has already been cancelled, do nothing
		return
	default:
	}
	responseID := nextResponseID()
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel() // Cancel treetop ctx when handler has done it's work.

	var part *Partial
	var contentType string
	if IsTreetopRequest(req) {
		part = h.Fragment
		if h.Fragment == nil {
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

	// TODO: use buffer pool
	var buf bytes.Buffer

	rsp := &responseImpl{
		ResponseWriter: resp,
		context:        ctx,
		responseID:     responseID,
		partial:        part,
	}

	if err := rsp.execute(&buf, h.Renderer, req); err != nil {
		log.Printf(err.Error())
		http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if rsp.finished {
		// http response has already been resolved somehow, do not proceed
		return
	} else {
		// mark response instance as finished
		rsp.finished = true
	}

	resp.Header().Set("Content-Type", contentType)

	if contentType == PartialContentType || contentType == FragmentContentType {
		// Execute postscript templates for partial requests only
		// each rendered template will be appended to the content body.
		for index := 0; index < len(h.Postscript); index++ {
			postRsp := &responseImpl{
				ResponseWriter: resp,
				context:        ctx,
				responseID:     responseID,
				partial:        &h.Postscript[index],
			}

			if err := postRsp.execute(&buf, h.Renderer, req); err != nil {
				log.Printf(err.Error())
				http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			} else if postRsp.finished {
				// http response has already been resolved somehow, do not proceed
				return
			} else {
				// mark response instance as finished
				postRsp.finished = true
			}
		}

		// this is useful for XHR requests because if a redirect occurred
		// the final response URL is not necessarily available to the client
		resp.Header().Set("X-Response-Url", req.URL.RequestURI())
	}

	// Since we are modulating the representation based upon a header value, it is
	// necessary to inform the caches. See https://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html#sec13.6
	resp.Header().Set("Vary", "Accept")

	// if a status code was specified, use it. Otherwise fallback to the net/http default.
	if rsp.status > 0 {
		resp.WriteHeader(rsp.status)
	}
	buf.WriteTo(resp)
}

func (h *Handler) PageOnly() *Handler {
	return &Handler{
		Page:     h.Page,
		Renderer: h.Renderer,
	}
}

func (h *Handler) FragmentOnly() *Handler {
	return &Handler{
		Fragment:   h.Fragment,
		Postscript: append([]Partial{}, h.Postscript...),
		Renderer:   h.Renderer,
	}
}

// TemplateList is used to obtain all partial templates dependent through block
// associations, sorted topologically
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
	}
	return nil
}
