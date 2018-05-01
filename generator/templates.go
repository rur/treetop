package generator

// TODO: setup `go generate ...` to create this file

var handlerTempl = `package server

import (
	"net/http"

	"github.com/rur/treetop"
)

{{ range $index, $handler := .Handlers }}
{{ if $handler.Doc }}// {{ $handler.Doc }}
{{- end }}{{ if len $handler.Blocks}}
func {{ $handler.Identifier }}(cxt *Context, w treetop.DataWriter, req *http.Request) {
	{{ range $index, $block := .Blocks -}}
	{{ $block.Identifier }}, ok := w.PartialData("{{ $block.Name }}", req)
	if !ok {
		// default {{ $block.Name }} data
		{{ $block.Identifier }} = nil
	}
	{{ end }}
	w.Data(struct {
		HandlerInfo string{{ range $index, $block := .Blocks }}
		{{ $block.FieldName }} interface{}
		{{- end }}
	}{
		HandlerInfo: "{{ $handler.Info }}",{{ range $index, $block := .Blocks }}
		{{ $block.FieldName }}: {{ $block.Identifier }},
		{{- end }}
	})
}{{ else }}
func {{ $handler.Identifier }}(cxt *Context, w treetop.DataWriter, req *http.Request) {
	w.Data("{{ $handler.Info }} template data here!")
}{{ end }}
{{ end }}
`

var serverTempl = `package server

import (
	"net/http"

	"github.com/rur/treetop"
)

type Context struct {
	Example bool
}

func (c *Context) bind(f func(*Context, treetop.DataWriter, *http.Request)) treetop.HandlerFunc {
	return func(w treetop.DataWriter, req *http.Request) {
		f(c, w, req)
	}
}

type mux interface {
	Handle(pattern string, handler http.Handler)
}

func NewTreetopServer(m mux, cxt Context) {
	renderer := treetop.NewRenderer(treetop.DefaultTemplateExec){{ range $index, $page := .Pages }}
	{{ $page.Identifier }}Page(&cxt, m, renderer)
{{- end }}
}

{{ range $index, $page := .Pages -}}
func {{ $page.Identifier }}Page(context *Context, m mux, renderer treetop.Renderer) {
	{{ $page.Identifier }} := renderer.Page("{{ $page.Template.Path }}", context.bind({{ $page.Handler }}))
	{{ range $index, $block := $page.Blocks -}}
		{{ $block.Identifier }} := {{ $page.Identifier }}.Block("{{ $block.Name }}")
	{{ end }}
	{{ range $index, $entry := $page.Entries -}}
	{{ if eq $entry.Type "Block" -}}
	{{ $entry.Identifier }} := {{ $entry.Extends }}.Block("{{ $entry.Name }}")
	{{- else if eq $entry.Type "DefaultPartial" -}}
	{{ $entry.Identifier }} := {{ $entry.Extends }}.DefaultPartial("{{ $entry.Template }}", context.bind({{ $entry.Handler }}))
	{{- else if eq $entry.Type "Partial" -}}
	{{ $entry.Identifier }} := {{ $entry.Extends }}.Partial("{{ $entry.Template }}", context.bind({{ $entry.Handler }}))
	{{- else if eq $entry.Type "Fragment" -}}
	{{ $entry.Identifier }} := {{ $entry.Extends }}.Fragment("{{ $entry.Template }}", context.bind({{ $entry.Handler }}))
{{- else if eq $entry.Type "Spacer" }}{{ else -}}
	nil // unknown entry type: {{ $entry.Type }}
	{{- end }}
	{{ end }}{{ range $index, $route := .Routes }}
	m.Handle("{{ $route.Path }}", {{ $route.Identifier }})
	{{- end }}
}

{{ end }}
`

var pageTempl = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Page [[ .Name ]]</title>
</head>
<body>
<h1>Page [[ .Name ]]</h1>[[ range $index, $block := .Blocks ]]
    <h2>Block [[ $block.Name ]]</h2>
    <ul>[[ range $partial := $block.Partials ]][[ if $partial.Path ]][[ if $partial.Default ]]
        <li style="display: inline;"><a href="[[ $partial.Path ]]" treetop>[[ $partial.Name ]]*</a></li>[[ else ]]
        <li style="display: inline;"><a href="[[ $partial.Path ]]" treetop>[[ $partial.Name ]]</a></li>
        [[- end ]][[- end ]]
    [[ end ]]</ul>
    {{ block "[[ $block.Name ]]" .[[ $block.FieldName ]] }}
    <div id="[[ $block.Name  ]]" style="background-color: lightsalmon;">
        <p>default for block named [[ $block.Name ]]</p>
    </div>
    {{ end }}
[[ end ]]
<script src="/static/lib/treetop.js" async></script>
</body>
</html>
`

var partialTempl = `{{ block "[[ .Extends ]]" . }}
<div id="[[ .Extends ]]" style="background-color: rgba(0, 0, 0, 0.1)">
    <p>View named [[.Name]]</p>[[ range $index, $block := .Blocks ]]
<ul>[[ range $partial := $block.Partials ]][[ if $partial.Path ]][[ if $partial.Default ]]
    <li style="display: inline;"><a href="[[ $partial.Path ]]" treetop>[[ $partial.Name ]]*</a></li>[[ else ]]
    <li style="display: inline;"><a href="[[ $partial.Path ]]" treetop>[[ $partial.Name ]]</a></li>
    [[- end ]][[- end ]]
[[ end ]]</ul>
{{ block "[[ $block.Name ]]" .[[ $block.FieldName ]] }}
<div id="[[ $block.Name  ]]" style="background-color: lightsalmon;">
    <p>default for block named [[ $block.Name ]]</p>
</div>
{{ end }}
[[ end ]]
</div>
{{ end }}
`
