package treetop

type renderer struct {
	execute TemplateExec
}

func NewRenderer(execute TemplateExec) Renderer {
	return &renderer{execute}
}

func (r *renderer) Define(template string, handlerFunc HandlerFunc) PartialDef {
	def := partialDefImpl{
		template: template,
		handler:  handlerFunc,
		renderer: r.execute,
	}
	return &def
}

// define uses default template exec
func Define(template string, handlerFunc HandlerFunc) PartialDef {
	def := partialDefImpl{
		template: template,
		handler:  handlerFunc,
		renderer: DefaultTemplateExec,
	}
	return &def
}
