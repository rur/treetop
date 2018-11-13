[![Build Status](https://travis-ci.org/rur/treetop.svg?branch=master)](https://travis-ci.org/rur/treetop)

# Treetop

### Modern web UX for multi-page applications

_N.B. This is a prototype. The API is not stable and has not yet been extensively tested._

## TL;DR

Treetop is a library for managing HTTP requests that enable in-page browser navigation with server-side templates.

Try it yourself, clone the repo and run the example server.

    $ go run example/greeter.go

Tip. Activate your network tab to observe what's going on.

_TODO: Add more examples_

## Introduction

### Why was this created?

Multi-page web applications are secure, well supported and promote loose coupling. They are a great solution for systems with a broad range of content or that encapsulate other complex systems. The main drawback versus native or single-page apps is interactivity. Modern software must deliver a modern user experience; linking documents together doesn't always cut it.

Treetop is a prototype which aims to close the gap by extending conventional multi-page endpoints with 'partials' and 'fragments'. Treetop endpoints are capable of rendering a valid HTML document, or snippets for modifying a loaded page, depending upon which is requested.

The main focus of this implementation is to find an uncomplicated mechanism to achieve this, while staying as close as possible to the standard model of HTML over HTTP.

#### No client configuration necessary

A JavaScript library is provided to help negotiate Treetop requests in the browser. While some component hooks are supported, this library does not require any specific configuration.

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

    base := treetop.NewView("base.templ.html", baseHandler)
    landing := base.SubView(
        "content",
        "landing.templ.html",
        landingHandler,
    )
    contactForm := base.SubView(
        "content",
        "contact.templ.html",
        contactHandler,
    )
    submit := contactForm.SubView(
        "message",
        "contactSubmit.templ.html",
        contactSubmitHandler,
    )

    mux.Handler("/", landing.PartialHandler())
    mux.Handler("/contact", contactForm.PartialHandler())
    mux.Handler("/contact/submit", submit.FragmentHandler())

All handler instances implement the `http.Handler` interface so you are free to use whatever routing library you wish.

Each template file path is paired with a data handler. This function is responsible for yielding execution data for the corresponding template. For example,

    func contactSubmitHandler(dw treetop.DataWriter, req *http.Request) {
        // do stuff...
        dw.Data("Say Thanks!")
    }

The standard Go [html/template](https://golang.org/pkg/html/template/) library is used under the hood. However, a preferred engine can be configured without much fuss (once it supports inheritance).

_TODO: add examples and guides for 'Handling Inheritance'_

## Client Library

The __treetop.js__ script must be sourced by the browser to enable in-page navigation.

See [Client Library](https://github.com/rur/treetop-client) for more information.


## References
1. Go supports template inheritance through [nested template definitions](https://tip.golang.org/pkg/text/template/#hdr-Nested_template_definitions)
    * Treetop is intended for use with `{{block "name" ...}}` for convenience sake, `{{define "name"}}` could also be used. See Go [template actions](https://tip.golang.org/pkg/text/template/#hdr-Actions).
