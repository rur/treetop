package treetop

import (
	"bytes"
	"net/http"
	"strings"
)

func NewFragment(template string, handlerFunc HandlerFunc) Handler {
	rootBlock := blockInternal{name: "fragment-root"}
	handler := fragmentInternal{
		template:    template,
		handlerFunc: handlerFunc,
		extends:     &rootBlock,
		includes:    make(map[Block]Handler),
	}
	rootBlock.defaultHandler = &handler
	return &handler
}

type fragmentInternal struct {
	template    string
	handlerFunc HandlerFunc
	// private:
	blocks   map[string]Block
	includes map[Block]Handler
	extends  Block
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

func (h *fragmentInternal) GetBlocks() map[string]Block {
	return h.blocks
}
func (h *fragmentInternal) Includes(includes ...Handler) Handler {
	newHandler := fragmentInternal{
		template:    h.template,
		handlerFunc: h.handlerFunc,
		extends:     h.extends,
		includes:    make(map[Block]Handler),
		blocks:      make(map[string]Block),
	}
	for block, handler := range h.includes {
		newHandler.includes[block] = handler
	}
	for name, block := range h.blocks {
		newHandler.blocks[name] = block
	}
	for _, handler := range includes {
		newHandler.includes[handler.Extends()] = handler
	}
	return &newHandler
}
func (h *fragmentInternal) GetIncludes() map[Block]Handler {
	return h.includes
}

// Allow the use of treetop Hander as a HTTP handler
func (h *fragmentInternal) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var isFragment bool
	for _, accept := range strings.Split(r.Header.Get("Accept"), ",") {
		if strings.Trim(accept, " ") == FragmentContentType {
			isFragment = true
			break
		}
	}

	root := h.extends
	if !isFragment {
		http.Error(w, "Not Acceptable", http.StatusNotAcceptable)
		return
	}

	var render bytes.Buffer
	blockMap, templates := resolveTemplatesForHandler(root, h)
	if proceed := executeTemplate(templates, root, blockMap, w, r, &render); !proceed {
		return
	}

	// this will execute any includes that have not already been resolved
	for block, handler := range h.GetIncludes() {
		if _, found := blockMap[block]; !found {
			partialBlockMap, partialTempl := resolveTemplatesForHandler(block, handler)
			if proceed := executeTemplate(partialTempl, block, partialBlockMap, w, r, &render); !proceed {
				return
			}
		}
	}
	// content type should indicate a treetop partial
	w.Header().Set("Content-Type", FragmentContentType)
	w.Header().Set("X-Response-Url", r.URL.RequestURI())

	// write response body from byte buffer
	render.WriteTo(w)
}
