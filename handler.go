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
	Content     string
	HandlerFunc HandlerFunc
	Parent      *Partial
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

	// cached templates
	partialTemplates []string
	pageTemplates    []string
}

// implement http.Handler interface, see https://golang.org/pkg/net/http/?#Handler
func (h *Handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	dw := &dataWriter{
		writer:     resp,
		localToken: nextLocalToken(),
		partial:    h.Partial,
	}

	h.Partial.HandlerFunc(dw, req)

	tmpls, err := aggregateTemplates([]string{h.Partial.Extends}, h.Partial.Blocks)
	if err != nil {
		log.Printf(err.Error())
		http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	tmpls = append([]string{h.Partial.Content}, tmpls...)

	// TODO: use buffer pool
	var buf bytes.Buffer
	h.Renderer(&buf, tmpls, dw.data)
	buf.WriteTo(resp)
}

func aggregateTemplates(seen []string, tmpls []*Partial) ([]string, error) {
	var these []string
	var next []string
	for i := 0; i < len(tmpls); i++ {
		if contains(seen, tmpls[i].Extends) {
			return nil, fmt.Errorf(
				"aggregateTemplates: Encountered naming cycle within nested blocks:\n* %s",
				strings.Join(append(seen, tmpls[i].Extends), " -> "),
			)
		} else {
			seen = append(seen, tmpls[i].Extends)
		}
		agg, err := aggregateTemplates(seen, tmpls[i].Blocks)
		if err != nil {
			return agg, err
		}
		these = append(these, tmpls[i].Content)
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
