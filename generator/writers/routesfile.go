package writers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rur/treetop/generator"
)

type pageBlockData struct {
	Name string
}

type pageEntryData struct {
	Assignment      string
	Name            string
	Extends         string
	Block           string
	Handler         string
	OverrideHandler bool
	Type            string
	Template        string
	Path            string
}

type pageRouteData struct {
	Reference string
	Path      string
	Type      string
	Includes  []string
}

type pageTemplateData struct {
	Assignment string
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

func WriteRoutesFile(dir string, pageDef *generator.PartialDef, namespace string) (string, error) {
	pageName, err := SanitizeName(pageDef.Name)
	if err != nil {
		return "", fmt.Errorf("Invalid page name '%s'.", err)
	}

	fileName := "routes.go"
	filePath := filepath.Join(dir, "routes.go")
	sf, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer sf.Close()

	var entries []pageEntryData
	var routes []pageRouteData
	if pageDef.Path != "" {
		routes = append(routes, pageRouteData{
			Reference: "pageView",
			Path:      strings.Trim(pageDef.Path, " "),
			Type:      "Page",
			Includes:  append([]string{}, pageDef.Includes...),
		})
	}

	blocks := make([]pageBlockData, 0, len(pageDef.Blocks))

	sortedBlocks, err := iterateSortedBlocks(pageDef.Blocks)
	if err != nil {
		return fileName, err
	}
	for _, block := range sortedBlocks {
		entries = append(entries, pageEntryData{
			Name: block.name,
			Type: "Spacer",
		})

		for i, partial := range block.partials {
			blockEntries, blockRoutes, err := processEntries(
				"pageView",
				block.name,
				[]string{pageDef.Name, partial.Name},
				&partial,
				filepath.Join("page", pageName, "templates", block.ident),
				block.name,
			)
			if err != nil {
				return "", err
			}
			entries = append(entries, blockEntries...)
			routes = append(routes, blockRoutes...)
			if len(blockEntries) > 1 && i < len(block.partials)-1 {
				entries = append(entries, pageEntryData{
					Name: block.name,
					Type: "Spacer",
				})
			}
		}
	}

	if len(routes) == 0 {
		return "", fmt.Errorf("Page '%s' does not have any routes!", pageName)
	}

	// process includes in routes by scanning entries for matching paths
	pathMap := make(map[string]int)
	for index, en := range entries {
		if en.Path != "" {
			pathMap[en.Path] = index
		}
	}
	for _, route := range routes {
		for i, incl := range route.Includes {
			if j, ok := pathMap[incl]; ok {
				route.Includes[i] = entries[j].Name
				entries[j].Assignment = entries[j].Name + " :="
			} else {
				return "", fmt.Errorf("Failed to match include path '%s' to a sub view entry for route '%s'", incl, route.Path)
			}
		}
	}

	handler := pageDef.Handler
	if handler == "" {
		handler = pageName + "PageHandler"
	}

	template := pageDef.Template
	if template == "" {
		template = filepath.Join("page", pageName, "templates", "index.html.tmpl")
	}

	page := pageData{
		Namespace:       namespace,
		Name:            pageName,
		Template:        template,
		Handler:         handler,
		OverrideHandler: pageDef.Handler != "",
		Blocks:          blocks,
		Entries:         entries,
		Routes:          routes,
	}

	err = routesTemplate.Execute(sf, page)
	if err != nil {
		return fileName, err
	}

	return fileName, nil
}

func processEntries(extends, blockName string, names []string, def *generator.PartialDef, templatePath string, seen ...string) ([]pageEntryData, []pageRouteData, error) {
	var entryType string
	var entries []pageEntryData
	var routes []pageRouteData

	if def.Default {
		entryType = "DefaultSubView"
	} else {
		entryType = "SubView"
	}

	entryName, err := SanitizeName(def.Name)
	if err != nil {
		return entries, routes, fmt.Errorf("Invalid %s name '%s' @ %s", entryType, def.Name, strings.Join(seen, " -> "))
	}

	handler := def.Handler
	if handler == "" {
		handler = entryName + "Handler"
	}

	template := def.Template
	if template == "" {
		template = filepath.Join(templatePath, entryName+".html.tmpl")
	}

	entry := pageEntryData{
		Name:            entryName,
		Extends:         extends,
		Block:           blockName,
		Handler:         handler,
		OverrideHandler: def.Handler != "",
		Type:            entryType,
		Template:        template,
		Path:            strings.Join(names, " > "),
	}

	// the assignment for an entry must be blanked if there are no routes or subviews
	// assignment may be reinstated if this entry is used as an include for another route
	if def.Path == "" && len(def.Blocks) == 0 {
		entry.Assignment = "_ ="
	} else {
		entry.Assignment = entryName + " :="
	}

	if def.Path != "" {
		route := pageRouteData{
			Reference: entryName,
			Path:      strings.Trim(def.Path, " "),
			Includes:  append([]string{}, def.Includes...),
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
		if len(block.partials) == 0 {
			continue
		}
		entries = append(entries, pageEntryData{
			Name: strings.Join(append(seen, block.name), " -> "),
			Type: "Spacer",
		})

		for _, partial := range block.partials {
			blockEntries, blockRoutes, err := processEntries(
				entry.Name,
				block.name,
				append(names, partial.Name),
				&partial,
				filepath.Join(templatePath, block.ident),
				append(seen, block.name)...,
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
