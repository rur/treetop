## [0.4.0] - 2021-08-08

Finalize release candidate for v0.4 and embedded client code

### Changes

- Upgrade embedded treetop-client code following v0.10.0 release

## [0.4.0-rc.2] - 2021-06-24

Quick fix for an issue in the previous release affecting the FS executor.

### Bugfix

- `FileSystemExecutor` was memoizing template contents
  - This is of little benefit since it only helps at startup time
  - The template contents stay in memory for the lifetime of the app

### Housekeeping

- Add test coverage for DeveloperExecutor forced reloading

## [0.4.0-rc.0] - 2021-06-20

A refactor of executors, with stricter template validation which may cause breakage.

### Added

- New `treetop.ServeClientLibrary` handler serving an embedded copy of the _treetop-client_ JS, for convenience sake
- Add `view.HasSubView` method to assert the existence of a sub view name without specifying a template.

        myPage := treetop.NewView("base.html.tmpl", BaseHandler)
        myPage.HasSubView("sidebar") // no template or handler specified

### Breaking Changes

- A template will be rejected **with an error** if it does not contain a template or block declaration
  for all defined sub-view names.

        tErr.Error()
        => template example.html.tmpl: missing template declaration(s) for sub view blocks: "my-block"

  - The above can be corrected by adding `{{ template "my-block" .MyBlock }}` to _example.html.tmpl_.

### Housekeeping

- Major refactor of Executors to remove duplicate code. The behavior of all executors should remain the same!
- Replace deprecated use of `HeaderMap` in favor of `.Header()`

## [0.3.2] - 2021-04-10

### Housekeeping

- Test with latest stable Golang version 1.16
- Move Treetop Demo code into a separate repo https://github.com/rur/treetop-demo

## [0.3.1] - 2020-04-07

### New Features

- `FileExecutor` and `FileSystemExecutor` now have an optional `KeyedString` property to mix files with hard coded templates
- Added `Template` interface so that `TemplateHandler` no longer depends directly on html/template

### Bugfix

- Template cache not used/not working when constructing template in file executors
- Files not being closed properly in FS executor

## [0.3.0] - 2020-03-10

Protocol and API overhaul, improve docs and examples, transitioning from prototype
to a usable library. Improve test coverage >= 90% range.

### Protocol Change

Treetop is a 'HTML template' protocol with one content-type value.

    application/x.treetop-html-template+xml

Use of the terms 'fragment' and 'partial' have been done away with. The following
headers can be included (optional) in the response to control navigation history.

    X-Page-URL: /some/path
    X-Response-History: replace

All HTML content in the response body should be wrapped in a single HTMLTemplateElement (this is implicit if not present).

    <template>
        <p id="first">this is the first fragment</p>
        <p id="second">this is the second fragment</p>
    </template>

This will make the intention of the protocol more obvious for developers already
familiar with the use of HTML5 templates and avoid some issues with non-rooted
markup in a response body.

### Change Redirect Header

Treetop redirects now use the Location header value as the destination URL. The only difference
from a normal HTTP redirect is the status code of 200. The `X-Treetop-Redirect`
header will have a value of "SeeOther" to signal to the XHR client that a new location
should be forced.

### Views and Executors

Implementation of the `View` type has been simplified.

The `ViewExecutor` interface was created to encapsulate responsibility for converting a view
to a HTTP handler. The `TemplateExec` function interface is gone. The following template loaders
have been implemented with the treetop package:

1. `FileExecutor` - Loads template files from the files system, similar to template.ParseFiles
2. `FileSystemExecutor` - resolves templates through a http.FileSystem interface
3. `StringExecutor` - treats view.Template string as a literal template string
4. `KeyedStringExecutor` - treats view.Template as a key to a template map
5. `DeveloperExecutor` - Wrap any other executor and the templates will be reloaded for every request

Example usage

    base := NewView("base.html", MyHandler)
    content := base.NewSubView(
        "content", "content_a.html", MyContentAHandler)

    exec := treetop.NewKeyedStringExecutor(map[string]string{
        "base.html": "Base here",
        "content_a.html": "Content A here",
    })
    mux.Handle("/some/path", exec.NewViewHandler(content))

### View Debugging

`SprintViewTree` is a to-string function that is helpful for showing how a given view
hierarchy is configured.

Example output

    - View("base.html", ...Constant.func1)
      |- A: SubView("A", "A.html", ...Constant.func1)
      |  |- A1: SubView("A1", "A1.html", ...Constant.func1)
      |  '- A2: SubView("A2", "A2.html", ...Constant.func1)
      |
      '- B: SubView("B", "B.html", ...Constant.func1)
          |- B1: SubView("B1", "B1.html", ...Constant.func1)
          '- B2: SubView("B2", "B2.html", ...Constant.func1)

#### DeveloperExecutor Error Page

During development you can wrap your server template executor in a `DeveloperExecutor`
It will reload/re-parse templates for every request. Any template errors will be
rendered to the client in a formatted error page.

    exec := treetop.FileExecutor{}
    if devMode {
        exec = treetop.DeveloperExecutor{exec}
    }
    mux.Handle(....)

### View Helpers

Removed `SeeOtherPage` function, the `Redirect` helper is the only redirect function.

Renamed `IsTreetopRequest` predicate to a `IsTemplateRequest`, this is more consistent with
changes in protocol terminology.

### Fixes

- The `Content-Length` header is now set on all template responses.
- Templates are no longer re-parsed for every request

### Examples

- Add bootstrap CDN link for nicer styling and to help with demos
- Re-organize how demos are loaded to allow for more to be added
- Create 'Writer' greeter example for comparison

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
