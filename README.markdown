[![Build Status](https://travis-ci.org/rur/treetop.svg?branch=v0.3.0)](https://travis-ci.org/rur/treetop)

# Treetop

## Hierarchical Web Handlers

Treetop is a utility for defining hierarchies of nested templates
from which HTTP handlers can be constructed.

HTML applications typically share a lot of common structure between endpoints.
Composable templates are supported in Go<sup>1</sup> to reduce HTML boilerplate.
Treetop combines a function with each template to make nesting
easier to manage, reducing boilerplate in request handlers.

_Example of a basic hierarchy_


                  BaseHandler(...)
                | base.html ========================|
                | …                                 |
                | {{ template "content" .Content }} |
                | …               ^                 |
                |_________________|_________________|
                                  |
                           ______/ \______
      ContentAHandler(...)                ContentBHandler(...)
    | contentA.html ========== |        | contentB.html ========== |
    |                          |        |                          |
    | <div id="content">...</… |        | <div id="content">...</… |
    |__________________________|        |__________________________|

A 'View' is a template string (usually file path) paired with a handler function.
Defining a 'SubView' creates a new template + handler pair associated with
an embedded block. HTTP endpoints can then be constructed for various page configurations.

The code below extends this example to bind the routes `"/content_a"` and `"/content_b"` with composite handlers.

    base := treetop.NewView(
        "base.html", BaseHandler)
    nav := base.NewSubView(
        "nav", "nav.html", NavHandler)
    _ = base.NewDefaultSubView(
        "sidebar", "sidebar.html", SidebarHandler)
    contentA := base.NewSubView(
        "content", "content_a.html", ContentAHandler)
    contentB := base.NewSubView(
        "content", "content_b.html", ContentBHandler)

    exec := treetop.FileExecutor{}
    mux.Handle("/content_a", exec.NewViewHandler(contentA, nav))
    mux.Handle("/content_b", exec.NewViewHandler(contentB, nav))

The 'base', 'nav' and 'sidebar' views will be incorporated into both endpoints.

Example of named template blocks in `"base.html"`,

	...
	<div class="layout">
		{{ block "nav" .Nav }}  <div id="nav">default nav</div> {{ end }}

		{{ template "sidebar" .SideBar }}

		{{ template "content" .Content }}
	</div>
	...

Views can have as many levels of nesting as needed.

### No Third-Party Dependencies

The Treetop package wraps features of the Go standard library, mostly within "net/http" and "html/template".


##  HTML Template Protocol

### Hot-swap sections of a page without JS boilerplate

Since views are self-contained, they can be rendered in isolation. Treetop
handlers support rendering template fragments that can be 'applied' to a loaded document.
The following is an illustration of the protocol.

    > GET /content_a HTTP/1.1
      Accept: application/x.treetop-html-template+xml

    < HTTP/1.1 200 OK
      Content-Type: application/x.treetop-html-template+xml
      Vary: Accept

      <template>
          <div id="content">...</div>
          <div id="nav">...</div>
      </template>

There are many ways hot-swapping views can be used to improve user experience <sup>[docs needed]</sup>.

A [Treetop Client Library](https://github.com/rur/treetop-client) is available. It sends these requests
using XHR and applies template fragments to the DOM with a simple find and replace mechanism.

## Example

A very basic example can be run from this repo <sup>[needs improvement]</sup>.

    $ go run ./example/

Tip. Activate your network tab to observe what's going on.

### Other Examples:

- [Todo \*Without\* MVC](https://github.com/rur/todowithoutmvc) - Treetop implementation of [TodoMVC](http://todomvc.com) app using the template protocol.

## Template Executor

An 'Executor' is responsible for loading and configuring templates. It constructs a [HTTP Handler](https://golang.org/pkg/net/http/#Handler) instance to manage the plumbing between loading data and executing templates for a request. You can implement your own template loader <sup>[docs needed]</sup>, but the following are provided:
- `FileExecutor` - load template files using os.Open
- `FileSytemExecutor` - loads templates from a supplied http.FileSystem instance
- `StringExecutor` - treat the view template property as an inline template string
- `KeyedStringExecutor` - treat the view template property is key into a template map
- `DeveloperExecutor` - force per request template parsing

## View Handler Function

A view handler function loads data for a corresponding Go template. Just as nested templates are embedded in a parent, nested handler data is embedded in the _data_ of it's parent.

Example of a handler loading data for a child template,

    func ParentHandler(rsp treetop.Response, req *http.Request) interface{} {
        return struct {
            ...
            Child interface{}
        }{
            ...
            Child: rsp.HandleSubView("child", req),
        }
    }

Data is subsequently passed down within the template like so,

    <div id="parent">
        ...
        {{ template "child" .Child }}
    </div>

### Hijacking the Response

The `treetop.Response` type implements `http.ResponseWriter`. Any handler can halt the executor and
take full responsibility for the response by using one of the write methods of that interface<sup>2</sup>.

## Response Writer

The Treetop Go library provides utilities for writing ad-hoc template responses when needed.
PartialWriter and FragmentWriter wrap a supplied `http.ResponseWriter`.

For example,

    func myHandler(w http.ResponseWriter, req *http.Request) {
        // check for treetop request and construct a writer
        if tw, ok := treetop.NewFragmentWriter(w, req); ok {
            fmt.Fprintf(tw, `<h3 id="greeting">Hello, %s</h3>`, "Treetop")
        }
    }

This is useful when you want to use the template protocol with a conventional handler function.

The difference between a "Partial" and "Fragment" writer has to with navigation history
in the web browser <sup>[docs needed]</sup>

## Client Library

The client library is used to send template requests from the browser using XHR. Response fragments are handled mechanically on the client. This avoids the need for callbacks or other IO boilerplate.

    treetop.request("GET", "/some/path")

See [Client Library](https://github.com/rur/treetop-client) for more information.


## _Footnotes_
1. Go supports template inheritance through [nested template definitions](https://tip.golang.org/pkg/text/template/#hdr-Nested_template_definitions).
2. A [http.ResponseWriter](https://golang.org/pkg/net/http/#ResponseWriter) will flush headers when either `WriteHeaders(..)` or `Write(..)` methods are invoked.