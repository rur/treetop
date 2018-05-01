package treetop

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

func NewFragment(template string, handlerFunc HandlerFunc) Fragment {
	handler := fragmentInternal{
		template:    template,
		handlerFunc: handlerFunc,
		execute:     DefaultTemplateExec,
	}
	return &handler
}

type fragmentInternal struct {
	template    string
	handlerFunc HandlerFunc
	execute     TemplateExec
	extends     Block
}

func (h *fragmentInternal) String() string {
	var details []string

	if h.template != "" {
		details = append(details, fmt.Sprintf("Template: '%s'", h.template))
	}

	return fmt.Sprintf("<Fragment %s>", strings.Join(details, " "))
}

func (h *fragmentInternal) Func() HandlerFunc {
	return h.handlerFunc
}
func (h *fragmentInternal) Template() string {
	return h.template
}
func (h *fragmentInternal) Extends() Block {
	return h.extends
}

// Allow the use of treetop Hander as a HTTP handler
func (h *fragmentInternal) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var isFragment bool
	var status int
	for _, accept := range strings.Split(r.Header.Get("Accept"), ",") {
		if strings.Trim(accept, " ") == FragmentContentType {
			isFragment = true
			break
		}
	}
	if !isFragment {
		http.Error(w, "Not Acceptable", http.StatusNotAcceptable)
		return
	}

	render := bytes.NewBuffer(make([]byte, 0))

	if resp, proceed := ExecuteFragment(h, map[string]interface{}{}, w, r); proceed {
		if resp.Status > status {
			status = resp.Status
		}
		// data was loaded successfully, now execute the templates
		if err := h.execute(render, []string{h.template}, resp.Data); err != nil {
			http.Error(w, fmt.Sprintf("Error executing templates: %s", err.Error()), http.StatusInternalServerError)
			return
		}
	} else {
		// handler has indicated that the request has already been satisfied, do not proceed any further
		return
	}

	// content type should indicate a treetop partial
	w.Header().Set("Content-Type", FragmentContentType)
	w.Header().Set("X-Response-Url", r.URL.RequestURI())

	// Since we are modulating the representation based upon a header value, it is
	// necessary to inform the caches. See https://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html#sec13.6
	w.Header().Set("Vary", "Accept")

	if status > 0 {
		w.WriteHeader(status)
	}
	// write response body from byte buffer
	render.WriteTo(w)
}
