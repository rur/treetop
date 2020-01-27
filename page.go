package treetop

// Page is an API for defining a hierarchy of top level and nested views
// each view has an associated handler and template string.
type Page struct {
	Execute TemplateExec
}

// NewPage creates a new default
func NewPage(execute TemplateExec) *Page {
	return &Page{
		Execute: execute,
	}
}

// NewView create a top level view definition with configuration
// derived from the page instance.
func (r *Page) NewView(template string, handlerFunc HandlerFunc) View {
	view := &viewImpl{
		template: template,
		handler:  handlerFunc,
		renderer: r.Execute,
	}
	return view
}
