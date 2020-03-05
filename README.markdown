[![Build Status](https://travis-ci.org/rur/treetop.svg?branch=v0.3.0)](https://travis-ci.org/rur/treetop)

# Treetop

## Hierarchical Request Handlers

Construct complex HTML endpoints from composable templates and functions.

    base := treetop.NewView("base.html.tmpl", BaseHandler)
    nav := base.NewSubView("nav", "nav.html.tmpl", NavHandler)
    _ = base.NewDefaultSubView("sidebar", "sidebar.html.tmpl", SidebarHandler)
    contentA := base.NewSubView("content", "content_a.html.tmpl", ContentAHandler)
    contentB := base.NewSubView("content", "content_b.html.tmpl", ContentBHandler)

    exec := treetop.FileExecutor{}
    mux.Handle("/content_a", exec.NewViewHandler(contentA, nav))
    mux.Handle("/content_b", exec.NewViewHandler(contentB, nav))

Nested definitions have first-class support in the Go template library<sup>1</sup>. Treetop takes this a step further adding nested handlers so that page variations can be assembled for different endpoints.

###  HTML Template Protocol

Self-contained views have a special benefit, sections of a page can be rendered in isolation. Thus, Treetop handlers are capable of generating 'partial' fragments enclosed in a HTML template tag.

    > GET / HTTP/1.1
      Accept: application/x.treetop-html-template+xml

    < HTTP/1.1 200 OK
      Content-Type: application/x.treetop-html-template+xml
      Vary: Accept

      <template>
          <div id="content">...</div>
          <div id="nav">...</div>
      </template>

A [Treetop Client Library](https://github.com/rur/treetop-client) is used to issue these XHR requests and apply the template to the DOM.


## Example

A very basic example can be run from this repo (still needs some work!)

    $ go run example/

Tip. Activate your network tab to observe what's going on.

### Other Examples

- [Todo \*Without\* MVC](https://github.com/rur/todowithoutmvc) - Treetop implementation of [TodoMVC](http://todomvc.com) app.

## View Handlers

View handlers load data for their corresponding template. Just as nested templates are embedded in their parent, nested template data is embedded in the data of it's parent. For example,

    func ParentHandler(rsp treetop.Response, req *http.Request) interface{} {
        return struct {
            ...
            Child interface{}
        }{
            ...
            Child: rsp.HandleSubView("child", req),
        }
    }

Data is passed within the template like so,

    <div id="parent">
        ...
        {{ template "child" .Child }}
    </div>


## Response Writer

The Treetop Go library provides utilities for writing ad-hoc template responses when needed. PartialWriter and FragmentWriter wrap the `http.ResponseWriter`,

    func myHandler(w http.ResponseWriter, req *http.Request) {
        // check for treetop request and construct a writer
        if tw, ok := treetop.NewFragmentWriter(w, req); ok {
            fmt.Fprintf(tw, `<h3 id="greeting">Hello, %s</h3>`, "Treetop")
        }
    }


## Client Library

The client library will parse the HTML template and use those fragments to update the DOM. The `id` attribute of each top level element will be matched to an existing node in the document.

* Matched elements in the current DOM will be replaced.
* Unmatched contents of the template will be ignored.

See [Client Library](https://github.com/rur/treetop-client) for more information.


## References
1. Go supports template inheritance through [nested template definitions](https://tip.golang.org/pkg/text/template/#hdr-Nested_template_definitions).
