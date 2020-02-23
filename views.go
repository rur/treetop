package treetop

import "net/http"

// View is a utility for building hierarchies of nested templates
// from which HTTP request handlers can be constructed.
//
// Multi-page web apps require a lot of endpoints. Template inheritance
// is commonly used to reduce HTML boilerplate and improve reuse. Treetop views incorporate
// request handlers into the hierarchy to gain the same advantage.
//
// A 'View' is a template string (usually file path) paired with a handler function.
// Go templates can contain named nested blocks. Defining a 'SubView' associates
// a handler and a template with a block embedded within a parent template.
// HTTP handlers can then be constructed for various page configurations.
//
// Example of a basic template hierarchy
//
//                  baseHandler(...)
//                | base.html ========================|
//                | …                                 |
//                | {{ template "content" .Content }} |
//                | …               ^                 |
//                |_________________|_________________|
//                                  |
//                           ______/ \______
//      contentAHandler(...)               contentBHandler(...)
//    | contentA.html ========== |        | contentB.html ========== |
//    |                          |        |                          |
//    | {{ block "content" . }}… |        | {{ block "content" . }}… |
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
// Example of using the library to bind constructed handlers to a HTTP router.
//
// 		base := treetop.NewView(
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
//		exec := treetop.DefaultExecutor{}
//		mymux.Handle("/path/to/a", exec.ViewHandler(contentA))
//		mymux.Handle("/path/to/b", exec.ViewHandler(contentB))
//
//
// This is useful for creating Treetop enabled endpoints because the constructed handler
// is capable of loading either a full page or just the "content" part of the page depending
// upon the request.
//
type View struct {
	Template    string
	HandlerFunc ViewHandlerFunc
	SubViews    map[string]*View
	Defines     string
	Parent      *View
}

// ViewHandlerFunc is the interface for treetop handler functions that support hierarchical
// partial data loading.
type ViewHandlerFunc func(Response, *http.Request) interface{}

// ViewHandler is an extention of the http.Handler interface with methods added
// for extra treetop endpoint configuration
type ViewHandler interface {
	http.Handler
	FragmentOnly() ViewHandler
	PageOnly() ViewHandler
}

// ViewExecutor is an interface for objects that implement transforming a View definition
// into a ViewHandler that supports pages, partial and fragment requests.
type ViewExecutor interface {
	NewViewHandler(view *View, includes ...*View) ViewHandler
}

// NewView creates an instance of a view given a template + handler pair
func NewView(tmpl string, handler ViewHandlerFunc) *View {
	return &View{
		Template:    tmpl,
		HandlerFunc: handler,
		SubViews:    make(map[string]*View),
	}
}

// NewSubView creates an instance of a view given a template + handler pair
// this view is a detached subview, in that is does not reference a parent
func NewSubView(defines, tmpl string, handler ViewHandlerFunc) *View {
	v := NewView(tmpl, handler)
	v.Defines = defines
	return v
}

// NewSubView create a new view extending a named block within the current view
func (v *View) NewSubView(defines string, tmpl string, handler ViewHandlerFunc) *View {
	sub := NewView(tmpl, handler)
	sub.Defines = defines
	sub.Parent = v
	if _, ok := v.SubViews[defines]; !ok {
		v.SubViews[defines] = nil
	}
	return sub
}

// NewDefaultSubView create a new view extending a named block within the current view
// and updates the parent to use this view by default
func (v *View) NewDefaultSubView(defines string, tmpl string, handler ViewHandlerFunc) *View {
	sub := v.NewSubView(defines, tmpl, handler)
	v.SubViews[defines] = sub
	return sub
}

// Copy creates a duplicate so that the original is not affected by
// changes
func (v *View) Copy() *View {
	copy := NewView(v.Template, v.HandlerFunc)
	copy.Defines = v.Defines
	copy.Parent = v.Parent
	for name, sub := range v.SubViews {
		copy.SubViews[name] = sub
	}
	return copy
}

// CompileViews is used to create an endpoint configuration combining supplied view
// definitions based upon the template names they define.
//
// This returns:
//   - a full-page view instance,
//   - a partial page view instance, and
//   - any disconnect fragment views that should be appended to partial requests.
//
// TODO: Implementation needed
func CompileViews(view *View, includes ...*View) (page, part *View, postscript []*View) {
	// Merge the includes and the view where possible.
	// Views to the left 'consume' those to the right when a match is found.
	// 'Postscripts' are includes that could not be merged.
	// TODO: implementation here
	part = view.Copy()

	for _, incl := range includes {
		postscript = append(postscript, incl.Copy())
	}

	// NOTE: this is pseudocode
	// constructing the 'page' involes modifying the series of parents
	// to ensure that this view is reachable from the root. The modified root is our page
	root := view
	for root.Parent != nil {
		// make a copy of the parent and ensure that it points to
		// the child sub view
		pCopy := root.Parent.Copy()
		pCopy.SubViews[root.Defines] = root
		root = root.Parent
	}
	if root != view {
		page, _ = insertView(root, view)
	} else {
		page = view
	}
	for _, incl := range includes {
		page, _ = insertView(page, incl)
	}

	return page, part, postscript
}

// insertView attempts to incorporate the child into the template hierarchy of this view.
// If a match is found for the definition name, views will be copied & modified as necessary and
// a flag is returned to indicate whether a match was found.
//
// TODO: Implementation needed
func insertView(view, child *View) (*View, bool) {
	// NOTE: this is pseudocode
	if _, found := view.SubViews[child.Defines]; found {
		copy := view.Copy()
		copy.SubViews[child.Defines] = child
		return copy, true
	}

	// At this point a match was not found directly in this view,
	// Attempt to apply the child to reachable subviews
	//
	// NOTE: this is pseudocode
	for _, sub := range view.SubViews {
		copiedSub, found := insertView(sub, view)
		if found {
			copy := view.Copy()
			copy.SubViews[copiedSub.Defines] = copiedSub
			return copy, true
		}
	}
	return view, false
}
