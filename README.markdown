# Treetop

### Modern web UX for server side applications

_N.B. This is a prototype. The API is not stable and has not yet been extensively tested._

## TL;DR

Treetop is a library for managing HTTP requests that enable in-page browser navigation with server side templates.

Try it yourself, clone the repo and run the example server.

    $ go run example/greeter.go

_Example requires Go 1.6 or greater._

## Introduction

### Why was this created?

Integrating a modern web UI with a server side application can be frustrating. The single-page application approach ([SPA](https://en.wikipedia.org/wiki/Single-page_application)) is not always a good option. Treetop is a prototype which aims to extend the navigation model for multi-page apps with HTML partials and fragments. Insted of returning a full HTML page, the server can respond with 'updates' for the current web page.

#### No client configuration necessary

A lightweight JS library is the only thing required for the browser to facilitate in-page navigation. It is designed not to get in the way if other JS components.

See [Client Library](https://github.com/rur/treetop-client) for more information.


## How Treetop Requests Work

The client library uses XHR to fullfil in-page requests. These can be triggered explicitly in the browser like so,

    // JavaScript
    treetop.request("GET", "/some/path")

The request sent by the client will include the following headers

    GET /some/path HTTP/1.1
    Host: [...]
    Accept: application/x.treetop-html-partial+xml, application/x.treetop-html-fragment+xml
    [...]

The client library will be expecting the response from the server to looks like this,

    HTTP/1.1 200 OK
    [...]
    Content-Type: application/x.treetop-html-fragment+xml
    Vary: Accept
    X-Response-URL: /some/path
    [...]

    <section id="content"><p>Hello, Treetop!</p></section>

    <div id="sidebar"><a href="/">Go Home</a></div>

Once the `Content-Type` is a recognized (see below) the response body will be parsed as a list of HTMLElements. Each of which will be matched to an existing node in the document using element ID.

* Matched elements in the current DOM will be replaced.
* Unmatched elements from the response will be silently discarded.


#### Fragment vs Partial

Fragment content type, `application/x.treetop-html-fragment+xml`

* Transient view update,
* This endpoint is not necessarily capable of yielding a valid html document.

Partial content type, `application/x.treetop-html-partial+xml`

* HTML snippets are 'part' of a full page,
* A new browser history entry will be pushed (updating the URL bar),
* Browser window refresh/navigation must produce a valid HTML document.
    * Not necessarily the same HTML document.

## Server Side Handlers

The Treetop Go library provides utilities for writing compatible HTTP responses. Ad hoc integration is supported with a `ResponseWriter` wrapper like so,

    func myHandler(w http.ResponseWriter, req *http.Request) {
        // check for treetop request and construct a writer
        if tw, ok := treetop.FragmentWriter(w, req); ok {
            fmt.Fprintf(tw, `<h3 id="greeting">Hello, %s</h3>`, "Treetop")
        }
    }

### Page Handlers

The Treetop library includes a `Page` abstraction for creating more complex networks of handlers. It was designed to take advantage of template inheritance<sup>1</sup>, which is supported by the Go standard library. _Page_, _Partial_ and _Fragment_ instances implement the `http.Handler` interface, so you are free to use whatever routing library you wish.

    base := treetop.Page("base.templ.html", treetop.Noop)
    content := base.Block("content")
    landing := content.Partial("landing.templ.html", treetop.Noop)
    contactForm := content.Partial("contact.templ.html", treetop.Noop)
    message := contactForm.Block("message")
    thanks := message.Fragment("contact-complete.templ.html", contactSubmitHandler)

    mux.Handler("/", landing)
    mux.Handler("/contact", contactForm)
    mux.Handler("/contact/submit", thanks)

Notice that each template file path is paired with a function. This handler function is responsible for yielding the template data from the request. For example,

    func contactSubmitHandler(dw treetop.DataWriter, req *http.Request) {
        // do stuff...
        dw.Data("Thanks!")
    }

The standard Go [html/template](https://golang.org/pkg/html/template/) library is used under the hood, however a preferred engine can be configured without much fuss (once it supports inheritance).

See [Handling Inheritance](#TODO) for more details.

## Client Library

The __treetop.js__ script must be sourced by the browser to enable in-page navigation.

See [Client Library](https://github.com/rur/treetop-client) for more information.


## References
1. Go supports template inheritance through [nested template definitions](https://tip.golang.org/pkg/text/template/#hdr-Nested_template_definitions)
    * Treetop is intended for use with `{{block "name" ...}}` for convenience sake, `{{define "name"}}` could also be used however. See Go [template actions](https://tip.golang.org/pkg/text/template/#hdr-Actions).
