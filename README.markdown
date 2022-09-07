[![Go Report Card](https://goreportcard.com/badge/rur/treetop)](https://goreportcard.com/report/rur/treetop) [![Coverage Status](https://coveralls.io/repos/github/rur/treetop/badge.svg?branch=master)](https://coveralls.io/github/rur/treetop?branch=master) [![Build Status](https://travis-ci.org/rur/treetop.svg?branch=master)](https://travis-ci.org/rur/treetop)

# Treetop

[![GoDoc](https://godoc.org/github.com/rur/treetop?status.svg)](https://godoc.org/github.com/rur/treetop)

### A lightweight bridge from _net/http_ to _html/template_

Build HTML endpoints using a hierarchy of nested templates, as supported by the Go standard library <sup>[[html/template](https://golang.org/pkg/html/template/)]</sup>.

- Lightweight by design
- No 3rd-party dependencies
- Optional protocol extension for fragment hot-swapping <sup>[[1](#request-protocol-extension)]

#### Template Hierarchy

Each template is paired with an individual data loading function.
Pages are constructed by combining templates in different configurations and
binding endpoints to your application router.

```
                         BaseFunc(…)
            ┌──────────── base.html ─────────────┐
            │ <html>                             │
            │ …                                  │
            │ {{ template "content" .Content }}  │
            │ …               ▲                  │
            │ </html>         │                  │
            │~~~~~~~~~~~~~~~~ │ ~~~~~~~~~~~~~~~~~│
                              │
                         ┌────┴────┐
        ContentAFunc(…)  │         │   ContentBFunc(…)
  ┌──── content_a.html ──┴──┐   ┌──┴── content_b.html ─────┐
  │                         │   │                          │
  │ <div id="content">A</…  │   │  <div id="content">B</…  │
  │~~~~~~~~~~~~~~~~~~~~~~~~~│   │~~~~~~~~~~~~~~~~~~~~~~~~~~│

```

_Basic example of a page hierarchy showing content A and B sharing the same 'base' template_

**Note.** Nested hierarchy is supported, see Golang docs for details <sup>[[doc](https://tip.golang.org/pkg/text/template/#hdr-Nested_template_definitions)]</sup>

### Library Overview

The code example shows how `http.Handler` endpoints are created from 
a hierarchy of templates and handler pairs. 

The example binds the `"/content_a"` and `"/content_b"` routes shown in the diagram 
with additional "nav" and "sidebar" templates.

    base := treetop.NewView("base.html", BaseHandler)
    nav := base.NewSubView("nav", "nav.html", NavHandler)
    _ = base.NewDefaultSubView("sidebar", "sidebar.html", SidebarHandler)
    contentA := base.NewSubView("content", "content_a.html", ContentAHandler)
    contentB := base.NewSubView("content", "content_b.html", ContentBHandler)

    exec := treetop.FileExecutor{}
    myMux.Handle("/content_a", exec.NewViewHandler(contentA, nav))
    myMux.Handle("/content_b", exec.NewViewHandler(contentB, nav))


Note that `exec := treetop.FileExecutor{}` is responsible for compiling a template tree
from your view definitions, plumbing togeather handler functions and exposing a `http.Handler` 
interface to your application router.

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
        // data for `base.html` template
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
        // data for the `content_a.html` template
        return struct {
            ...
        }{
            ...
        }
    }

### No Third-Party Dependencies

The Treetop package wraps features of the Go standard library, mostly within "net/http" and "html/template".

## Request Protocol Extension

Hot-swap sections of a page with minimal boilerplate.

#### [DEMO](https://treetop-demo.herokuapp.com/)

Since Treetop views are self-contained, they can be rendered in isolation. Treetop
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

A [Client Library](https://github.com/rur/treetop-client) handles the
the client side of the protocol, passively updating the DOM based on the response. 

### Examples

- Demo Apps ([README](https://github.com/rur/treetop-demo#treetop-demo) / [DEMO](https://treetop-demo.herokuapp.com/))
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
