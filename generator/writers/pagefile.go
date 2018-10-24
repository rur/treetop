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
	Identifier      string
	Name            string
	Extends         string
	Handler         string
	OverrideHandler bool
	Type            string
	Template        string
}

type pageRouteData struct {
	Reference string
	Path      string
	Type      string
}

type pageTemplateData struct {
	Identifier string
	Path       string
	Type       string
}

type pageData struct {
	Namespace       string
	Name            string
	Template        string
	Handler         string
	OverrideHandler bool
	Blocks          []pageBlockData
	Entries         []pageEntryData
	Routes          []pageRouteData
}

func assignHandler(def *generator.PartialDef, name string) string {
	if def.Path == "" && len(def.Blocks) == 0 {
		return "_ ="
	} else {
		return name + " :="
	}
}

func assignBlock(defs []generator.PartialDef, name string) string {
	if len(defs) == 0 {
		return "_ ="
	} else {
		return name + " :="
	}
}

func WritePageFile(dir string, pageDef *generator.PartialDef, namespace string) (string, error) {
	pageName, err := SanitizeName(pageDef.Name)
	if err != nil {
		return "", fmt.Errorf("Invalid page name '%s'.", err)
	}

	fileName := "page.go"
	filePath := filepath.Join(dir, "page.go")
	sf, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer sf.Close()

	var entries []pageEntryData
	var routes []pageRouteData
	if pageDef.Path != "" {
		routes = append(routes, pageRouteData{
			Reference: "page",
			Path:      strings.Trim(pageDef.Path, " "),
			Type:      "Page",
		})
	}

	blocks := make([]pageBlockData, 0, len(pageDef.Blocks))

	sortedBlocks, err := iterateSortedBlocks(pageDef.Blocks)
	if err != nil {
		return fileName, err
	}
	for _, block := range sortedBlocks {
		// for nme, partials := range pageDef.Blocks {
		blocks = append(blocks, pageBlockData{
			Identifier: assignBlock(block.partials, block.ident+"Blk"),
			Name:       block.name,
		})

		for _, partial := range block.partials {
			blockEntries, blockRoutes, err := processEntries(
				block.ident+"Blk",
				&partial,
				filepath.Join("page", pageName, "templates", block.ident),
			)
			if err != nil {
				return "", err
			}
			entries = append(entries, blockEntries...)
			routes = append(routes, blockRoutes...)
		}
	}

	if len(routes) == 0 {
		return "", fmt.Errorf("Page '%s' does not have any routes!", pageName)
	}

	handler := pageDef.Handler
	if handler == "" {
		handler = pageName + "PageHandler"
	}

	page := pageData{
		Namespace:       namespace,
		Name:            pageName,
		Template:        filepath.Join("page", pageName, "templates", "index.templ.html"),
		Handler:         handler,
		OverrideHandler: pageDef.Handler != "",
		Blocks:          blocks,
		Entries:         entries,
		Routes:          routes,
	}

	err = pageTemplate.Execute(sf, page)
	if err != nil {
		return fileName, err
	}

	return fileName, nil
}

func processEntries(extends string, def *generator.PartialDef, templatePath string) ([]pageEntryData, []pageRouteData, error) {
	var entryType string
	var entries []pageEntryData
	var routes []pageRouteData

	if def.Default {
		entryType = "Default"
	} else {
		entryType = "Extend"
	}

	entryName, err := SanitizeName(def.Name)
	if err != nil {
		return entries, routes, fmt.Errorf("Invalid %s name '%s'", entryType, def.Name)
	}

	handler := def.Handler
	if handler == "" {
		handler = entryName + "Handler"
	}

	entry := pageEntryData{
		Identifier:      assignHandler(def, entryName+"Def"),
		Name:            entryName,
		Extends:         extends,
		Handler:         handler,
		OverrideHandler: def.Handler != "",
		Type:            entryType,
		Template:        filepath.Join(templatePath, entryName+".templ.html"),
	}

	if def.Path != "" {
		route := pageRouteData{
			Reference: entryName + "Def",
			Path:      strings.Trim(def.Path, " "),
		}
		if def.Fragment {
			route.Type = "Fragment"
		} else {
			route.Type = "Partial"
		}
		routes = append(routes, route)
	}

	entries = append(entries, entry)

	sortedBlocks, err := iterateSortedBlocks(def.Blocks)
	if err != nil {
		return entries, routes, err
	}
	for _, block := range sortedBlocks {
		entries = append(entries, pageEntryData{
			Type: "Spacer",
		}, pageEntryData{
			Identifier: assignBlock(block.partials, entryName+"_"+block.ident+"Blk"),
			Name:       block.ident,
			Extends:    entryName + "Def",
			Type:       "Block",
		})

		for _, partial := range block.partials {
			blockEntries, blockRoutes, err := processEntries(
				entryName+"_"+block.ident+"Blk",
				&partial,
				filepath.Join(templatePath, block.ident),
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
