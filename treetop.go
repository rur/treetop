package treetop

import (
	"sort"
	"net/http"
)

const (
	ContentType = "application/x.treetop-html-partial+xml"
)

type DataWriter interface {
	http.ResponseWriter
	// render the template with the supplied data
	Data(interface{})
	// load data from specified child handler
	Delegate(string, *http.Request) (interface{}, bool)
}

type Block interface {
	WithDefault(string, HandlerFunc) Block
	Default() Handler
	Extend(string, HandlerFunc) Handler
	Container() Handler
}

type Handler interface {
	http.Handler
	Func() HandlerFunc
	Template() string
	Extends() Block
	DefineBlock(string) Block
	GetBlocks() map[string]Block
	Includes(...Handler) Handler
	GetIncludes() map[Block]Handler
}

type HandlerFunc func(DataWriter, *http.Request)

// assemble an index of how each block in the hierarchy is mapped to a handler based upon a 'primary' handler
func resolveTemplatesForHandler(block Block, primary Handler) (map[Block]Handler, []string) {
	var handler Handler
	if primary != nil {
		if block == primary.Extends() {
			handler = primary
		} else if include, found := primary.GetIncludes()[block]; found {
			handler = include
		} else if blockDefault := block.Default(); blockDefault != nil {
			handler = blockDefault
		}
	}

	var templates []string
	var blockMap map[Block]Handler

	if handler != nil {
		blockMap = map[Block]Handler{
			block: handler,
		}
		templates = []string{
			handler.Template(),
		}
		var subBlockMap map[Block]Handler
		var subTemplates []string
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
			subBlockMap, subTemplates = resolveTemplatesForHandler(childBlock, primary)
			for childBlock, childHandler := range subBlockMap {
				blockMap[childBlock] = childHandler
			}
			templates = append(templates, subTemplates...)
		}
	}
	filtered := templates[:0]
	for _, templ := range templates {
		if templ != "" {
			filtered = append(filtered, templ)
		}
	}

	return blockMap, filtered
}
