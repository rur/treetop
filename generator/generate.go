package generator

//go:generate go run gen.go

import (
	"fmt"
	html "html/template"
	"log"
	"os"
	"path/filepath"
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
	Template   string
	Type       string
	Name       string
}

type blockPartial struct {
	Path     string
	Name     string
	Fragment bool
	Default  bool
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
	Template   *template
	Handler    string
	Doc        string
	Name       string
	Entries    []*entry
	Routes     []*route
	Blocks     []*block
	Handlers   []*handler
	Templates  []*template
}

type template struct {
	Path    string
	Extends string
	Name    string
	Blocks  []*block
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
			newPage.Template.Blocks = append(newPage.Template.Blocks, &newBlock)

			for _, partial := range partials {
				newBlock.Partials = append(newBlock.Partials, &blockPartial{
					Path:     partial.Path,
					Name:     partial.Name,
					Fragment: partial.Fragment,
					Default:  partial.Default,
				})

				entries, routes, handlers, templates, err := createEntries(
					&idents,
					lowercaseName(def.Name),
					entry{
						Identifier: newBlock.Identifier,
						Type:       "Block",
						Name:       newBlock.Name,
					},
					partial,
				)
				if err != nil {
					return created, err
				}
				newPage.Entries = append(newPage.Entries, entries...)
				newPage.Routes = append(newPage.Routes, routes...)
				newPage.Handlers = append(newPage.Handlers, handlers...)
				newPage.Templates = append(newPage.Templates, templates...)
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
		pageFiles, err := writePageFiles(dir, sitePage)
		if err != nil {
			return created, err
		}
		created = append(created, pageFiles...)
	}

	return created, nil
}

func createPage(idents *uniqueIdentifiers, def PartialDef) (page, *handler) {
	var pageHandler *handler
	newPage := page{
		Name:       def.Name,
		Identifier: idents.new(def.Name, "Page"),
		Entries:    make([]*entry, 0),
		Blocks:     make([]*block, 0, len(def.Blocks)),
	}
	if def.Template == "" {
		newPage.Template = &template{
			Path: filepath.Join("templates", fmt.Sprintf("%s.templ.html", lowercaseName(def.Name))),
			Name: def.Name,
		}
	} else {
		newPage.Template = &template{
			Path: def.Template,
			Name: def.Name,
		}
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

func createEntries(idents *uniqueIdentifiers, prefix string, extends entry, def PartialDef) ([]*entry, []*route, []*handler, []*template, error) {
	var prefixN string
	if prefix != "" {
		prefixN = fmt.Sprintf("%s_%s", prefix, lowercaseName(def.Name))
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
		Extends:    extends.Identifier,
		Identifier: idents.new(def.Name, entryType),
		Type:       entryType,
		Template:   def.Template,
		Name:       def.Name,
	}

	if newEntry.Template == "" {
		newEntry.Template = filepath.Join("templates", fmt.Sprintf("%s.templ.html", prefixN))
	}

	partTemplate := template{
		Path:    newEntry.Template,
		Extends: extends.Name,
		Name:    def.Name,
	}
	templates := []*template{&partTemplate}

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
		partBlock := block{
			Identifier: validIdentifier(blockName),
			Name:       blockName,
			FieldName:  validPublicIdentifier(blockName),
		}
		if partHandler != nil {
			partHandler.Blocks = append(partHandler.Blocks, &partBlock)
		}
		partTemplate.Blocks = append(partTemplate.Blocks, &partBlock)

		blockEntry := entry{
			Extends:    newEntry.Identifier,
			Identifier: idents.new(blockName, "Block"),
			Type:       "Block",
			Name:       blockName,
		}
		entries = append(
			entries,
			&entry{
				Type: "Spacer",
			},
			&blockEntry,
		)
		for _, partial := range partials {
			partBlock.Partials = append(partBlock.Partials, &blockPartial{
				Path:     partial.Path,
				Name:     partial.Name,
				Fragment: partial.Fragment,
				Default:  partial.Default,
			})

			subEntries, subRoutes, subHandlers, subTemplates, err := createEntries(idents, prefixN, blockEntry, partial)
			if err != nil {
				return entries, routes, handlers, templates, err
			}
			entries = append(entries, subEntries...)
			routes = append(routes, subRoutes...)
			handlers = append(handlers, subHandlers...)
			templates = append(templates, subTemplates...)
		}
		entries = append(entries, &entry{
			Type: "Spacer",
		})
	}

	return entries, routes, handlers, templates, nil
}

func writePageFiles(dir string, p page) ([]string, error) {
	var created []string
	handlerFile := filepath.Join("server", p.Identifier+"_handlers.go")
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
		Handlers: p.Handlers,
	})
	if err != nil {
		return created, err
	}
	if p.Template != nil {
		// page template
		pt := p.Template
		path := filepath.Join(dir, pt.Path)
		tf, err := os.Create(path)
		if err != nil {
			return created, err
		}
		created = append(created, pt.Path)
		defer tf.Close()
		err = pageTemplate.Execute(tf, pt)
		if err != nil {
			return created, err
		}
	}
	for _, templ := range p.Templates {
		// partial template
		path := filepath.Join(dir, templ.Path)
		tf, err := os.Create(path)
		if err != nil {
			return created, err
		}
		created = append(created, templ.Path)
		defer tf.Close()
		err = partialTemplate.Execute(tf, templ)
		if err != nil {
			return created, err
		}
	}
	return created, err
}
