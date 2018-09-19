// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots at
// 2018-09-18 17:31:41.481261 -0700 PDT m=+0.004435196
package writers

import (
	html "html/template"
	"log"
	text "text/template"
)

var (
	contextTemplate *text.Template
	handlerTemplate *text.Template
	pageGoTemplate *text.Template
	pageTemplate *html.Template
	partialTemplate *html.Template
	startTemplate *text.Template
	
)

func init() {
	var err error
	contextTemplate, err = text.New("templates/context.go.templ").Parse(contextTempl)
	if err != nil {
		log.Fatal(err)
	}
	handlerTemplate, err = text.New("templates/handler.go.templ").Parse(handlerTempl)
	if err != nil {
		log.Fatal(err)
	}
	pageGoTemplate, err = text.New("templates/page.go.templ").Parse(pageGoTempl)
	if err != nil {
		log.Fatal(err)
	}
	pageTemplate, err = html.New("templates/page.templ.html").Delims("[[", "]]").Parse(pageTempl)
	if err != nil {
		log.Fatal(err)
	}
	partialTemplate, err = html.New("templates/partial.templ.html").Delims("[[", "]]").Parse(partialTempl)
	if err != nil {
		log.Fatal(err)
	}
	startTemplate, err = text.New("templates/start.go.templ").Parse(startTempl)
	if err != nil {
		log.Fatal(err)
	}
	
}

// templates/context.go.templ
var contextTempl = `package pages

import (
	"net/http"

	"github.com/rur/treetop"
)

type ServerContext struct {
	Example bool
}

type RequestContext struct {
	Example bool
}

func (s *ServerContext) Bind(f func(*RequestContext, treetop.DataWriter, *http.Request)) treetop.HandlerFunc {
	return func(w treetop.DataWriter, req *http.Request) {
		// this will be called for every request to a bound handler
		cxt := RequestContext{s.Example}
		f(&cxt, w, req)
	}
}

type Mux interface {
	Handle(pattern string, handler http.Handler)
}
`

// templates/handler.go.templ
var handlerTempl = `package {{ .PageName }}

import (
	"net/http"

	"github.com/rur/treetop"
	"{{ .Namespace }}/treetop-workspace/pages"
)

{{ range $index, $handler := .Handlers }}
// {{ $handler.Info }}{{ if $handler.Doc }}
// {{ $handler.Doc }}
{{- end }}{{ if len $handler.Blocks}}
func {{ $handler.Identifier }}(cxt *pages.RequestContext, w treetop.DataWriter, req *http.Request) {
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
func {{ $handler.Identifier }}(cxt *RequestContext, w treetop.DataWriter, req *http.Request) {
	w.Data("{{ $handler.Info }} template data here!")
}{{ end }}
{{ end }}
`

// templates/page.go.templ
var pageGoTempl = `package {{ .Name }}

import (
	"github.com/rur/treetop"
	"{{ .Namespace }}/pages"
)

func Page(server *pages.ServerContext, m mux, renderer treetop.Renderer) {
	page := renderer.Page("{{ .Template }}", server.bind({{ .Handler }}))
	{{ range $index, $block := .Blocks -}}
		{{ $block.Identifier }} := page.Block("{{ $block.Name }}")
	{{ end }}
	{{ range $index, $entry := .Entries -}}
	{{ if eq $entry.Type "Block" -}}
	{{ $entry.Identifier }} := {{ $entry.Extends }}.Block("{{ $entry.Name }}")
	{{- else if eq $entry.Type "DefaultPartial" -}}
	{{ $entry.Identifier }} := {{ $entry.Extends }}.DefaultPartial(
		"{{ $entry.Template }}",
		server.bind({{ $entry.Handler }}),
	)
	{{- else if eq $entry.Type "Partial" -}}
	{{ $entry.Identifier }} := {{ $entry.Extends }}.Partial(
		"{{ $entry.Template }}",
		server.bind({{ $entry.Handler }}),
	)
	{{- else if eq $entry.Type "Fragment" -}}
	{{ $entry.Identifier }} := {{ $entry.Extends }}.Fragment(,
		"{{ $entry.Template }}",
		server.bind({{ $entry.Handler }}),
	)
{{- else if eq $entry.Type "Spacer" }}{{ else -}}
	nil // unknown entry type: {{ $entry.Type }}
	{{- end }}
	{{ end }}{{ range $index, $route := .Routes }}
	m.Handle("{{ $route.Path }}", {{ $route.Identifier }})  // {{ $route.Type }}
	{{- end }}
}
`

// templates/page.templ.html
var pageTempl = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Page [[ .Page.Name ]]</title>
    <style>

    .page-nav {
        float: right;
    }

    .nav li {
        display: inline-block;
        margin-right: 1rem;
        margin-bottom: 1rem;
        background-color: aqua;
        padding: 0.2rem 1rem;
    }

    .page-nav li {
        background-color: greenyellow;
    }

    .block-extended {
        background-color: rgba(0, 0, 0, 0.1);
        padding: 0.5rem;
    }

    .block-default {
        background-color: lightsalmon;
        padding: 0.5rem;
    }

    </style>
</head>
<body>
    <ul class="page-nav nav">[[ range $link := .SiteLinks ]]
        [[ if $link.URI ]]<li><a treetop href="[[ $link.URI ]]">[[ $link.Label ]]</a></li>[[ else ]]
        <li>[[ $link.Label ]]</li>[[ end ]][[ end ]]
    </ul>
    <h1>[[ .Page.Name ]] page</h1>[[ range $index, $block := .Page.Blocks ]]
    <h2>Block [[ $block.Name ]]</h2>
    <ul class="nav">[[ range $partial := $block.Partials ]][[ if $partial.Path ]][[ if $partial.Default ]]
        <li><a href="[[ $partial.Path ]]" treetop>[[ $partial.Name ]]*</a></li>
        [[- else -]]
        <li><a href="[[ $partial.Path ]]" treetop>[[ $partial.Name ]]</a></li>
        [[- end ]][[- end ]]
    [[ end ]]</ul>
    {{ block "[[ $block.Name ]]" .[[ $block.FieldName ]] }}
    <div id="[[ $block.Name  ]]" class="block-default">
        <p>default for block named [[ $block.Name ]]</p>
    </div>
    {{ end }}
[[ end ]]
<script src="/static/lib/treetop.js" async></script>
</body>
</html>`

// templates/partial.templ.html
var partialTempl = `{{ block "[[ .Extends ]]" . }}
<div id="[[ .Extends ]]" class="block-extended">
    <p>View named [[.Name]]</p>[[ range $index, $block := .Blocks ]]
<h3>Block [[ $block.Name ]]</h3>
<ul class="nav">[[ range $partial := $block.Partials ]][[ if $partial.Path ]][[ if $partial.Default ]]
    <li><a href="[[ $partial.Path ]]" treetop>[[ $partial.Name ]]*</a></li>[[ else ]]
    <li><a href="[[ $partial.Path ]]" treetop>[[ $partial.Name ]]</a></li>
    [[- end ]][[- end ]]
[[ end ]]</ul>
{{ block "[[ $block.Name ]]" .[[ $block.FieldName ]] }}
<div id="[[ $block.Name  ]]" class="block-default">
    <p>default for block named [[ $block.Name ]]</p>
</div>
{{ end }}
[[ end ]]
</div>
{{ end }}`

// templates/start.go.templ
var startTempl = `package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rur/treetop"

	"{{ .Namespace }}/pages"
	{{ range $index, $page := .Pages -}}
	"{{ $.Namespace }}/pages/{{ $page }}"
	{{ end }}
)

var (
	addr = ":8000"
)

func main() {
	m := http.NewServeMux()

	cxt := pages.ServerContext{Example: true}

	renderer := treetop.NewRenderer(treetop.DefaultTemplateExec){{ range $index, $page := .Pages }}
	{{ $page }}.Page(&cxt, m, renderer)
{{- end }}

	fmt.Printf("Starting treetop-workspace server at %s", addr)
	// Bind to a addr and pass our router in
	log.Fatal(http.ListenAndServe(addr, m))
}

`
