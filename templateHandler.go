package treetop

import (
	"html/template"
	"io"
	"net/http"
)

// TemplateHandler implements the treetop.ViewHandler interface for endpoints that support the treetop protocol
type TemplateHandler struct {
	Page             *View
	PageTemplate     *template.Template
	Partial          *View
	PartialTemplate  *template.Template
	Includes         []*View
	IncludeTemplates []*template.Template
}

// FragmentOnly creates a new Handler that only responds to fragment requests
func (h TemplateHandler) FragmentOnly() ViewHandler {
	return TemplateHandler{
		Partial:          h.Partial,
		Includes:         h.Includes,
		PartialTemplate:  h.PartialTemplate,
		IncludeTemplates: h.IncludeTemplates,
	}
}

// PageOnly create a new handler that will only respond to non-fragment (full page) requests
func (h TemplateHandler) PageOnly() ViewHandler {
	return TemplateHandler{
		Page:         h.Page,
		PageTemplate: h.PageTemplate,
	}
}

//
// TODO: Implementation needed
func (h TemplateHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// NOTE: this is pseudocode
	resp := BeginResponse(req.Context(), w)
	defer resp.Cancel()

	// use buffer pool and treetop Writer
	// first figure out if this is a full page, partial or fragment request
	// if it is a full page only execute handlers and template for the

	// NOTE: this is pseudocode
	if !IsTreetopRequest(req) {
		// this is a full page request, just render the page
		if h.Page == nil {
			http.Error(w, http.StatusText(http.StatusNotAcceptable), http.StatusNotAcceptable)
			return
		}
		pageResp := resp.WithView(h.Page)
		data := h.Page.Handler(pageResp, req)
		if pageResp.Finished() {
			return
		}
		err := h.PageTemplate.Execute(resp, data)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// NOTE: this is pseudocode
	var (
		views = append([]*View{h.Partial}, h.Includes...)
		data  = make([]interface{}, len(views))
		tmpls = append([]*template.Template{h.PartialTemplate}, h.IncludeTemplates...)
	)

	for i, view := range views {
		data[i] = view.Handler(resp.WithView(view), req)
		if resp.Finished() {
			return
		}
	}

	// NOTE: this is pseudocode
	var (
		ttW io.Writer
		ok  bool
	)
	if h.Page == nil {
		ttW, ok = resp.NewFragmentWriter(req)
	} else {
		ttW, ok = resp.NewPartialWriter(req)
	}
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotAcceptable), http.StatusNotAcceptable)
		return
	}

	// NOTE: this is pseudocode
	for i, tmpl := range tmpls {
		tmpl.Execute(ttW, data[i])
	}
}
