## [0.2.0] - 2020-01-26

### Changed

- Renamed `treetop.Renderer` struct to `treetop.Page`
- Remove `treetop.NewView..` API method for creating base view with default template exec, doesn't add anything to the API

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
