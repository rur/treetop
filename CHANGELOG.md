## [0.2.0] - 2020-01-26

### Breaking API Changes

Taking Treetop from POC to Alpha gives me an opportunity to
execute on a wishlist of API changes.

- Rename `treetop.Renderer` type to `treetop.Page`
- Remove `treetop.NewView..` API method for creating base view with default template exec
- Rename `treetop.TreetopWriter` interface to `treetop.Writer` to conform to naming guidelines
- Remove `treetop.Test`, testing recipes and resources belong elsewhere
- Split `treetop.Writer` function into `treetop.NewPartialWriter` and `treetop.NewFragmentWriter` and remove the confusing `isPartial` flag
- Change `treetop.View` from an interface to a struct and expose internals to make debugging easier


#### Defining a page with views

Relatively minor change but makes more sense now I think.

```
page := treetop.NewPage(treetop.DefaultTemplateExec)
base := page.NewView(pageHandler, "base.html.tmpl")
nav := base.DefaultSubView("nav", navHandler)
content := base.DefaultSubView("content", contentHandler)
content2 := base.SubView("content", content2Handler)
```

## [0.1.0] - 2020-01-26

### Changed

- Added go.mod file
- added v0.1.0 tag as outlined in go blog [Migrating to Go Modules](https://blog.golang.org/migrating-to-go-modules)
