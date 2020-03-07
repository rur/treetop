[![Build Status](https://travis-ci.org/rur/treetop.svg?branch=v0.3.0)](https://travis-ci.org/rur/treetop)

# Treetop

## Hierarchical Web Handlers

Mix and match templates and functions to create HTML endpoints. Treetop makes it more convenient to take advantage of nested templates<sup>1</sup> in Go web applications. Thus you can assemble many endpoints, reusing common structural elements.

Example 1, a hierarchy of views are used to bind the routes `"/content_a"` and `"/content_b"` to composite handlers.

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

Example z `"base.html.tmpl"`

	...
	<div id="base">
		{{ block "nav" .Nav }}  <div id="nav">default nav</div> {{ end }}

		{{ template "sidebar" .SideBar }}

		{{ template "content" .Content }}
	</div>
	...

There can be multiple levels of nesting, as needed.

### No Third-Party Dependencies

The treetop package wraps features of the Go standard library only, mostly within "net/http" and "html/template".


##  HTML Template Protocol

Hot-swap sections of a page. Views are self-contained, they can be rendered in isolation. Treetop handlers are capable of generating template fragments which can be 'applied' to the loaded document. The following illustrates a template request.


    > GET / HTTP/1.1
      Accept: application/x.treetop-html-template+xml

    < HTTP/1.1 200 OK
      Content-Type: application/x.treetop-html-template+xml
      Vary: Accept

      <template>
          <div id="content">...</div>
          <div id="nav">...</div>
      </template>

There are many ways this can be used to improve user experience <sup>[docs needed]</sup>. A [Treetop Client Library](https://github.com/rur/treetop-client) is available to help manage these requests using XHR. The JS client applies template fragments to the DOM with a very simple find and replace mechanism.

## Example

A very basic example can be run from this repo <sup>[needs improvement]</sup>.

    $ go run example/

Tip. Activate your network tab to observe what's going on.

### Other Examples:

- [Todo \*Without\* MVC](https://github.com/rur/todowithoutmvc) - Treetop implementation of [TodoMVC](http://todomvc.com) app using the template protocol.

## Template Executor

The 'Executor' is responsible for loading and configuring the templates. It creates a [HTTP Handler](https://golang.org/pkg/net/http/#Handler) which will manage the plumbing between loading data and executing templates. You can implement your own template loader <sup>[docs needed]</sup>, but the following are provided:
- `FileExecutor` - load template files using os.Open
- `FileSytemExecutor` - loads templates from a supplied http.FileSystem instance
- `StringExecutor` - treat the view template property as an inline template string
- `KeyedStringExecutor` - treat the view template property is key into a template map
- `DeveloperExecutor` - force per request template parsing

## View Handlers Function

View handler functions load data for a corresponding Go template. Just as nested templates are embedded in their parent, nested handler data is embedded in the data of it's parent. Example of a parent handler loading data from a child handler,

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


## Response Writer

The Treetop Go library provides utilities for writing ad-hoc template responses when needed. PartialWriter and FragmentWriter wrap the `http.ResponseWriter`,

    func myHandler(w http.ResponseWriter, req *http.Request) {
        // check for treetop request and construct a writer
        if tw, ok := treetop.NewFragmentWriter(w, req); ok {
            fmt.Fprintf(tw, `<h3 id="greeting">Hello, %s</h3>`, "Treetop")
        }
    }


## Client Library

The client library is used to send template requests from the browser. Response template fragments are handled mechanically so there are no callbacks or other IO boilerplate involved.

See [Client Library](https://github.com/rur/treetop-client) for more information.


## References
1. Go supports template inheritance through [nested template definitions](https://tip.golang.org/pkg/text/template/#hdr-Nested_template_definitions).
