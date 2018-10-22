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

As an approach to software development, conventional multi-page web applications are secure, well supported, maintainable and loosely coupled. They are a great solution for systems with a broad range of content or that encapsulate other complex systems. The main drawback versus native or single-page apps is usability. Modern applications must deliver an efficient user experience; linking documents doesn't always cut it.

Treetop is a prototype which aims to close the gap by extending the navigation model for multi-page apps with HTML partials and fragments. In addition to generating a full HTML page, a Treetop enabled endpoint can render HTML update fragments in response to HTTP requests.

The main focus of this implementation is to find an uncomplicated mechanism to achieve this while staying as close as possible to the standard HTTP + HTML application model.

#### No client configuration necessary

A lightweight JS library is the only thing required in the browser to facilitate in-page navigation. JavaScript component hooks are supported. For more information see [Treetop Client Library](https://github.com/rur/treetop-client).


## How a Treetop Request Works

The client library uses XHR to fullfil in-page requests. Each treetop request includes the following accept header,

    Accept: application/x.treetop-html-partial+xml, application/x.treetop-html-fragment+xml

If this server endpoint supports this content type, the response will include corresponding headers and a list of HTML snippets
to be applied to the current document. For example,

    HTTP/1.1 200 OK
    [...]
    Content-Type: application/x.treetop-html-fragment+xml
    Vary: Accept
    X-Response-URL: /some/path
    [...]

    <section id="content"><p>Hello, Treetop!</p></section>
    <div id="sidebar"><a href="/">Homepage</a></div>

Once the `Content-Type` has been recognized in the response headers, the client library will parse the body as an HTMLTemplate. The `id` attribute of each top level element will be matched to an existing node in the document.

* Matched elements in the current DOM will be replaced.
* Unmatched elements from the response will be discarded.


### Fragment vs Partial

Fragment content type, `application/x.treetop-html-fragment+xml`

* Transient view update,
* This endpoint is not necessarily capable of yielding a valid HTML document.

Partial content type, `application/x.treetop-html-partial+xml`

* 'Part' of a full page,
* This endpoint supports rendering a valid HTML document.
* A new browser history entry will be pushed (updating the URL bar),

## Server Side Handlers

The Treetop Go library provides utilities for writing compatible HTTP responses. Ad hoc integration is supported with a `ResponseWriter` wrapper like so,

    func myHandler(w http.ResponseWriter, req *http.Request) {
        // check for treetop request and construct a writer
        if tw, ok := treetop.Writer(w, req, false); ok {
            fmt.Fprintf(tw, `<h3 id="greeting">Hello, %s</h3>`, "Treetop")
        }
    }

### Hierarchical Handlers

The Treetop library includes an abstraction for creating more complex networks of handlers. A definition API is available for building handler instances which take advantage of the template inheritance feature supported by the Go standard library<sup>(1)</sup>.

    base := treetop.Define("base.templ.html", baseHandler)
    content := base.Block("content")
    landing := content.Extend("landing.templ.html", landingHandler)
    contactForm := content.Extend("contact.templ.html", contactHandler)
    message := contactForm.Block("message")
    submit := message.Extend("contactSubmit.templ.html", contactSubmitHandler)

    mux.Handler("/", landing.PartialHandler())
    mux.Handler("/contact", contactForm.PartialHandler())
    mux.Handler("/contact/submit", submit.FragmentHandler())

These handlers implement the `http.Handler` interface so you are free to use whatever routing library you wish.

Notice that each template file path is paired with a function. This is responsible for yielding the template data from the request. For example,

    func contactSubmitHandler(dw treetop.DataWriter, req *http.Request) {
        // do stuff...
        dw.Data("Thanks!")
    }

The standard Go [html/template](https://golang.org/pkg/html/template/) library is used under the hood. However, a preferred engine can be configured without much fuss (once it supports inheritance).

_TODO: add examples and guides for 'Handling Inheritance'_

## Client Library

The __treetop.js__ script must be sourced by the browser to enable in-page navigation.

See [Client Library](https://github.com/rur/treetop-client) for more information.


## References
1. Go supports template inheritance through [nested template definitions](https://tip.golang.org/pkg/text/template/#hdr-Nested_template_definitions)
    * Treetop is intended for use with `{{block "name" ...}}` for convenience sake, `{{define "name"}}` could also be used. See Go [template actions](https://tip.golang.org/pkg/text/template/#hdr-Actions).
