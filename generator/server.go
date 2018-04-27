package generator

import (
	"log"
	"text/template"
)

var (
	serverTemplate  *template.Template
	handlerTemplate *template.Template
)

type entry struct {
	Extends    string
	Handler    string
	Identifier string
	Type       string
	Value      string
}

type block struct {
	FieldName  string
	Identifier string
	Name       string
}

type handler struct {
	Blocks []block
	Info   string
	Name   string
}

type route struct {
	Path    string
	Handler string
}

type page struct {
	Identifier string
	Doc        string
	Entries    []entry
	Routes     []route
	Handlers   []handler
	Blocks     []block
}

func init() {
	var err error
	// TODO: inline templates
	serverTemplate, err = template.New("server").ParseFiles("generator/server.go.templ")
	if err != nil {
		log.Fatal(err)
	}
	handlerTemplate, err = template.New("handler").ParseFiles("generator/handler.go.templ")
	if err != nil {
		log.Fatal(err)
	}
}

func CreateSeverFiles(dir string, pages []PartialDef) ([]string, error) {
	var created []string
	return created, nil
}
