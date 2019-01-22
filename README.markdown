[![Build Status](https://travis-ci.org/rur/treetop.svg?branch=master)](https://travis-ci.org/rur/treetop)

# Treetop

### Modern web UX for multi-page applications

_N.B. This is a prototype. The API is not stable and has not yet been extensively tested._

## TL;DR

Treetop is a library for managing HTTP requests that enable in-page browser navigation with server-side templates.

Try it yourself, clone the repo and run the example server.

    $ go run example/greeter.go

Tip. Activate your network tab to observe what's going on.

### Other Examples

Runnable example projects.

- [Todo \*Without\* MVC](https://github.com/rur/todowithoutmvc) - Treetop implementation of [TodoMVC](http://todomvc.com) app.

## Introduction

Multi-page navigation is ideal for web apps that are content heavy or have a sophisticated backend. The main drawback versus client apps is interactivity. Linking pages together doesn't always cut it in terms of user experience.

Treetop aims to enhance server-side web apps by allowing them to support partial page updates, without the need for custom client-side code. A Treetop enabled request handler is capable of yielding either a normal HTML document, or a list of HTML fragments. Fragments can be applied automatically to update the loaded page.

The goal of this project is to find a mechanism that is as close as possible to the standard HTTP application model.

### No client configuration necessary

A JS client library is provided to help mediate partial requests. Aside from optional component integration, no configuration is involved.

For more information see [Treetop Client Library](https://github.com/rur/treetop-client).


## How a Treetop Request Works

The client library uses XHR to fullfil in-page requests. Each treetop request includes the following accept header,

    Accept: application/x.treetop-html-partial+xml, application/x.treetop-html-fragment+xml

If the server path supports either content type, the response will include corresponding headers and a list of HTML snippets
to be applied to the current document. For example,

    HTTP/1.1 200 OK
    [...]
    Content-Type: application/x.treetop-html-fragment+xml
    Vary: Accept
    X-Response-URL: /some/path
    [...]

    <section id="content"><p>Hello, Treetop!</p></section>
    <div id="sidebar"><a href="/">Homepage</a></div>

Once the `Content-Type` has been recognized in the response headers, the client library will parse the body as a list of HTML fragments. The `id` attribute of each top level element will be matched to an existing node in the document.

* Matched elements in the current DOM will be replaced.
* Unmatched fragments will be silently discarded.


### Fragment vs Partial

Partial content type, `application/x.treetop-html-partial+xml`

* 'Part' of a full page,
* This endpoint supports rendering a valid HTML document,
* A new browser history entry will be pushed (updating the URL bar).

Fragment content type, `application/x.treetop-html-fragment+xml`

* Transient view update,
* This endpoint is not necessarily capable of yielding a valid HTML document.

## Server Side Handlers

The Treetop Go library provides utilities for writing compatible HTTP responses. Ad hoc integration is supported with a `ResponseWriter` wrapper like so,

    func myHandler(w http.ResponseWriter, req *http.Request) {
        // check for treetop request and construct a writer
        if tw, ok := treetop.Writer(w, req, false); ok {
            fmt.Fprintf(tw, `<h3 id="greeting">Hello, %s</h3>`, "Treetop")
        }
    }

### Hierarchical Handlers

The Treetop library includes an abstraction for creating more complex networks of handlers. The 'PageView' API is available for building handler instances which take advantage of the template inheritance feature supported by the Go standard library<sup>(1)</sup>.

    base := treetop.NewView("base.html.tmpl", baseHandler)
    landing := base.SubView(
        "content",
        "landing.html.tmpl",
        landingHandler,
    )
    contactForm := base.SubView(
        "content",
        "contact.html.tmpl",
        contactHandler,
    )
    submit := contactForm.SubView(
        "message",
        "contactSubmit.html.tmpl",
        contactSubmitHandler,
    )

    mux.Handler("/", treetop.ViewHandler(landing))
    mux.Handler("/contact", treetop.ViewHandler(contactForm))
    mux.Handler("/contact/submit", treetop.ViewHandler(submit).FragmentOnly())

All handler instances implement the `http.Handler` interface so you are free to use whatever routing library you wish.

Each template file path is paired with a data handler. This function is responsible for yielding execution data for the corresponding template. For example,


    func contactSubmitHandler(_ treetop.Response, req *http.Request) interface{} {
        // ...handle request here...
        // data for template
        return "Thanks friend!"
    }

Hierarchy works by chaining handlers together to assemble tiers of template data into one data structure.

    // top-level handler delegates 'content' data loading to a sub handler
    func baseHandler(rsp treetop.Response, req *http.Request) interface{} {
        return struct{
            Content interface{}
        }{
            Content: rsp.HandlePartial("content", req),
        }
    }

    // sub-view which has the option to delegate further
    func contactHandler(_ treetop.Response, _ *http.Request) interface{} {
        return "...Contact form template data..."
    }

The standard Go [html/template](https://golang.org/pkg/html/template/) library is used under the hood. However, a preferred engine can be configured without much fuss (once it supports inheritance).

_TODO: add examples and guides for 'Handling Inheritance'_

## Client Library

The __treetop.js__ script must be sourced by the browser to enable in-page navigation.

See [Client Library](https://github.com/rur/treetop-client) for more information.


## References
1. Go supports template inheritance through [nested template definitions](https://tip.golang.org/pkg/text/template/#hdr-Nested_template_definitions)
    * Treetop is intended for use with `{{block "name" ...}}` for convenience sake, `{{define "name"}}` could also be used. See Go [template actions](https://tip.golang.org/pkg/text/template/#hdr-Actions).
