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
	Blocks     []*block
	Info       string
	Doc        string
	Identifier string
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
	Entries    []*entry
	Routes     []*route
	Blocks     []*block
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
	var handlers []*handler

	// all 'identifiers' must be unique,
	idents := newIdentifiers()

	for _, def := range pageDefs {
		newPage := page{
			Identifier: idents.new(def.Name, "Page"),
			Template:   fmt.Sprintf("%s.templ.html", namelike(def.Name)),
			Entries:    make([]*entry, 0),
			Blocks:     make([]*block, 0, len(def.Blocks)),
		}
		pageHandler := handler{
			Info:       def.Name,
			Doc:        def.Doc,
			Identifier: newPage.Identifier,
		}

		handlers = append(handlers, &pageHandler)

		if def.Path != "" {
			newPage.Routes = []*route{&route{
				Identifier: newPage.Identifier,
				Path:       def.Path,
			}}
		}

		for blockName, partials := range def.Blocks {
			pageHandler.Blocks = append(pageHandler.Blocks, &block{
				Identifier: validIdentifier(blockName),
				Name:       blockName,
				FieldName:  validPublicIdentifier(blockName),
			})
			newBlock := block{
				FieldName:  blockName,
				Identifier: idents.new(blockName, "Block"),
				Name:       blockName,
			}
			newPage.Blocks = append(newPage.Blocks, &newBlock)
			for _, partial := range partials {
				entries, routes, partHandlers, err := aggregateEntries(&idents, namelike(def.Name), newBlock.Identifier, partial)
				if err != nil {
					return created, err
				}
				newPage.Entries = append(newPage.Entries, entries...)
				newPage.Routes = append(newPage.Routes, routes...)
				handlers = append(handlers, partHandlers...)
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

	handlerFile := "handlers.go"
	handlerPath := filepath.Join(dir, handlerFile)
	hf, err := os.Create(handlerPath)
	if err != nil {
		return created, err
	}
	created = append(created, handlerFile)
	defer hf.Close()
	err = handlerTemplate.Execute(hf, struct {
		Handlers []*handler
	}{
		Handlers: handlers,
	})
	if err != nil {
		return created, err
	}

	return created, nil
}

func aggregateEntries(idents *uniqueIdentifiers, prefix, extends string, part PartialDef) ([]*entry, []*route, []*handler, error) {
	var prefixN string
	if prefix != "" {
		prefixN = strings.Join([]string{prefix, namelike(part.Name)}, "_")
	} else {
		prefixN = namelike(part.Name)
	}

	var entryType string
	if part.Fragment {
		entryType = "Fragment"
	} else if part.Default {
		entryType = "DefaultPartial"
	} else {
		entryType = "Partial"
	}

	newEntry := entry{
		Extends:    extends,
		Identifier: idents.new(part.Name, entryType),
		Type:       entryType,
		Value:      fmt.Sprintf("%s.templ.html", prefixN),
	}

	var routes []*route
	entries := []*entry{&newEntry}

	if part.Path != "" {
		routes = []*route{&route{
			Identifier: newEntry.Identifier,
			Path:       part.Path,
		}}
	}

	partHandler := handler{
		Info:       part.Name,
		Doc:        part.Doc,
		Identifier: newEntry.Identifier,
	}
	handlers := []*handler{&partHandler}

	for blockName, partials := range part.Blocks {
		partHandler.Blocks = append(partHandler.Blocks, &block{
			Identifier: validIdentifier(blockName),
			Name:       blockName,
			FieldName:  validPublicIdentifier(blockName),
		})
		blockEntry := entry{
			Extends:    newEntry.Identifier,
			Identifier: idents.new(blockName, "Block"),
			Type:       "Block",
			Value:      blockName,
		}
		entries = append(
			entries,
			&entry{
				Type: "Spacer",
			},
			&blockEntry,
		)
		for _, partial := range partials {
			subEntries, subRoutes, subHandlers, err := aggregateEntries(idents, prefixN, blockEntry.Identifier, partial)
			if err != nil {
				return entries, routes, handlers, err
			}
			entries = append(entries, subEntries...)
			routes = append(routes, subRoutes...)
			handlers = append(handlers, subHandlers...)
		}
		entries = append(entries, &entry{
			Type: "Spacer",
		})
	}

	return entries, routes, handlers, nil
}
