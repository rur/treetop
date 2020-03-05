[![Build Status](https://travis-ci.org/rur/treetop.svg?branch=v0.3.0)](https://travis-ci.org/rur/treetop)

# Treetop

## Hierarchical Web Handlers

Take advantage of Go's nested template support<sup>1</sup> to create handlers that mix and match reusable templates and functions.

The following example defines a simple view hierarchy which is used it to generate handlers for two routes: `"/content_a"` and `"/content_b"`.

    base := treetop.NewView(
        "base.html.tmpl", BaseHandler)
    nav := base.NewSubView(
        "nav", "nav.html.tmpl", NavHandler)
    _ = base.NewDefaultSubView(
        "sidebar", "sidebar.html.tmpl", SidebarHandler)
    contentA := base.NewSubView(
        "content", "content_a.html.tmpl", ContentAHandler)
    contentB := base.NewSubView(
        "content", "content_b.html.tmpl", ContentBHandler)

    exec := treetop.FileExecutor{}
    mux.Handle("/content_a", exec.NewViewHandler(contentA, nav))
    mux.Handle("/content_b", exec.NewViewHandler(contentB, nav))

Example `"base.html.tmpl"`

	...
	<div id="base">
		{{ block "nav" .Nav }}  <div id="nav">default nav</div> {{ end }}

		{{ template "sidebar" .SideBar }}

		{{ template "content" .Content }}
	</div>
	...

This is beneficial for HTML web apps which have many endpoints using similar structural components.

### No 3rd-Party Dependencies!

The treetop package wraps features of the Go standard library only, mostly within "net/http" and "html/template".

## Example

A very basic example can be run from this repo <sup>[needs improvement]</sup>.

    $ go run example/

Tip. Activate your network tab to observe what's going on.


##  HTML Template Protocol

Self-contained views have a special benefit, sections of a page can be rendered in isolation. Thus, Treetop handlers are capable of generating a list of HTML template fragments which can be 'applied' to the requesting document.

    > GET / HTTP/1.1
      Accept: application/x.treetop-html-template+xml

    < HTTP/1.1 200 OK
      Content-Type: application/x.treetop-html-template+xml
      Vary: Accept

      <template>
          <div id="content">...</div>
          <div id="nav">...</div>
      </template>

There are many ways that this can improve user experience <sup>[docs needed]</sup>. A [Treetop Client Library](https://github.com/rur/treetop-client) is available to manage these requests using XHR. It applies templates to the DOM with a very simple find and replace mechanism.

### Protocol Examples:

- [Todo \*Without\* MVC](https://github.com/rur/todowithoutmvc) - Treetop implementation of [TodoMVC](http://todomvc.com) app using the template protocol.

## Template Executor

The 'Executor' is responsible for loading and configuring the templates. It create a [HTTP Handler](https://golang.org/pkg/net/http/#Handler) which will manage the plumbing to serve requests between loading data and executing templates. You can implement your own template loader <sup>[docs needed]</sup>, the following are provided:
- `FileExecutor` - load template files using os.Open
- `FileSytemExecutor` - loads templates from a supplied http.FileSystem instance
- `StringExecutor` - treat the view template property as an inline template string
- `KeyedStringExecutor` - treat the view template property is key into a template map
- `DeveloperExecutor` - force per request template parsing


## View Handlers

View handlers load data for a corresponding Go template. Just as nested templates are embedded in their parent, nested handler data is embedded in the data of it's parent. Example of a child handler passing data _back_ to the parent,

    func ParentHandler(rsp treetop.Response, req *http.Request) interface{} {
        return struct {
            ...
            Child interface{}
        }{
            ...
            Child: rsp.HandleSubView("child", req),
        }
    }

Data is subsequently passed _down_ within the template like so,

    <div id="parent">
        ...
        {{ template "child" .Child }}
    </div>


## Response Writer

The Treetop Go library provides utilities for writing ad-hoc template responses when needed. PartialWriter and FragmentWriter wrap the `http.ResponseWriter`,

    func myHandler(w http.ResponseWriter, req *http.Request) {
        // check for treetop request and construct a writer
        if tw, ok := treetop.NewFragmentWriter(w, req); ok {
            fmt.Fprintf(tw, `<h3 id="greeting">Hello, %s</h3>`, "Treetop")
        }
    }


## Client Library

The client library will parse the HTML template and use those fragments to update the DOM. The `id` attribute of each top level element will be matched to an existing node in the document.

* Matched elements in the current DOM will be replaced.
* Unmatched contents of the template will be ignored.

See [Client Library](https://github.com/rur/treetop-client) for more information.


## References
1. Go supports template inheritance through [nested template definitions](https://tip.golang.org/pkg/text/template/#hdr-Nested_template_definitions).
