package generator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	serverTemplate  *template.Template
	handlerTemplate *template.Template
)

type entry struct {
	Extends    string
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
	Path       string
	Identifier string
}

type page struct {
	Identifier string
	Template   string
	Handler    string
	Doc        string
	Entries    []entry
	Routes     []route
	Blocks     []block
}

func init() {
	var err error
	// TODO: inline templates
	serverTemplate, err = template.New("server.go.templ").ParseFiles("generator/server.go.templ")
	if err != nil {
		log.Fatal(err)
	}
	handlerTemplate, err = template.New("handler.go.templ").ParseFiles("generator/handler.go.templ")
	if err != nil {
		log.Fatal(err)
	}
}

// create server files inside supplied dir and returns a list of all files created
func CreateSeverFiles(dir string, pageDefs []PartialDef) ([]string, error) {
	var created []string
	pages := make([]page, 0, len(pageDefs))
	var handlers []handler

	// all 'identifiers' must be unique,
	idents := newIdentifiers()

	for _, def := range pageDefs {
		newPage := page{
			Identifier: idents.new(def.Name, "Page"),
			Template:   fmt.Sprintf("%s.templ.html", namelike(def.Name)),
			Entries:    make([]entry, 0),
			Routes:     make([]route, 0),
			Blocks:     make([]block, 0, len(def.Blocks)),
		}

		for blockName, partials := range def.Blocks {
			newBlock := block{
				FieldName:  blockName,
				Identifier: idents.new(blockName, "Block"),
				Name:       blockName,
			}
			newPage.Blocks = append(newPage.Blocks, newBlock)
			for _, partial := range partials {
				entries, err := aggregateEntries(&idents, namelike(def.Name), newBlock.Identifier, partial)
				if err != nil {
					return created, err
				}
				newPage.Entries = append(newPage.Entries, entries...)
			}
		}
		pages = append(pages, newPage)
	}

	serverFile := "server.go"
	serverPath := filepath.Join(dir, serverFile)
	sf, err := os.Create(serverPath)
	if err != nil {
		return created, err
	}
	created = append(created, serverFile)
	defer sf.Close()
	err = serverTemplate.Execute(sf, struct {
		Pages []page
	}{
		Pages: pages,
	})
	if err != nil {
		return created, err
	}

	handlerFile := "handler.go"
	handlerPath := filepath.Join(dir, handlerFile)
	hf, err := os.Create(handlerPath)
	if err != nil {
		return created, err
	}
	created = append(created, handlerFile)
	defer hf.Close()
	err = handlerTemplate.Execute(hf, struct {
		Handlers []handler
	}{
		Handlers: handlers,
	})
	if err != nil {
		return created, err
	}

	return created, nil
}

func aggregateEntries(idents *uniqueIdentifiers, prefix, extends string, part PartialDef) ([]entry, error) {
	var prefixN string
	if prefix != "" {
		prefixN = strings.Join([]string{prefix, namelike(part.Name)}, "_")
	} else {
		prefixN = namelike(part.Name)
	}

	newEntry := entry{
		Extends:    extends,
		Identifier: idents.new(part.Name, "Partial"),
		Type:       "Partial",
		Value:      fmt.Sprintf("%s.templ.html", prefixN),
	}

	entries := []entry{newEntry}
	for blockName, partials := range part.Blocks {
		blockEntry := entry{
			Extends:    newEntry.Identifier,
			Identifier: idents.new(blockName, "Block"),
			Type:       "Block",
			Value:      blockName,
		}
		entries = append(
			entries,
			entry{
				Type: "Spacer",
			},
			blockEntry,
		)
		for _, partial := range partials {
			subEntries, err := aggregateEntries(idents, prefixN, blockEntry.Identifier, partial)
			if err != nil {
				return entries, err
			}
			entries = append(entries, subEntries...)
		}
		entries = append(entries, entry{
			Type: "Spacer",
		})
	}

	return entries, nil
}
