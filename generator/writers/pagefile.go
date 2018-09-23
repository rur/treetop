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
	Reference         string
	Path              string
	PartialAsFragment bool
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
		})
	}

	blocks := make([]pageBlockData, 0, len(pageDef.Blocks))
	for nme, partials := range pageDef.Blocks {
		blockName, err := SanitizeName(nme)
		if err != nil {
			return "", fmt.Errorf("Invalid block name '%s'", nme)
		}
		blocks = append(blocks, pageBlockData{
			Identifier: assignBlock(partials, blockName),
			Name:       nme,
		})

		for _, partial := range partials {
			blockEntries, blockRoutes, err := processEntries(
				blockName,
				&partial,
				filepath.Join("pages", pageName, "templates", blockName),
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

	page := pageData{
		Namespace: namespace,
		Name:      pageName,
		Template:  filepath.Join("pages", pageName, "templates", "index.templ.html"),
		Handler:   pageName + "PageHandler",
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

func processEntries(extends string, def *generator.PartialDef, templatePath string) ([]pageEntryData, []pageRouteData, error) {
	var entryType string
	var suffix string
	var entries []pageEntryData
	var routes []pageRouteData

	if def.Default {
		entryType = "DefaultPartial"
		suffix = "dfl"
	} else if def.Fragment {
		entryType = "Fragment"
		suffix = "frg"
	} else {
		entryType = "Partial"
		suffix = "ptl"
	}

	entryName, err := SanitizeName(def.Name)
	if err != nil {
		return entries, routes, fmt.Errorf("Invalid %s name '%s'", entryType, def.Name)
	}

	entry := pageEntryData{
		Identifier: assignHandler(def, entryName+"_"+suffix),
		Name:       entryName,
		Extends:    extends,
		Handler:    entryName + "Handler",
		Type:       entryType,
		Template:   filepath.Join(templatePath, entryName+".templ.html"),
	}

	if def.Path != "" {
		routes = append(routes, pageRouteData{
			Reference:         entryName + "_" + suffix,
			Path:              strings.Trim(def.Path, " "),
			PartialAsFragment: def.Fragment && def.Default,
		})
	}

	entries = append(entries, entry)

	for nme, partials := range def.Blocks {
		blockName, err := SanitizeName(nme)
		if err != nil {
			return entries, routes, fmt.Errorf("Invalid block name '%s'", nme)
		}

		entries = append(entries, pageEntryData{
			Type: "Spacer",
		}, pageEntryData{
			Identifier: assignBlock(partials, entryName+"_"+blockName),
			Name:       blockName,
			Extends:    entryName + "_" + suffix,
			Type:       "Block",
		})

		for _, partial := range partials {
			blockEntries, blockRoutes, err := processEntries(
				entryName+"_"+blockName,
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
