# Treetop

### Modern web UX for server side applications

_N.B. This is a prototype. The API is not stable and has not yet been extensively tested._

## TL;DR

Treetop is a library for managing HTTP requests that enable in-page browser navigation without client side templates.

Try it yourself, clone the repo and run the example server.

    $ go run example/main.go

_Example requires Go 1.6 or greater._

## Introduction

### Why was this created?

Integrating a modern web UI with a server side application can be frustrating. When a single-page application ([SPA](https://en.wikipedia.org/wiki/Single-page_application)) is not an option, hybrid client & server templates are needed. This blurs the lines between conventional page navigation and data-driven components. Dual APIs have severe limitations that can cause maintenance headaches over time.

Treetop is a prototype which aims to bridge this gap by extending the standard model of HTML page navigation with partials and fragments. This helps alleviate the need for client side components to fetch data from the server, instead context can be pushed via HTML in the conventional way.

### No client configuration necessary

A lightweight JS library is the only thing required on the client to facilitate in-page navigation. It is fairly unobtrusive and follows an 'opt in' activation principle. Configuration is not required. Simple declarative-style script bindings are available for custom components.

See [Client Library](#TODO) for more information.

## How Treetop Requests Work

The client library uses XHR to fullfil in-page requests. These can be triggered explicitly in the browser like so,

    // JavaScript
    treetop.request("GET", "/some/path")

The client library will be expecting a response from the server that looks like this,

    HTTP/1.1 200 OK
    [...]
    Content-Type: application/x.treetop-html-fragment+xml
    Vary: Content-Type
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

### The `treetop` Attribute

The most convenient way to enable in-page nav is declaratively. The behavior of specific elements can be overloaded by adding a `treetop` attribute. This allows your template to decide which navigation actions should trigger full-page vs. in-page loading.

Here are two examples:

```
<a treetop href="/some/path">treetop link</a>

<script>
// equivalent to
treetop.request("GET", "/some/path")
</script>

<form treetop action="/some/path" method="POST">
    <input name="field" value="123"  type="submit"/>
</form>

<script>
// equivalent to
treetop.request("POST", "/some/path", {field: 123})
</script>
```

### Client Library API

The client library exposes the `window.treetop` instance with the following methods:

#### treetop.request
Issue a treetop request. Notice that no callback mechanism is available. This is by design. Response handling is mandated by the protocol, see [Treetop Request](#Treetop+Request)

##### Usage
```
treetop.request( [method], [url], [data], [encoding], [suppressPushState])
```

##### Arguments:

| Param             | Type    | Details                                          |
|-------------------|---------|--------------------------------------------------|
| method            | string  | The HTTP request method  to use                  |
| url               | string  | The URL path                                     |
| data              | *any       | the request data, pre-encoding                   |
| encoding          | *string  | encoding request data serialized as, default urlencoded |
| suppressPushState | *boolean | stipulate that a history entry must NOT be created |

_*optional_

### treetop.push

Mount our un-mount a component script to any element matching either tagName or attrName property after it has be inserted (mount) or removed (unmount) by treetop during the course of handing a DOM update. This is useful for bootstrapping/deconstructing custom components.

All child elements are 'mounted' and 'unmounted', not just the fragment root.

#### Usage
```
window.treetop = (window.treetop || []).push({
    tagName: "",
    attrName: "",
    mount: (el) => {},
    unmount: (el) => {},
})
```

#### Arguments:

| Param             |  Type      | Details                                         |
|-------------------|------------|-------------------------------------------------|
| tagName           | *string    | Case insensitive HTMLElement tag name           |
| attrName          | *string    | Case insensitive HTMLElement attr name          |
| mount             | *function  | Function accepting the HTMLElement as parameter |
| unmount           | *function  | Function accepting the HTMLElement as parameter |

_*optional_

### Browser support

__TODO: More browser testing is needed, please help!__

The client library aims to be backwards compatible with legacy browsers for the most part. A HTML5 history shim is needed for the full experience in older browsers.

## Handling Inheritance

(TODO)

## References
1. Go supports template inheritance through [nested template definitions](https://tip.golang.org/pkg/text/template/#hdr-Nested_template_definitions)
    * Treetop is intended for use with `{{block "name" ...}}` for convenience sake, `{{define "name"}}` could also be used however. See Go [template actions](https://tip.golang.org/pkg/text/template/#hdr-Actions).