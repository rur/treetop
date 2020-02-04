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

// nextResponseID generates a token which can be used to identify treetop
// responses *locally*. The only uniqueness requirement
// is that concurrent active requests must not possess the same value.
func nextResponseID() uint32 {
	return atomic.AddUint32(&token, 1)
}

// ViewHandler returns an instance implementing the http.Handler interface, given a series of treetop.View definitions.
//
// In addition to the primary (and mandatory) view, 'include' views can be added.
//
// Includes will affect request handling in the following way:
//  	- Any subview of the primary view with the same name as the include, will be overloaded;
//		- Otherwise, if no primary subview with the same name exists:
// 			- includes are treated as 'siblings'
// 			- include handlers will be executed *after* all preceding view hierarchies are resolved;
// 			- the rendered include template will be appended to the response fragment.
//
// Note: it is the responsibility of the client library to decide how to handle sibling HTML nodes
//       in the template fragment.
//
func ViewHandler(view *View, includes ...*View) *Handler {
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

// ServeHTTP implements the http.Handler interface, see https://golang.org/pkg/net/http/?#Handler
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

	if contentType == PartialContentType || contentType == FragmentContentType {
		specified := rsp.pageURLSpecified
		pageURL := rsp.pageURL
		resplaceState := rsp.replaceURL
		// Execute postscript templates for partial requests only
		// each rendered template will be appended to the content body.
		for i := range h.Postscript {
			postRsp := &responseImpl{
				ResponseWriter: resp,
				context:        ctx,
				responseID:     responseID,
				partial:        &h.Postscript[i],
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
			if !specified && postRsp.pageURLSpecified {
				specified = true
				pageURL = postRsp.pageURL
				resplaceState = postRsp.replaceURL
			}
		}

		if specified {
			// one or other of the handlers has specified a 'PartialURL'
			// this URL should be used as the page location
			contentType = PartialContentType
			resp.Header().Set("X-Response-Url", pageURL)
			if resplaceState {
				resp.Header().Set("X-Response-History", "replace")
			}
		} else {
			// This is still useful for XHR requests because if a redirect occurred
			// the final response URL is not necessarily available to the client
			resp.Header().Set("X-Response-Url", req.URL.RequestURI())
		}
	}
	resp.Header().Set("Content-Type", contentType)

	// Since we are modulating the representation based upon a header value, it is
	// necessary to inform the caches. See https://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html#sec13.6
	resp.Header().Set("Vary", "Accept")

	// if a status code was specified, use it. Otherwise fallback to the net/http default.
	if rsp.status > 0 {
		resp.WriteHeader(rsp.status)
	}
	buf.WriteTo(resp)
}

// PageOnly restricts the handler end-point from accepting requests for partial data. Only full page
// requests are to be served. Put another way, the Treetop accept header is not supported by this handler.
// Response content type will be `"text/html"`.
func (h *Handler) PageOnly() *Handler {
	return &Handler{
		Page:     h.Page,
		Renderer: h.Renderer,
	}
}

// FragmentOnly restricts the handler end-point from accepting full page requests. Only a
// valid treetop partial request will be acceptable.
// Response content type will be `treetop.FragmentContentType`
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
	tmpl, err := aggregateTemplates(p.Blocks, p.Extends)
	if err != nil {
		return nil, err
	}
	tmpl = append([]string{p.Template}, tmpl...)
	return tmpl, nil
}

// ---------
// Internal
// ---------

// aggregateTemplates traverses a list of partial hierarchies and returns the
// template strings that were used to define the associated sub-views.
func aggregateTemplates(partials []Partial, seen ...string) ([]string, error) {
	var these []string
	var next []string
	for _, partial := range partials {
		if contains(seen, partial.Extends) {
			return nil, fmt.Errorf(
				"aggregateTemplates: Encountered naming cycle within nested blocks:\n* %s",
				strings.Join(append(seen, partial.Extends), " -> "),
			)
		}
		agg, err := aggregateTemplates(partial.Blocks, append(seen, partial.Extends)...)
		if err != nil {
			return agg, err
		}
		if partial.Template != "" {
			these = append(these, partial.Template)
		}
		next = append(next, agg...)
	}
	return append(these, next...), nil
}

// contains checks if a list of values contains an element matching a query string exactly
func contains(values []string, query string) bool {
	for _, value := range values {
		if value == query {
			return true
		}
	}
	return false
}

// insertPartial creates a copy of the parent with the child partial incorporated into the template hierarchy,
// if possible.
//
// If the child partial does not match any blocks in the hierarchy,
// a nil pointer will be returned.
func insertPartial(parent, child *Partial, seen ...string) *Partial {
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
