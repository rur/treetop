package treetop

// View is a utility for generating request handlers given a definition
// of a template hierarchies.
//
// A view is a pair: a template string (usually file path) and a handler function.
// A view can contain zero or more 'blocks', these are named sub-sections which allow
// other views to be swapped in as needed.
//
// Since multi-page web applications require an endpoint for every action.
// The template inheritance feature supported by the Go standard library is useful
// for reusing HTML template fragments.
//
// Example of a basic template hierarchy
//
//                      baseHandler(...)
//                    / base.html ==========\
//                    |                     |
//                    |  / "content" ===\   |
//                    |  |              |   |
//                    |  |      ^       |   |
//                    |  \______^_______/   |
//                    \_________^___________/
//                              ^
//                             / \
//     contentAHandler(...) __/   \__ contentBHandler(...)
//     contentA.html                  contentB.html
//
//
// Request/response example
//
//     GET /path/to/a
//     > HTTP/1.1 200 OK
//     > ... { base.html + contentA.html }
//
//     GET /path/to/b
//     > HTTP/1.1 200 OK
//     > ... { base.html + contentB.html }
//
//
// Example of using Treetop View type to generate handers for these endpoints
//
// 		base := page.NewView(
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
// Notice that each template is paired with it's own handler function. As a result
// request handling can be modularized along with the markup. In the example above,
// the HTTP handlers created by the Treetop library are capable of rendering either
// a full HTML document or the relevant "content" section alone.
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
