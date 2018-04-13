package treetop

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

func NewPage(template string, handlerFunc HandlerFunc) Partial {
	rootBlock := blockInternal{
		name:    "page root",
		execute: DefaultTemplateExec,
	}
	partial := partialInternal{
		template:    template,
		handlerFunc: handlerFunc,
		extends:     &rootBlock,
		includes:    make(map[Block]Partial),
		blocks:      make(map[string]Block),
		execute:     DefaultTemplateExec,
	}
	rootBlock.defaultPartial = &partial
	rootBlock.execute = partial.execute
	return &partial
}

type partialInternal struct {
	template    string
	handlerFunc HandlerFunc
	// private:
	blocks   map[string]Block
	includes map[Block]Partial
	extends  Block
	execute  TemplateExec
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
		execute:   h.execute,
	}
	h.blocks[name] = &block
	return &block
}
func (h *partialInternal) GetBlocks() map[string]Block {
	return h.blocks
}
func (h *partialInternal) Includes(includes ...Partial) Partial {
	newHandler := partialInternal{
		template:    h.template,
		handlerFunc: h.handlerFunc,
		extends:     h.extends,
		execute:     h.execute,
		includes:    make(map[Block]Partial),
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

func (h *partialInternal) GetIncludes() map[Block]Partial {
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
	blockMap := resolveBlockMap(root, h)
	rootHandler, ok := blockMap[root]
	if !ok {
		http.Error(w, fmt.Sprintf("Error resolving handler for block %s", root), http.StatusInternalServerError)
		return
	}
	templates := resolvePartialTemplates(rootHandler, blockMap)
	if data, proceed := ExecutePartial(rootHandler, blockMap, w, r); proceed {
		// data was loaded successfully, now execute the templates
		if err := h.execute(&render, templates, data); err != nil {
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
				if data, proceed := ExecutePartial(partHandler, partBlockMap, w, r); proceed {
					// data was loaded successfully, now execute the templates
					if err := h.execute(&render, partTemplates, data); err != nil {
						http.Error(w, fmt.Sprintf("Error executing templates: %s", err.Error()), http.StatusInternalServerError)
						return
					}
				} else {
					// handler has indicated that the request has already been satisfied, do not proceed any further
					return
				}
			}
		}
		// content type should indicate a treetop partial
		w.Header().Set("Content-Type", PartialContentType)
		w.Header().Set("X-Response-Url", r.URL.RequestURI())
	}

	// Since we are modulating the representation based upon a header value, it is
	// necessary to inform the caches. See https://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html#sec13.6
	w.Header().Set("Vary", "Accept")

	// write response body from byte buffer
	render.WriteTo(w)
}

// assemble an index of how each block in the hierarchy is mapped to a handler
// based upon a 'primary' handler which acts as the entry point to the hierarchy.
func resolveBlockMap(block Block, primary Partial) map[Block]Partial {
	handler := primary
	for handler != nil {
		if block == handler.Extends() {
			break
		} else if include, found := handler.GetIncludes()[block]; found {
			handler = include
			break
		}
		handler = handler.Extends().Container()
	}

	if handler == nil {
		if blockDefault := block.Default(); blockDefault != nil {
			handler = blockDefault
		}
	}
	var blockMap map[Block]Partial

	if handler != nil {
		blockMap = map[Block]Partial{
			block: handler,
		}
		var subBlockMap map[Block]Partial
		var names []string
		var childBlock Block
		blocks := handler.GetBlocks()
		// sort block names because we want the order of templates to be stable even if it
		// isn't a total order in general
		for k := range blocks {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, name := range names {
			childBlock = blocks[name]
			subBlockMap = resolveBlockMap(childBlock, primary)
			for childBlock, childHandler := range subBlockMap {
				blockMap[childBlock] = childHandler
			}
		}
	}
	return blockMap
}

func resolvePartialTemplates(partial Partial, handlerMap map[Block]Partial) []string {
	templates := []string{partial.Template()}
	blocks := partial.GetBlocks()
	keys := make([]string, len(blocks))

	i := 0
	for k := range blocks {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	for _, name := range keys {
		block := blocks[name]
		if childPartial, ok := handlerMap[block]; ok {
			for _, t := range resolvePartialTemplates(childPartial, handlerMap) {
				templates = append(templates, t)
			}
		}
	}
	return templates
}
