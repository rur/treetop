package treetop

import (
	"bytes"
	"math"
	"net/http"
	"sync/atomic"
)

var count uint32

// Generate a numeric token which can be used to identify treetop
// responses *locally*. The only uniqueness requirement
// is that simultaneous requests must not possess the same value.
// (MaxUnit32 - 1) is a more than adaquite threshold for this purpose.
func nextLocalToken() uint32 {
	v := atomic.AddUint32(&count, 1)
	if v == math.MaxUint32 {
		// cycle once we reach max value
		atomic.StoreUint32(&count, 0)
		v = 0
	}
	return v
}

type Handler struct {
	// Pointer within a hierarchy of templates connected via 'blocks'
	Template *Template
	// Handlers that will be appended to response *only* for a partial request
	Postscript []*Template
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
		template:   h.Template,
	}

	h.Template.HandlerFunc(dw, req)

	// TODO: use buffer pool in the future
	var buf bytes.Buffer
	h.Renderer(&buf, []string{}, dw.data)
	buf.WriteTo(resp)
}
