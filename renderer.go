package treetop

type renderer struct {
	execute TemplateExec
}

func NewRenderer(execute TemplateExec) Renderer {
	return &renderer{execute}
}

func (r *renderer) NewPage(template string, handlerFunc HandlerFunc) Partial {
	rootBlock := blockInternal{name: "page root"}
	partial := partialInternal{
		template:    template,
		handlerFunc: handlerFunc,
		extends:     &rootBlock,
		includes:    make(map[Block]Partial),
		blocks:      make(map[string]Block),
		execute:     r.execute,
	}
	rootBlock.defaultPartial = &partial
	rootBlock.execute = partial.execute
	return &partial
}

func (r *renderer) NewFragment(template string, handlerFunc HandlerFunc) Handler {
	fragment := fragmentInternal{
		template:    template,
		handlerFunc: handlerFunc,
		execute:     r.execute,
	}
	return &fragment
}
