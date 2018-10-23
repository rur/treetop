package treetop

type Renderer struct {
	Execute TemplateExec
}

func NewRenderer(execute TemplateExec) *Renderer {
	return &Renderer{
		Execute: execute,
	}
}

func (r *Renderer) Define(template string, handlerFunc HandlerFunc) PartialDef {
	def := partialDefImpl{
		template: template,
		handler:  handlerFunc,
		renderer: r.Execute,
	}
	return &def
}

// module level define uses default template exec
func Define(template string, handlerFunc HandlerFunc) PartialDef {
	def := partialDefImpl{
		template: template,
		handler:  handlerFunc,
		renderer: DefaultTemplateExec,
	}
	return &def
}
