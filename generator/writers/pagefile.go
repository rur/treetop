package writers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rur/treetop/generator"
)

type pageBlockData struct {
	Identifier string
	Name       string
}

type pageEntryData struct {
	Identifier string
	Name       string
	Extends    string
	Handler    string
	Type       string
	Template   string
}

type pageRouteData struct {
	Identifier string
	Path       string
}

type pageTemplateData struct {
	Identifier string
	Path       string
	Type       string
}

type pageData struct {
	Namespace string
	Name      string
	Template  string
	Handler   string
	Blocks    []pageBlockData
	Entries   []pageEntryData
	Routes    []pageRouteData
}

func WritePageFile(dir string, pageDef *generator.PartialDef, namespace string) (string, error) {
	pageName, err := sanitizeName(pageDef.Name)
	if err != nil {
		return "", fmt.Errorf("Invalid page name '%s'.", err)
	}

	fileName := filepath.Join("pages", pageName, "page.go")
	filePath := filepath.Join(dir, "page.go")
	sf, err := os.Create(filePath)
	if err != nil {
		return fileName, err
	}
	defer sf.Close()

	var entries []pageEntryData
	var routes []pageRouteData
	if pageDef.Path != "" {
		routes = append(routes, pageRouteData{
			Identifier: "page",
			Path:       strings.Trim(pageDef.Path, " "),
		})
	}

	blocks := make([]pageBlockData, 0, len(pageDef.Blocks))
	for nme, partials := range pageDef.Blocks {
		blockName, err := sanitizeName(nme)
		if err != nil {
			return fileName, fmt.Errorf("Invalid block name '%s'", nme)
		}
		blocks = append(blocks, pageBlockData{
			Identifier: nme,
			Name:       nme,
		})

		for _, partial := range partials {

			blockEntries, blockRoutes, err := processPartialDef(
				blockName,
				&partial,
				filepath.Join("pages", pageName, "templates", blockName),
			)

			if err != nil {
				return fileName, err
			}
			entries = append(entries, blockEntries...)
			routes = append(routes, blockRoutes...)
		}
	}

	page := pageData{
		Namespace: namespace,
		Name:      pageName,
		Template:  filepath.Join("pages", pageName, "templates", "index.html"),
		Handler:   pageName + "Handler",
		Blocks:    blocks,
		Entries:   entries,
		Routes:    routes,
	}

	err = pageTemplate.Execute(sf, page)
	if err != nil {
		return fileName, err
	}

	return fileName, nil
}

func sanitizeName(name string) (string, error) {
	return generator.ValidIdentifier(name), nil
}

func processPartialDef(extends string, def *generator.PartialDef, templatePath string) ([]pageEntryData, []pageRouteData, error) {
	var entryType string
	var suffix string
	var entries []pageEntryData
	var routes []pageRouteData

	if def.Fragment {
		entryType = "Fragment"
		suffix = "frg"
	} else if def.Default {
		entryType = "DefaultPartial"
		suffix = "dfl"
	} else {
		entryType = "Partial"
		suffix = "prt"
	}

	entryName, err := sanitizeName(def.Name)
	if err != nil {
		return entries, routes, fmt.Errorf("Invalid %s name '%s'", entryType, def.Name)
	}

	entry := pageEntryData{
		Identifier: entryName + "_" + suffix,
		Name:       entryName,
		Extends:    extends,
		Handler:    extends + "_" + entryName + "Handler",
		Type:       entryType,
		Template:   filepath.Join(templatePath, entryName+".templ.html"),
	}

	if def.Path != "" {
		routes = append(routes, pageRouteData{
			Identifier: entry.Identifier,
			Path:       strings.Trim(def.Path, " "),
		})
	}

	entries = append(entries, entry)

	for nme, partials := range def.Blocks {
		blockName, err := sanitizeName(nme)
		if err != nil {
			return entries, routes, fmt.Errorf("Invalid block name '%s'", nme)
		}

		entries = append(entries, pageEntryData{
			Type: "Spacer",
		}, pageEntryData{
			Identifier: blockName,
			Name:       blockName,
			Extends:    entryName,
			Type:       "Block",
		})

		for _, partial := range partials {

			blockEntries, blockRoutes, err := processPartialDef(
				blockName,
				&partial,
				filepath.Join(templatePath, blockName),
			)

			if err != nil {
				return entries, routes, err
			}
			entries = append(entries, blockEntries...)
			routes = append(routes, blockRoutes...)
		}
	}

	return entries, routes, nil
}
