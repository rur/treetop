package treetop

const (
	PartialContentType  = "application/x.treetop-html-partial+xml"
	FragmentContentType = "application/x.treetop-html-fragment+xml"
)

type Treetop struct {
	Execute TemplateExec
}

func NewTreetop(execute TemplateExec) *Treetop {
	return &Treetop{
		Execute: execute,
	}
}

func (r *Treetop) Define(template string, handlerFunc HandlerFunc) PartialDef {
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
