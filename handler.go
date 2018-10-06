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

	tmpls, err := aggregateTemplateContents([]string{h.Template.Content}, h.Template.Blocks)
	if err != nil {
		log.Printf(err.Error())
		http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	tmpls = append([]string{h.Template.Content}, tmpls...)

	// TODO: use buffer pool
	var buf bytes.Buffer
	h.Renderer(&buf, tmpls, dw.data)
	buf.WriteTo(resp)
}

func aggregateTemplateContents(seen []string, tmpls []*Template) ([]string, error) {
	if len(seen) > 1000 {
		return seen, fmt.Errorf(
			"aggregateTemplate: Max iterations reached, it is likely that there is a cycle in template definitions. Last 20 templates:\n- %s",
			strings.Join(seen[len(seen)-22:], "\n- "),
		)
	}
	var these []string
	var next []string
	for i := 0; i < len(tmpls); i++ {
		seen = append(seen, tmpls[i].Content)
		agg, err := aggregateTemplateContents(seen, tmpls[i].Blocks)
		if err != nil {
			return agg, err
		}
		these = append(these, tmpls[i].Content)
		next = append(next, agg...)
	}
	return append(these, next...), nil
}
