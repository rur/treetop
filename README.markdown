[![Go Report Card](https://goreportcard.com/badge/rur/treetop)](https://goreportcard.com/report/rur/treetop) [![Coverage Status](https://coveralls.io/repos/github/rur/treetop/badge.svg?branch=master)](https://coveralls.io/github/rur/treetop?branch=master) [![Build Status](https://travis-ci.org/rur/treetop.svg?branch=master)](https://travis-ci.org/rur/treetop)

# Treetop

[![GoDoc](https://godoc.org/github.com/rur/treetop?status.svg)](https://godoc.org/github.com/rur/treetop)

## Bridging _net/http_ with _html/template_

### Create request handlers for nested templates in Go

The Treetop library makes it easier to build HTML endpoints using a hierarchy of nested templates, as supported by the Go standard library <sup>[[html/template](https://golang.org/pkg/html/template/)]</sup>.

- Lightweight by design
- No 3rd-party dependencies
- Protocol for HTML hot-swapping (see [Online DEMO](https://treetop-demo.herokuapp.com/))

#### Template Hierarchy

Parent and child templates are paired with individual handler functions.
You can build pages by combining views in different ways and
binding endpoints to your application router.

                             BaseFunc(…)
                |============ base.html =============|
                | <html>                             |
                | …                                  |
                | {{ template "content" .Content }}  |
                | …               /\                 |
                | </html>         ||                 |
                |________________ || ________________|
                                 /  \
                                /    \
            ContentAFunc(…)    /      \   ContentBFunc(…)
      |==== content_a.html =====|   |==== content_b.html =====|
      |                         |   |                         |
      | <div id="content">A</…  |   | <div id="content">B</…  |
      |_________________________|   |_________________________|

_Basic example of a page hierarchy showing content A and B sharing the same 'base' template_

**Note.** Multiple levels of hierarchy are supported, see Golang doc for details [[doc](https://tip.golang.org/pkg/text/template/#hdr-Nested_template_definitions)]

### API Overview

The code below is an extension of the example hierarchy. It binds the routes `"/content_a"` and `"/content_b"` with two
handlers that share the same "base", "nav" and "sidebar" templates.

    base := treetop.NewView("base.html", BaseHandler)
    nav := base.NewSubView("nav", "nav.html", NavHandler)
    _ = base.NewDefaultSubView("sidebar", "sidebar.html", SidebarHandler)
    contentA := base.NewSubView("content", "content_a.html", ContentAHandler)
    contentB := base.NewSubView("content", "content_b.html", ContentBHandler)

    exec := treetop.FileExecutor{}
    myMux.Handle("/content_a", exec.NewViewHandler(contentA, nav))
    myMux.Handle("/content_b", exec.NewViewHandler(contentB, nav))

#### Template Executor

The 'Executor' is responsible for collecting related views,
configuring templates and plumbing it all together to produce a `http.Handler` instance
for your router.

Example of embedded template blocks in `"base.html"`,

    ...
    <div class="layout">
    	{{ block "nav" .Nav }}  <div id="nav">default nav</div> {{ end }}

    	{{ template "sidebar" .SideBar }}

    	{{ template "content" .Content }}
    </div>
    ...

_See text/template [Nested Template Definitions](https://tip.golang.org/pkg/text/template/#hdr-Nested_template_definitions) for more info._

Note the loose coupling between content handlers in the outline below.

    func BaseHandler(rsp treetop.Response, req *http.Request) interface{} {
        // data for base template
        return struct {
            ...
        }{
            ...
            Nav: rsp.HandleSubView("nav", req),
            SideBar: rsp.HandleSubView("sidebar", req),
            Content: rsp.HandleSubView("content", req),
        }
    }

    func ContentAHandler(rsp treetop.Response, req *http.Request) interface{} {
        // data for Content A template
        return ...
    }

### No Third-Party Dependencies

The Treetop package wraps features of the Go standard library, mostly within "net/http" and "html/template".

## HTML Template Protocol

### Hot-swap sections of a page without JS boilerplate

#### [Online DEMO](https://treetop-demo.herokuapp.com/)

Since views are self-contained, they can be rendered in isolation. Treetop
handlers support rendering template fragments that can be applied to a loaded document.
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

Hot-swapping can be used to enhance user experience in several ways.
See demo for more details.

## Examples

- Demo Apps ([README](https://github.com/rur/treetop-demo#treetop-demo) / [Online DEMO](https://treetop-demo.herokuapp.com/))
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

<a name="ref_1"></a>1. Go supports template inheritance through [nested template definitions](https://tip.golang.org/pkg/text/template/#hdr-Nested_template_definitions). 2. A [http.ResponseWriter](https://golang.org/pkg/net/http/#ResponseWriter) will flush headers when either `WriteHeaders(..)` or `Write(..)` methods are invoked.
