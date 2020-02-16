## [0.2.0] - 2020-01-26

### Breaking API Changes

Taking Treetop from POC to Alpha gives me an opportunity to
execute on a wishlist of API changes.

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
