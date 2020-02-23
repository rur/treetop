## [0.3.0] - WIP

Protocol and API overhaul, transitioning from prototype to alpha.

### Protocol Change

Treetop is a 'HTML template' protocol with exactly one content-type value. 

    application/x.treetop-html-template+xml

Use of the terms 'fragment' and 'partial' have been done away with. When a response corresponds of a part of a full page, the following header will be included in the
response.

    X-Page-URL: /some/path

All HTML content in the response body should be wrapped in a single HTMLTemplateElement (this is implicit if not present).

    <template>
        <p id="first">this is the first fragment</p>
        <p id="second">this is the second fragment</p>
    </template>

This will make the intention of the protocol more obvious for developers already 
familiar with the use of HTML5 templates and avoid some issues with non-rooted
markup in a response body.

### Change Redirect Header

Treetop redirects use the Location header for the destination URL. The only difference
from a normal HTTP redirect is the status code of 200. The `X-Treetop-Redirect` 
header will have a value of "SeeOther" to signal to the XHR client that a new location
should be forced.

### Views and Executors

Implementation of the `View` type has been greatly simplified.

The `Executor` interface was created to encapsulate responsibility for converting a view
to a HTTP handler. Custom view executor implementations have been made easier to create.
The following implementations are supplied with the treetop package:

1. `DefaultExecutor` - Loads template files from the files system, wrapper for template.ParseFiles
2. `FileSystemExecutor` - resolves templates through a http.FileSystem interface 
3. `StringTemplateExecutor` - treats view.Template string as a literal template string
4. `DevelopmentFSExecutor` - FileSystemExecutor without caching the parsed templates

Example 

    base := NewView("base.html.tmpl", MyHandler)
    content := base.NewSubView(
        "content", "content_a.html.tmpl", MyContentAHandler)

    exec := treetop.DefaultExecutor{}
    mux.Handle("/some/path", exec.NewViewHandler(content))


## [0.2.0] - 2020-02-16

Add clarifications to the prototype library API and improve code docs.

### Bugfix

- Writer implementation was not including query in response URL header

### Breaking API Changes

- Remove `treetop.Renderer` type, unnecessary since it only wraps `TemplateExec`
- Add TemplateExec parameter to `treetop.NewView..` API method for creating base views
- Rename `treetop.TreetopWriter` interface to `treetop.Writer` to conform to naming guidelines
- Remove `treetop.Test`, testing recipes and resources belong elsewhere
- Split `treetop.Writer` function into `treetop.NewPartialWriter` and `treetop.NewFragmentWriter` and remove the confusing `isPartial` flag
- Change `treetop.View` from an interface to a struct and expose internals to make debugging easier
- Rename `HandlePartial` method to `HandleSubView` in `treetop.Response` interface to be consistent with the view builder
- Rename `Done` method to `Finished` in `treetop.Response` interface
- Rename `HandlerFunc` signature type to `ViewHandlerFunc` to make a clearer association with the view builder 


#### Defining a page with views

Relatively minor change but makes more sense now I think.

```
base := treetop.NewView(
    treetop.DefaultTemplateExec,
    pageHandler,
    "base.html.tmpl",
)
nav := base.DefaultSubView("nav", navHandler)
content := base.DefaultSubView("content", contentHandler)
content2 := base.SubView("content", content2Handler)
```

## [0.1.0] - 2020-01-26

### Changed

- Added go.mod file
- added v0.1.0 tag as outlined in go blog [Migrating to Go Modules](https://blog.golang.org/migrating-to-go-modules)
