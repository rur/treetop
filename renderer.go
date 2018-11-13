package treetop

type Renderer struct {
	Execute TemplateExec
}

func NewRenderer(execute TemplateExec) *Renderer {
	return &Renderer{
		Execute: execute,
	}
}

func (r *Renderer) NewView(template string, handlerFunc HandlerFunc) View {
	view := &viewImpl{
		template: template,
		handler:  handlerFunc,
		renderer: r.Execute,
	}
	return view
}

// module level define uses default template exec
func NewView(template string, handlerFunc HandlerFunc) View {
	view := &viewImpl{
		template: template,
		handler:  handlerFunc,
		renderer: DefaultTemplateExec,
	}
	return view
}
