package treetop

// Treetop includes this view definition utility which is designed
// for constructing hierarchies of template files paired with handler functions.
//
// Creating an endpoint for every dynamic part of a page can result in
// a lot more handlers than are typically needed for a serverside web app.
//
// The template inheritance feature supported by the Go standard library is an
// ideal way to reuse HTML template fragments using hierarchies. This page view utility
// pairs each template file with a corresponding handler func, so that loading view data can
// be similarly modularized.
//
// Example:
//
// 		page := treetop.NewPage(treetop.DefaultTemplateExec)
// 		base := page.NewView("base.html", baseHandler)   // top level request handler
// 		content := base.NewSubView("content", "content.html", contentHandler)   // em
//
// 		// Register a http.Handler that is capable of rendering a full document
// 		// or just the a content section
//		appMux.Handle("/", treetop.ViewHandler(content))
//

// Page is an API for defining a hierarchy of top level and nested views
// each view has an associated handler and template string.
type Page struct {
	Execute TemplateExec
}

// NewPage will instantiate a page with the necessary configuration to
// for defining hierarchies of views
func NewPage(execute TemplateExec) *Page {
	return &Page{
		Execute: execute,
	}
}

// NewView create a top level view definition with configuration
// derived from the page instance.
func (r *Page) NewView(template string, handlerFunc HandlerFunc) *View {
	return &View{
		Template:    template,
		HandlerFunc: handlerFunc,
		Renderer:    r.Execute,
	}
}

// View is a paring of a template string with a treetop.HandlerFunc
// each view can contain a tree of named subviews
type View struct {
	Template    string
	Extends     *Block
	HandlerFunc HandlerFunc
	Blocks      []*Block
	Renderer    TemplateExec
}

// Block represent a slot that other sub-views can inhabit
// within an enclosing treetop.View definition
type Block struct {
	Name           string
	Parent         *View
	DefaultPartial *View
}

// SubView defines a new view (sub-view) that references to its parent via a
// named block.
func (v *View) SubView(blockName, template string, handler HandlerFunc) *View {
	var block *Block
	for i := 0; i < len(v.Blocks); i++ {
		if v.Blocks[i].Name == blockName {
			block = v.Blocks[i]
		}
	}
	if block == nil {
		block = &Block{
			Parent: v,
			Name:   blockName,
		}
		v.Blocks = append(v.Blocks, block)
	}
	return &View{
		Extends:     block,
		Template:    template,
		HandlerFunc: handler,
		Renderer:    v.Renderer,
	}
}

// DefaultSubView defines a new view (sub-view) that references it's parent via a
// named block, equivalent to SubView method. The difference is that the parent will also
// have a return reference this the new view, and will use it for the specified block
// when no other 'overriding' view is involved.
func (v *View) DefaultSubView(blockName, template string, handler HandlerFunc) *View {
	sub := v.SubView(blockName, template, handler)
	sub.Extends.DefaultPartial = sub
	return sub
}

// Handler will create a instance of a http.Handler designed to implement the
// Treetop protocol through the page view & subview inheritance system.
//
// The hander will be derived from the the current state of the View definition,
// subsequence changes to the View definition will not impact the Handler.
func (v *View) Handler() *Handler {
	part := v.derivePartial(nil)
	page := part
	root := v
	for root.Extends != nil && root.Extends.Parent != nil {
		root = root.Extends.Parent
		page = root.derivePartial(page)
	}
	return &Handler{
		Fragment: part,
		Page:     page,
		Renderer: v.Renderer,
	}
}

// derivePartial is an internal function used while constructing the HTTP
// treetop request handler instance.
func (v *View) derivePartial(override *Partial) *Partial {
	var extends string
	if v.Extends != nil {
		extends = v.Extends.Name
	}

	p := Partial{
		Extends:     extends,
		Template:    v.Template,
		HandlerFunc: v.HandlerFunc,
	}

	var blP *Partial
	for i := 0; i < len(v.Blocks); i++ {
		b := v.Blocks[i]
		blP = nil
		if override != nil && override.Extends == b.Name {
			blP = override
		} else if b.DefaultPartial != nil {
			blP = b.DefaultPartial.derivePartial(override)
		} else {
			// fallback when there is no default
			blP = &Partial{Extends: b.Name, HandlerFunc: Noop}
		}
		p.Blocks = append(p.Blocks, Partial{
			Extends:     b.Name,
			Template:    blP.Template,
			HandlerFunc: blP.HandlerFunc,
			Blocks:      blP.Blocks,
		})
	}
	return &p
}
