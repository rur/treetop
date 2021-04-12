[![Go Report Card](https://goreportcard.com/badge/rur/treetop)](https://goreportcard.com/report/rur/treetop) [![Coverage Status](https://coveralls.io/repos/github/rur/treetop/badge.svg?branch=master)](https://coveralls.io/github/rur/treetop?branch=master) [![Build Status](https://travis-ci.org/rur/treetop.svg?branch=master)](https://travis-ci.org/rur/treetop)

# Treetop

### A tool to create request handlers for nested templates in Go

[![GoDoc](https://godoc.org/github.com/rur/treetop?status.svg)](https://godoc.org/github.com/rur/treetop)

The Go standard library has powerful support for nested templates <sup>[[html/template](https://golang.org/pkg/html/template/)]</sup>.
The Treetop library aims to make it easier to construct HTML endpoints from a hierarchy of reusable fragments.

- Aims to be as lightweight as possible
- No 3rd-party dependencies
- Support for HTML hot-swapping to achieve interactivity (see [Online DEMO](https://treetop-demo.herokuapp.com/))

#### Template Hierarchy

Parent and child templates are matched with individual functions. Template trees can then be bound together in
various configurations to construct endpoints.

                  BaseHandlerFunc(...)
                | base.html ========================|
                | …                                 |
                | {{ template "content" .Content }} |
                | …               ^                 |
                |_________________|_________________|
                                  |
                           ______/ \______
      ContentAHandlerFunc(...)            ContentBHandlerFunc(...)
    | contentA.html ========== |        | contentB.html ========== |
    |                          |        |                          |
    | <div id="content">...</… |        | <div id="content">...</… |
    |__________________________|        |__________________________|

_Basic example of a page hierarchy showing content A and B sharing the same 'base' template_

The code below is an extension of this example. It binds the routes `"/content_a"` and `"/content_b"` with two
handlers that share the same "base", "nav" and "sidebar" templates.

    base := treetop.NewView("base.html", BaseHandler)
    nav := base.NewSubView("nav", "nav.html", NavHandler)
    _ = base.NewDefaultSubView("sidebar", "sidebar.html", SidebarHandler)
    contentA := base.NewSubView("content", "content_a.html", ContentAHandler)
    contentB := base.NewSubView("content", "content_b.html", ContentBHandler)

    exec := treetop.FileExecutor{}
    mux.Handle("/content_a", exec.NewViewHandler(contentA, nav))
    mux.Handle("/content_b", exec.NewViewHandler(contentB, nav))

The 'Executor' is responsible for collecting related views,
configuring templates and plumbing it all together to produce a `http.Handler` instance
for each route.

Example of embedded template blocks in `"base.html"`,

    ...
    <div class="layout">
    	{{ block "nav" .Nav }}  <div id="nav">default nav</div> {{ end }}

    	{{ template "sidebar" .SideBar }}

    	{{ template "content" .Content }}
    </div>
    ...

_See text/template [Nested Template Definitions](https://tip.golang.org/pkg/text/template/#hdr-Nested_template_definitions) for more info._

Views can have as many levels of nesting as needed.

### No Third-Party Dependencies

The Treetop package wraps features of the Go standard library, mostly within "net/http" and "html/template".

## HTML Template Protocol

### Hot-swap sections of a page without JS boilerplate

#### [Online DEMO](https://treetop-demo.herokuapp.com/)

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

A [Treetop Client Library](https://github.com/rur/treetop-client) is available.
It sends template requests using XHR and applies fragments to the DOM with a simple
find and replace mechanism.

Hot-swapping can be used to improve user experience in several ways.
See demo for more details.

## Examples

#### Demo Apps ([README](https://github.com/rur/treetop-demo#treetop-demo) / [DEMO](https://treetop-demo.herokuapp.com/))

Demo can be run locally by cloning the [treetop-demo](https://github.com/rur/treetop-demo) repo and running the command,

    $ git clone https://github.com/rur/treetop-demo.git
    ...
    $ cd treetop-demo
    $ go run . 8080
    serving on http://0.0.0.0:8080/

### Other Examples:

- [Todo \*Without\* MVC](https://github.com/rur/todowithoutmvc) - Treetop implementation of [TodoMVC](http://todomvc.com) app using the template protocol.

## Template Executor

An 'Executor' is responsible for loading and configuring templates. It constructs a
[HTTP Handler](https://golang.org/pkg/net/http/#Handler) instance to manage the plumbing
between loading data and executing templates for a request. You can implement your own template
loader <sup>[docs needed]</sup>, but the following are provided:

- `FileExecutor` - load template files using os.Open
- `FileSytemExecutor` - loads templates from a supplied http.FileSystem instance
- `StringExecutor` - treat the view template property as an inline template string
- `KeyedStringExecutor` - treat the view template property is key into a template map
- `DeveloperExecutor` - force per request template parsing

## View Handler Function

A view handler function loads data for a corresponding Go template. Just as nested templates
are embedded in a parent, nested handler data is embedded in the _data_ of it's parent.

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
take full responsibility for the response by using one of the 'Write' methods of that interface<sup>2</sup>. This is a common and useful practice making things like error messages and redirects possible.

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

The client library is used to send template requests from the browser using XHR. Response fragments
are handled mechanically on the client. This avoids the need for callbacks or other IO boilerplate.

    treetop.request("GET", "/some/path")

See [Client Library](https://github.com/rur/treetop-client) for more information.

## _Footnotes_

1. Go supports template inheritance through [nested template definitions](https://tip.golang.org/pkg/text/template/#hdr-Nested_template_definitions).
2. A [http.ResponseWriter](https://golang.org/pkg/net/http/#ResponseWriter) will flush headers when either `WriteHeaders(..)` or `Write(..)` methods are invoked.
