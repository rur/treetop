package treetop

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

func NewPage(template string, handlerFunc HandlerFunc) Partial {
	rootBlock := blockInternal{name: "page root"}
	handler := partialInternal{
		template:    template,
		handlerFunc: handlerFunc,
		extends:     &rootBlock,
		includes:    make(map[Block]Handler),
		blocks:      make(map[string]Block),
	}
	rootBlock.defaultHandler = &handler
	return &handler
}

type partialInternal struct {
	template    string
	handlerFunc HandlerFunc
	// private:
	blocks   map[string]Block
	includes map[Block]Handler
	extends  Block
}

func (h *partialInternal) String() string {
	var details []string

	if h.template != "" {
		details = append(details, fmt.Sprintf("Template: '%s'", h.template))
	}

	if h.extends != nil {
		details = append(details, fmt.Sprintf("Extends: %s", h.extends))
	}

	inclTempl := make([]string, 0, len(h.includes))
	for _, incl := range h.includes {
		inclTempl = append(inclTempl, incl.Template())
	}
	if len(inclTempl) > 0 {
		details = append(details, fmt.Sprintf("Includes: x%d", len(inclTempl)))
	}

	return fmt.Sprintf("<Partial %s>", strings.Join(details, " "))
}

func (h *partialInternal) Func() HandlerFunc {
	return h.handlerFunc
}
func (h *partialInternal) Template() string {
	return h.template
}
func (h *partialInternal) Extends() Block {
	return h.extends
}
func (h *partialInternal) DefineBlock(name string) Block {
	block := blockInternal{
		name:      name,
		container: h,
	}
	h.blocks[name] = &block
	return &block
}
func (h *partialInternal) GetBlocks() map[string]Block {
	return h.blocks
}
func (h *partialInternal) Includes(includes ...Handler) Handler {
	newHandler := partialInternal{
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
func (h *partialInternal) GetIncludes() map[Block]Handler {
	return h.includes
}

// Allow the use of treetop Hander as a HTTP handler
func (h *partialInternal) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var isPartial bool
	for _, accept := range strings.Split(r.Header.Get("Accept"), ",") {
		if strings.Trim(accept, " ") == PartialContentType {
			isPartial = true
			break
		}
	}

	root := h.extends
	if !isPartial {
		// full page load, execute from the base handler up
		for root.Container() != nil {
			root = root.Container().Extends()
		}
	}

	var render bytes.Buffer
	blockMap, templates := resolveTemplatesForHandler(root, h)
	if proceed := executeTemplate(templates, root, blockMap, w, r, &render); !proceed {
		return
	}

	if isPartial {
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
		w.Header().Set("Content-Type", PartialContentType)
		w.Header().Set("X-Response-Url", r.URL.RequestURI())
	}

	// write response body from byte buffer
	render.WriteTo(w)
}
