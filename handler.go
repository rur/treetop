package treetop

import (
	"bytes"
	"net/http"
	"strings"
)

func NewHandler(template string, handlerFunc HandlerFunc) Handler {
	rootBlock := blockInternal{name: "root"}
	handler := handlerInternal{
		template:    template,
		handlerFunc: handlerFunc,
		extends:     &rootBlock,
		includes:    make(map[Block]Handler),
		blocks:      make(map[string]Block),
	}
	rootBlock.defaultHandler = &handler
	return &handler
}

type handlerInternal struct {
	template    string
	handlerFunc HandlerFunc
	// private:
	blocks   map[string]Block
	includes map[Block]Handler
	extends  Block
}

func (h *handlerInternal) Func() HandlerFunc {
	return h.handlerFunc
}
func (h *handlerInternal) Template() string {
	return h.template
}
func (h *handlerInternal) Extends() Block {
	return h.extends
}
func (h *handlerInternal) DefineBlock(name string) Block {
	block := blockInternal{
		name:      name,
		container: h,
	}
	h.blocks[name] = &block
	return &block
}
func (h *handlerInternal) GetBlocks() map[string]Block {
	return h.blocks
}
func (h *handlerInternal) Includes(includes ...Handler) Handler {
	newHandler := handlerInternal{
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
func (h *handlerInternal) GetIncludes() map[Block]Handler {
	return h.includes
}

// Allow the use of treetop Hander as a HTTP handler
func (h *handlerInternal) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
