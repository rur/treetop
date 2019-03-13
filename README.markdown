[![Build Status](https://travis-ci.org/rur/treetop.svg?branch=master)](https://travis-ci.org/rur/treetop)

# Treetop

## Modern UX for multi-page web applications

An uncomplicated approach to help eliminate business logic from client-side code.

### Please note - This is a prototype

The API is not stable and has not yet been extensively tested.

## TL;DR

Try it yourself, clone the repo and run the example server.

    $ go run example/greeter.go

Tip. Activate your network tab to observe what's going on.

### Other Examples

- [Todo \*Without\* MVC](https://github.com/rur/todowithoutmvc) - Treetop implementation of [TodoMVC](http://todomvc.com) app.

## Introduction

There are times when application logic is best kept on the server side. Treetop was created to help improve user experience without the need for client facing APIs.

Treetop supports partial page updates with a minimal extension to standard web navigation. An endpoint is enabled if it is capable of yielding either a normal HTML document, or a list of HTML fragments. Fragments that can be applied to update a loaded page.


### No client configuration necessary

A JS client library is provided to help mediate partial requests. Aside from optional component integration, no configuration is needed. For more information see [Treetop Client Library](https://github.com/rur/treetop-client).


## How a partial request works

A Treetop request is triggered in the browser using the [treetop client](https://github.com/rur/treetop-client) like so,

    treetop.request("GET", "/some/path")

An XHR request is sent that includes the following accept header,

    Accept: application/x.treetop-html-partial+xml, application/x.treetop-html-fragment+xml

If the Treetop content type is supported at that end-point, the response header will specify a partial or fragment response. The body will contain a list of HTML snippets to be applied to the current document.

For example,

    HTTP/1.1 200 OK
    [...]
    Content-Type: application/x.treetop-html-partial+xml
    Vary: Accept
    X-Response-URL: /some/path
    [...]

    <section id="content"><p>Hello, Treetop!</p></section>
    <div id="sidebar"><a href="/">Homepage</a></div>

Finally, once the `Content-Type` has been recognized in the response headers, the client library will parse the body as a list of HTML fragments. The `id` attribute of each top level element will be matched to an existing node in the document.

* Matched elements in the current DOM will be replaced.
* Unmatched fragments will be discarded.

_Note that aspects of the client processing can be configured and extended._


### Fragment vs Partial

The `Content-Type` response header denotes whether a request should be treated as a partial or a fragment by the client.

#### Partial content type

    application/x.treetop-html-partial+xml

A 'partial' URL supports rendering either a fragment or a full HTML document. When this Content-Type is received the [treetop client](https://github.com/rur/treetop-client) will update the browser location bar with the response URL.

#### Fragment content type

    application/x.treetop-html-fragment+xml

The contents of this response should be treated as a transient view update. The URL is not necessarily capable of yielding a valid HTML document so the location bar will not change.

## Server Side Helpers

### Response Writer

The Treetop Go library provides utilities for writing compatible HTTP responses. Ad hoc integration is supported with a `ResponseWriter` wrapper like so,

    func myHandler(w http.ResponseWriter, req *http.Request) {
        // check for treetop request and construct a writer
        if tw, ok := treetop.Writer(w, req, false); ok {
            fmt.Fprintf(tw, `<h3 id="greeting">Hello, %s</h3>`, "Treetop")
        }
    }

### Hierarchical Views

The Treetop library includes an abstraction for creating more complex networks of handlers. A page view API is available for building handler instances that take advantage of the template inheritance feature supported by the Go standard library<sup>(1)</sup>.

    base := treetop.NewView("base.html.tmpl", baseHandler)
    content := base.SubView(
        "content",
        "content.html.tmpl",
        contentHandler,
    )
    form := content.SubView(
        "form",
        "contact.html.tmpl",
        contactHandler,
    )
    submit := content.SubView(
        "form",
        "contactSubmit.html.tmpl",
        submitHandler,
    )

    mux.HandleGET("/", treetop.ViewHandler(content))
    mux.HandleGET("/contact", treetop.ViewHandler(form))
    mux.HandlePOST("/contact/submit", treetop.ViewHandler(submit).FragmentOnly())

All handler instances implement the `http.Handler` interface so you are free to use whatever routing library you wish.

Each template file path is paired with a data handler. This function is responsible for yielding execution data for the corresponding template. Hierarchy works by chaining handlers together to assemble tiers of template data into one data structure.

    // top-level handler delegates to zero or more sub handler
    func baseHandler(rsp treetop.Response, req *http.Request) interface{} {
        return struct{
            Content interface{}
        }{
            Content: rsp.HandlePartial("content", req),
        }
    }

    // "content" subview, delegates to "form"
    func contentHandler(rsp treetop.Response, req *http.Request) interface{} {
        return struct{
            Form interface{}
        }{
            Form: rsp.HandlePartial("form", req),
        }
    }

    // "form" sub-view
    func contactHandler(_ treetop.Response, _ *http.Request) interface{} {
        return "...Contact form config..."
    }

    // alternative "form" handler
    func submitHandler(_ treetop.Response, req *http.Request) interface{} {
        // ...handle POST data here...
        // output for template
        return "Thanks!"
    }


The standard Go [html/template](https://golang.org/pkg/html/template/) library is used under the hood. However, a preferred engine can be configured without much fuss (once it supports inheritance).

_TODO: This feature should have dedicated documentation._

## Client Library

The __treetop.js__ script must be sourced by the browser to enable in-page navigation.

See [Client Library](https://github.com/rur/treetop-client) for more information.


## References
1. Go supports template inheritance through [nested template definitions](https://tip.golang.org/pkg/text/template/#hdr-Nested_template_definitions).
