package treetop

import "net/http"

type Handler struct {
	// Pointer within a hierarchy of templates connected via 'blocks'
	Template *Template
	// Handlers that will be appended to response *only* for a partial request
	Postscript []*Template
	// Only accept request for treetop fragment content type (end-point will not serve a full page)
	FragmentOnly bool
	// Function that will be responsible for executing template contents against
	// data yielded from handlers
	Renderer RenderFunc
}

// implement http.Handler interface, see https://golang.org/pkg/net/http/?#Handler
func (h *Handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	// ,,, implementation here
	// Job is to create a manage DataWriter instance
}
