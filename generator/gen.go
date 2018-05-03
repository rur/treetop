// The following directive is necessary to make the package coherent:

// +build ignore

// This program generates templates.go. It can be invoked by running
// go generate
package main

import (
	"io/ioutil"
	"log"
	"os"
	"text/template"
	"time"
)

type genTemplate struct {
	Path, Ident, Content string
}

func main() {
	files := map[string]string{
		"handler.go.templ":   "handlerTempl",
		"page.templ.html":    "pageTempl",
		"partial.templ.html": "partialTempl",
		"server.go.templ":    "serverTempl",
	}
	content := make([]genTemplate, 0, 4)

	for _, file := range []string{"handler.go.templ", "server.go.templ", "page.templ.html", "partial.templ.html"} {
		bites, err := ioutil.ReadFile(file)
		if err != nil {
			die(err)
		}
		ident := files[file]
		content = append(content, genTemplate{
			Path:    file,
			Ident:   ident,
			Content: string(bites),
		})
	}

	f, err := os.Create("templates.go")
	die(err)
	defer f.Close()

	packageTemplate.Execute(f, struct {
		Timestamp time.Time
		Templates []genTemplate
	}{
		Timestamp: time.Now(),
		Templates: content,
	})
}

func die(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var packageTemplate = template.Must(template.New("").Parse(`// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots at
// {{ .Timestamp }}
package generator

{{- range $index, $template := .Templates }}

// {{ $template.Path }}
var {{ $template.Ident }} = ` + "`{{ $template.Content }}`" + `


{{- end }}
`))