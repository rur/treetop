package treetop

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

func Append(partial Partial, fragments ...Fragment) http.Handler {
	fs := make([]Fragment, len(fragments))
	for i, f := range fragments {
		fs[i] = f
	}
	return &appended{
		partial,
		fs,
		DefaultTemplateExec,
	}
}

type appended struct {
	partial   Partial
	fragments []Fragment
	execute   TemplateExec
}

func (a *appended) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var isPartial bool
	var status int
	for _, accept := range strings.Split(r.Header.Get("Accept"), ",") {
		if strings.Trim(accept, " ") == PartialContentType {
			isPartial = true
			break
		}
	}
	h := a.partial

	root := h.Extends()
	if !isPartial {
		// full page load, execute from the base handler up
		for root.Container() != nil {
			root = root.Container().Extends()
		}
	}

	var render bytes.Buffer
	blockMap := resolveBlockMap(root, h)
	rootHandler, ok := blockMap[root]
	if !ok {
		http.Error(w, fmt.Sprintf("Error resolving handler for block %s", root), http.StatusInternalServerError)
		return
	}
	templates := resolvePartialTemplates(rootHandler, blockMap)
	if resp, proceed := executePartial(rootHandler, blockMap, w, r); proceed {
		// data was loaded successfully, now execute the templates
		if resp.Status > status && status < 600 {
			status = resp.Status
		}
		if err := a.execute(&render, templates, resp.Data); err != nil {
			http.Error(w, fmt.Sprintf("Error executing templates: %s", err.Error()), http.StatusInternalServerError)
			return
		}
	} else {
		// handler has indicated that the request has already been satisfied, do not proceed any further
		return
	}

	if isPartial {
		// this will execute any includes that have not already been resolved
		for block, handler := range h.GetIncludes() {
			if _, found := blockMap[block]; !found {
				partBlockMap := resolveBlockMap(block, handler)
				partHandler, ok := partBlockMap[block]
				if !ok {
					http.Error(w, fmt.Sprintf("Error resolving handler for block %s", block), http.StatusInternalServerError)
					return
				}
				partTemplates := resolvePartialTemplates(partHandler, partBlockMap)
				if resp, proceed := executePartial(partHandler, partBlockMap, w, r); proceed {
					if resp.Status > status && status < 600 {
						status = resp.Status
					}
					// data was loaded successfully, now execute the templates
					if err := a.execute(&render, partTemplates, resp.Data); err != nil {
						http.Error(w, fmt.Sprintf("Error executing templates: %s", err.Error()), http.StatusInternalServerError)
						return
					}
				} else {
					// handler has indicated that the request has already been satisfied, do not proceed any further
					return
				}
			}
		}

		for _, fragment := range a.fragments {

			if resp, proceed := executeFragment(fragment, map[string]interface{}{}, w, r); proceed {
				if resp.Status > status && status < 600 {
					status = resp.Status
				}
				// data was loaded successfully, now execute the templates
				if err := a.execute(&render, []string{fragment.Template()}, resp.Data); err != nil {
					http.Error(w, fmt.Sprintf("Error executing templates: %s", err.Error()), http.StatusInternalServerError)
					return
				}
			} else {
				// handler has indicated that the request has already been satisfied, do not proceed any further
				return
			}
		}
		// content type should indicate a treetop partial
		w.Header().Set("Content-Type", PartialContentType)
		w.Header().Set("X-Response-Url", r.URL.RequestURI())
	}

	// Since we are modulating the representation based upon a header value, it is
	// necessary to inform the caches. See https://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html#sec13.6
	w.Header().Set("Vary", "Accept")

	if status > 0 {
		w.WriteHeader(status)
	}
	// write response body from byte buffer
	render.WriteTo(w)
}
