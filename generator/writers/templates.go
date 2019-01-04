// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots at
// 2019-01-03 16:45:53.825545 -0800 PST m=+0.004366956
package writers

import (
	html "html/template"
	"log"
	text "text/template"
)

var (
	handlerTemplate *text.Template
	indexTemplate *html.Template
	partialTemplate *html.Template
	resourcesTemplate *text.Template
	routesTemplate *text.Template
	startTemplate *text.Template
	
)

func init() {
	var err error
	handlerTemplate, err = text.New("templates/handler.go.templ").Parse(handlerTempl)
	if err != nil {
		log.Fatal(err)
	}
	indexTemplate, err = html.New("templates/index.templ.html").Delims("[[", "]]").Parse(indexTempl)
	if err != nil {
		log.Fatal(err)
	}
	partialTemplate, err = html.New("templates/partial.templ.html").Delims("[[", "]]").Parse(partialTempl)
	if err != nil {
		log.Fatal(err)
	}
	resourcesTemplate, err = text.New("templates/resources.go.templ").Parse(resourcesTempl)
	if err != nil {
		log.Fatal(err)
	}
	routesTemplate, err = text.New("templates/routes.go.templ").Parse(routesTempl)
	if err != nil {
		log.Fatal(err)
	}
	startTemplate, err = text.New("templates/start.go.templ").Parse(startTempl)
	if err != nil {
		log.Fatal(err)
	}
	
}

// templates/handler.go.templ
var handlerTempl = `package {{ .PageName }}

import (
	"net/http"

	"github.com/rur/treetop"
	"{{ .Namespace }}/page"
)

{{ range $index, $handler := .Handlers }}
// {{ $handler.Info }} {{ $handler.Type }}{{ if $handler.Extends }}
// Extends: {{ $handler.Extends }}{{ end }}{{ if $handler.Doc }}
// Doc: {{ $handler.Doc }}
{{- end }}{{ if len $handler.Blocks}}
func {{ $handler.Identifier }}(rsc page.Resources, rsp treetop.Response, req *http.Request) interface{} {
	return  struct {
		HandlerInfo string{{ range $index, $block := .Blocks }}
		{{ $block.FieldName }} interface{}
		{{- end }}
	}{
		HandlerInfo: "{{ $handler.Info }}",{{ range $index, $block := .Blocks }}
		{{ $block.FieldName }}: rsp.HandlePartial("{{ $block.Name }}", req),
		{{- end }}
	}
}{{ else }}
func {{ $handler.Identifier }}(rsc page.Resources, rsp treetop.Response, req *http.Request) interface{} {
	return  "{{ $handler.Info }} template data here!"
}{{ end }}
{{ end }}
`

// templates/index.templ.html
var indexTempl = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Page [[ .Title ]]</title>
    <style>

    body {
        padding: 1rem;
    }

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

    .fragment {
        background-color: darkcyan;
        color: white;
    }
    </style>
</head>
<body>
    <ul class="page-nav nav">[[ range $link := .SiteLinks ]]
        [[ if $link.Active ]]<li><a href="[[ $link.URI ]]">[[ $link.Label ]]</a></li>[[ else ]]
        <li>[[ $link.Label ]]</li>[[ end ]][[ end ]]
    </ul>
    <h1>[[ .Title ]] page</h1>[[ range $index, $block := .Blocks ]]
    <h2>Block [[ $block.Name ]]</h2>
    <ul class="nav">[[ range $partial := $block.Partials ]][[ if $partial.Path ]][[ if $partial.Fragment ]]
        <li><button treetop-link="[[ $partial.Path ]]">[[ $partial.Name ]][[ if $partial.Default ]]*[[end]]</button></li>[[ else ]]
        <li><a href="[[ $partial.Path ]]" treetop>[[ $partial.Name ]][[ if $partial.Default ]]*[[end]]</a></li>
        [[- end ]][[- end ]]
    [[ end ]]</ul>
    {{ block "[[ $block.Name ]]" .[[ $block.FieldName ]] }}
    <div id="[[ $block.Name  ]]" class="block-default">
        <p>default for block named [[ $block.Name ]]</p>
    </div>
    {{ end }}
[[ end ]]

    <script> window.TREETOP_CONFIG = {} </script>
    <script async src="https://rawgit.com/rur/treetop-client/master/treetop.js"></script>
</body>
</html>`

// templates/partial.templ.html
var partialTempl = `{{ block "[[ .Extends ]]" . }}
<div id="[[ .Extends ]]" class="block-extended[[ if .Fragment ]] fragment[[ end ]]">
    [[ if .Fragment ]]<p>Fragment view named [[.Name]]</p>
    [[- else ]]<p>Partial view named [[.Name]]</p>[[ end ]]
    [[- range $index, $block := .Blocks ]]
    <h3>Block [[ $block.Name ]]</h3>
    <ul class="nav">[[ range $partial := $block.Partials ]][[ if $partial.Path ]][[ if $partial.Fragment ]]
        <li><button treetop-link="[[ $partial.Path ]]">[[ $partial.Name ]][[ if $partial.Default ]]*[[end]]</button></li>[[ else ]]
        <li><a href="[[ $partial.Path ]]" treetop>[[ $partial.Name ]][[ if $partial.Default ]]*[[end]]</a></li>
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

// templates/resources.go.templ
var resourcesTempl = `package page

import (
	"net/http"
	"sync"

	"github.com/rur/treetop"
)

// Services, session info and configuration shared between handlers servicing the same request.
// eg: query facade, user info, roles, etc..
//
// NB: It is very important that data handlers do not communicate directly,
//     consider before adding a field to this type.
type Resources struct {
	Example bool
}

type Mux interface {
	Handle(pattern string, handler http.Handler)
}

type ResourcesHandler func(Resources, treetop.Response, *http.Request) interface{}

type Server interface {
	Bind(ResourcesHandler) treetop.HandlerFunc
}

func NewServer() Server {
	return &server{
		responses: make(map[uint32]*Resources),
	}
}

// Server configuration state
// eg: resource pools, auth provider, secrets, etc..
type server struct {
	sync.RWMutex
	responses map[uint32]*Resources
}

// attempt to load Resources from the server cache
func (s *server) get(respId uint32) *Resources {
	s.RLock()
	defer s.RUnlock()
	if rsc, ok := s.responses[respId]; ok {
		return rsc
	} else {
		return nil
	}
}

func (s *server) set(respId uint32, rsc *Resources) {
	s.Lock()
	defer s.Unlock()
	s.responses[respId] = rsc
}

// remove Resources from the cache for a given treetop response ID, delete is idempotent
func (s *server) delete(respId uint32) {
	s.Lock()
	defer s.Unlock()
	delete(s.responses, respId)
}

// wrap ;
func (s *server) Bind(f ResourcesHandler) treetop.HandlerFunc {
	return func(rsp treetop.Response, req *http.Request) interface{} {
		// Here the Treetop response ID is being used to permit resources to be shared
		// between data handlers, within the scope of a request.
		respId := rsp.ResponseID()
		rsc := s.get(respId)

		if rsc == nil {
			// potentially blocking resource initialization here
			rsc = &Resources{true}
			s.set(respId, rsc)
			go func() {
				<-rsp.Context().Done()
				// assume that the request lifecycle is finished, just free up resources
				s.delete(respId)
			}()
		}
		return f(*rsc, rsp, req)
	}
}
`

// templates/routes.go.templ
var routesTempl = `package {{ .Name }}

import (
	"github.com/rur/treetop"
	"{{ .Namespace }}/page"
)

func Routes(server page.Server, m page.Mux, renderer *treetop.Renderer) {
	pageView := renderer.NewView(
		"{{ .Template }}",
		{{ if .OverrideHandler -}}
		{{ .Handler }},
		{{- else -}}
		server.Bind({{ .Handler }}),
		{{- end }}
	)
	{{ range $index, $entry := .Entries -}}
	{{ if eq $entry.Type "DefaultSubView" -}}
	{{ $entry.Identifier }} {{ $entry.Extends }}.DefaultSubView(
		"{{ $entry.Block }}",
		"{{ $entry.Template }}",
		{{ if $entry.OverrideHandler -}}
		{{ $entry.Handler }},
		{{- else -}}
		server.Bind({{ $entry.Handler }}),
		{{- end }}
	)
	{{- else if eq $entry.Type "SubView" -}}
	{{ $entry.Identifier }} {{ $entry.Extends }}.SubView(
		"{{ $entry.Block }}",
		"{{ $entry.Template }}",
		{{ if $entry.OverrideHandler -}}
		{{ $entry.Handler }},
		{{- else -}}
		server.Bind({{ $entry.Handler }}),
		{{- end }}
	)
	{{- else if eq $entry.Type "Spacer" }}
	// {{ $entry.Name }}
	{{- else -}}
	nil // unknown entry type: {{ $entry.Type }}
	{{- end }}
	{{ end }}{{ range $index, $route := .Routes }}
	m.Handle("{{ $route.Path }}", treetop.ViewHandler({{ $route.Reference }})
	{{- if eq $route.Type "Page" }}.PageOnly()
	{{- else if eq $route.Type "Fragment" }}.FragmentOnly()
	{{- end }})
	{{- end }}
}
`

// templates/start.go.templ
var startTempl = `package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rur/treetop"

	"{{ .Namespace }}/page"
	{{ range $index, $page := .Pages -}}
	"{{ $.Namespace }}/page/{{ $page }}"
	{{ end }}
)

var (
	addr = ":8000"
)

func main() {
	m := http.NewServeMux()

	server := page.NewServer()

	renderer := treetop.NewRenderer(treetop.DefaultTemplateExec){{ range $index, $page := .Pages }}
	{{ $page }}.Routes(server, m, renderer)
{{- end }}

	fmt.Printf("Starting {{ $.Namespace }} server at %s", addr)
	// Bind to an addr and pass our router in
	log.Fatal(http.ListenAndServe(addr, m))
}

`
