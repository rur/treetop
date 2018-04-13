package treetop

import "net/http"

type renderer struct {
	execute TemplateExec
}

func NewRenderer(execute TemplateExec) Renderer {
	return &renderer{execute}
}

func (r *renderer) NewPage(template string, handlerFunc HandlerFunc) Partial {
	rootBlock := blockInternal{
		name:    "page root",
		execute: r.execute,
	}
	partial := partialInternal{
		template:    template,
		handlerFunc: handlerFunc,
		extends:     &rootBlock,
		includes:    make(map[Block]Partial),
		blocks:      make(map[string]Block),
		execute:     r.execute,
	}
	rootBlock.defaultPartial = &partial
	return &partial
}

func (r *renderer) NewFragment(template string, handlerFunc HandlerFunc) Fragment {
	fragment := fragmentInternal{
		template:    template,
		handlerFunc: handlerFunc,
		execute:     r.execute,
	}
	return &fragment
}

func (r *renderer) Append(partial Partial, fragments ...Fragment) http.Handler {
	fs := make([]Fragment, len(fragments))
	for i, f := range fragments {
		fs[i] = f
	}
	return &appended{
		partial,
		fs,
		r.execute,
	}
}
