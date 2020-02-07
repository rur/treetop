package treetop

// View is a utility for defining hierarchies of nested templates
// from which HTTP request handlers can be generated.
//
// Multi-page web applications require a lot of endpoints. Template inheritance
// is commonly used to reduce HTML boilerplate and improve reuse. Treetop views incorporate
// request handlers into the hierarchy to gain the same benefit.
//
// A 'View' is a template string (usually file path) paired with a handler function.
// In Go, templates can contain named nested blocks. Defining a 'SubView' associates
// a request handler and a fragment template with a parent. Thus HTTP handlers can
// be easily generated for various page configurations. Within a generated handler,
// data is passed between parent and child in a mechanical way.
//
// Example of a basic template hierarchy
//
//                  baseHandler(...)
//                | base.html ========================|
//                |                                   |
//                | {{ template "content" .Content }} |
//                |_________________^_________________|
//                                  |
//                           ______/ \______
//      contentAHandler(...)               contentBHandler(...)
//    | contentA.html ========== |        | contentB.html ========== |
//    |                          |        |                          |
//    | {{ define "content" }}.. |        | {{ define "content" }}.. |
//    |__________________________|        |__________________________|
//
// Pseudo request and response:
//
//     GET /path/to/a
//     > HTTP/1.1 200 OK
//     > ... base.html { Content: contentA.html }
//
//     GET /path/to/b
//     > HTTP/1.1 200 OK
//     > ... base.html { Content: contentB.html }
//
//
// Example of using the library to bind generated handlers to a HTTP router.
//
// 		base := treetop.NewView(
// 			treetop.DefaultTemplateExec,
// 			"base.html",
// 			baseHandler,
// 		)
//
// 		contentA := base.NewSubView(
// 			"content",
// 			"contentA.html",
// 			contentAHandler,
// 		)
//
// 		contentB := base.NewSubView(
// 			"content",
// 			"contentB.html",
// 			contentBHandler,
// 		)
//
//		mymux.Handle("/path/to/a", treetop.ViewHandler(contentA))
//		mymux.Handle("/path/to/b", treetop.ViewHandler(contentB))
//
//
// This is useful for creating Treetop enabled endpoints because we wish to be able
// to either load a full page or just a part of a page depending upon the request.
// The generated 'ViewHandler' supports Treetop partials by default.
//
type View struct {
	Template    string
	Extends     *Block
	HandlerFunc HandlerFunc
	Blocks      []*Block
	Renderer    TemplateExec
}

// NewView create a top level view definition which is designed
// for constructing hierarchies of template files paired with handler functions.
func NewView(execute TemplateExec, template string, handlerFunc HandlerFunc) *View {
	return &View{
		Template:    template,
		HandlerFunc: handlerFunc,
		Renderer:    execute,
	}
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
// named block. It is equivalent to the SubView method except that the parent will also
// have a return reference. The new view will become the 'default' template
// for the specified block in the parent.
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
