package generator

import (
	"fmt"
	html "html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	text "text/template"
)

var (
	serverTemplate  *text.Template
	handlerTemplate *text.Template
	pageTemplate    *html.Template
	partialTemplate *html.Template
)

type entry struct {
	Extends    string
	Identifier string
	Handler    string
	Type       string
	Value      string
}

type blockPartial struct {
	Identifier string
	Path       string
	Name       string
	Fragment   bool
	Default    bool
	Blocks     []*block
}

type block struct {
	FieldName  string
	Identifier string
	Name       string
	Partials   []*blockPartial
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
	Name       string
	Entries    []*entry
	Routes     []*route
	Blocks     []*block
	Handlers   []*handler
}

type template struct {
	Path string
	Page page
}

func init() {
	var err error
	// TODO: inline templates
	serverTemplate, err = text.New("server.go.templ").Parse(serverTempl)
	if err != nil {
		log.Fatal(err)
	}
	handlerTemplate, err = text.New("handler.go.templ").Parse(handlerTempl)
	if err != nil {
		log.Fatal(err)
	}
	pageTemplate, err = html.New("page.templ.html").Delims("[[", "]]").Parse(pageTempl)
	if err != nil {
		log.Fatal(err)
	}
	partialTemplate, err = html.New("partial.go.templ").Delims("[[", "]]").Parse(partialTempl)
	if err != nil {
		log.Fatal(err)
	}
}

// create server files inside supplied dir and returns a list of all files created
func CreateSeverFiles(dir string, pageDefs []PartialDef) ([]string, error) {
	var created []string
	site := make([]page, 0, len(pageDefs))

	// all 'identifiers' must be unique,
	idents := newIdentifiers()

	for _, def := range pageDefs {
		newPage, pageHandler := createPage(&idents, def)
		for blockName, partials := range def.Blocks {
			newBlock := block{
				FieldName:  validPublicIdentifier(blockName),
				Identifier: idents.new(blockName, "Block"),
				Name:       blockName,
			}
			if pageHandler != nil {
				pageHandler.Blocks = append(pageHandler.Blocks, &newBlock)
			}
			newPage.Blocks = append(newPage.Blocks, &newBlock)
			for _, partial := range partials {
				entries, routes, handlers, err := createEntries(&idents, lowercaseName(def.Name), newBlock.Identifier, partial)
				if err != nil {
					return created, err
				}
				newPage.Entries = append(newPage.Entries, entries...)
				newPage.Routes = append(newPage.Routes, routes...)
				newPage.Handlers = append(newPage.Handlers, handlers...)
			}
		}
		site = append(site, newPage)
	}

	serverFile := filepath.Join("server", "server.go")
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
		Pages: site,
	})
	if err != nil {
		return created, err
	}

	for _, sitePage := range site {
		handlerFile := filepath.Join("server", sitePage.Identifier+"_handlers.go")
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
			Handlers: sitePage.Handlers,
		})
		if err != nil {
			return created, err
		}
		// page template
		templateFile := filepath.Join("templates", sitePage.Identifier+".templ.html")
		path := filepath.Join(dir, templateFile)
		tf, err := os.Create(path)
		if err != nil {
			return created, err
		}
		created = append(created, templateFile)
		defer tf.Close()
		err = pageTemplate.Execute(tf, struct {
			Page page
		}{
			Page: sitePage,
		})
		if err != nil {
			return created, err
		}
	}

	return created, nil
}

func createPage(idents *uniqueIdentifiers, def PartialDef) (page, *handler) {
	var pageHandler *handler
	newPage := page{
		Name:       def.Name,
		Identifier: idents.new(def.Name, "Page"),
		Template:   fmt.Sprintf("%s.templ.html", lowercaseName(def.Name)),
		Entries:    make([]*entry, 0),
		Blocks:     make([]*block, 0, len(def.Blocks)),
	}
	if def.Handler == "" {
		// no specific handler name appears in the definition, create a unique one
		pageHandler = &handler{
			Info:       def.Name,
			Doc:        def.Doc,
			Identifier: idents.new(newPage.Identifier+"Handler", "Func"),
		}
	} else if !idents.exists(def.Handler) {
		// a handler name was specified, create a new definition if it does not already exist
		pageHandler = &handler{
			Info:       def.Name,
			Doc:        def.Doc,
			Identifier: def.Handler,
		}
	} else {
		newPage.Handler = def.Handler
	}
	if pageHandler != nil {
		newPage.Handler = pageHandler.Identifier
		newPage.Handlers = append(newPage.Handlers, pageHandler)
	}
	if def.Path != "" {
		newPage.Routes = []*route{&route{
			Identifier: newPage.Identifier,
			Path:       def.Path,
		}}
	}
	return newPage, pageHandler
}

func createEntries(idents *uniqueIdentifiers, prefix, extends string, def PartialDef) ([]*entry, []*route, []*handler, error) {
	var prefixN string
	if prefix != "" {
		prefixN = strings.Join([]string{prefix, lowercaseName(def.Name)}, "_")
	} else {
		prefixN = lowercaseName(def.Name)
	}

	var entryType string
	if def.Fragment {
		entryType = "Fragment"
	} else if def.Default {
		entryType = "DefaultPartial"
	} else {
		entryType = "Partial"
	}

	newEntry := entry{
		Extends:    extends,
		Identifier: idents.new(def.Name, entryType),
		Type:       entryType,
		Value:      fmt.Sprintf("%s.templ.html", prefixN),
	}

	var routes []*route
	entries := []*entry{&newEntry}

	if def.Path != "" {
		routes = []*route{&route{
			Identifier: newEntry.Identifier,
			Path:       def.Path,
		}}
	}

	var partHandler *handler
	var handlers []*handler
	if def.Handler == "" {
		// no specific handler name appears in the definition, create a unique one
		partHandler = &handler{
			Info:       def.Name,
			Doc:        def.Doc,
			Identifier: idents.new(newEntry.Identifier+"Handler", "Func"),
		}
		handlers = []*handler{partHandler}
	} else if !idents.exists(def.Handler) {
		// a handler name was specified, create a new definition if it does not already exist
		partHandler = &handler{
			Info:       def.Name,
			Doc:        def.Doc,
			Identifier: def.Handler,
		}
		idents.reserve(def.Handler)
	} else {
		newEntry.Handler = def.Handler
	}
	if partHandler != nil {
		newEntry.Handler = partHandler.Identifier
		handlers = []*handler{partHandler}
	}

	for blockName, partials := range def.Blocks {
		if partHandler != nil {
			partHandler.Blocks = append(partHandler.Blocks, &block{
				Identifier: validIdentifier(blockName),
				Name:       blockName,
				FieldName:  validPublicIdentifier(blockName),
			})
		}
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
			subEntries, subRoutes, subHandlers, err := createEntries(idents, prefixN, blockEntry.Identifier, partial)
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
