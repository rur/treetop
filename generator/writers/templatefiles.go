package writers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rur/treetop/generator"
)

type htmlBlockPartialData struct {
	Path     string
	Name     string
	Fragment bool
	Default  bool
}

type htmlBlockData struct {
	FieldName  string
	Identifier string
	Name       string
	Partials   []*htmlBlockPartialData
}

type partialData struct {
	Path    string
	Extends string
	Name    string
	Blocks  []*htmlBlockData
	Type    string
}

type indexSiteLinksData struct {
	URI    string
	Label  string
	Active bool
}

type indexData struct {
	Title     string
	SiteLinks []*indexSiteLinksData
	Blocks    []*htmlBlockData
}

func WriteIndexFile(dir string, pageDef *generator.PartialDef, otherPages []generator.PartialDef) (string, error) {
	pageName, err := sanitizeName(pageDef.Name)
	if err != nil {
		return "", fmt.Errorf("Invalid page name '%s'.", err)
	}

	fileName := filepath.Join("pages", pageName, "templates", "index.templ.html")
	filePath := filepath.Join(dir, "index.templ.html")
	sf, err := os.Create(filePath)
	if err != nil {
		return fileName, err
	}
	defer sf.Close()

	links := make([]*indexSiteLinksData, 0, len(otherPages))
	for _, other := range otherPages {
		if other.URI != "" {
			links = append(links, &indexSiteLinksData{
				URI:    other.URI,
				Label:  other.Name,
				Active: other.URI != pageDef.URI,
			})
		}
	}

	blocks := make([]*htmlBlockData, 0, len(pageDef.Blocks))
	for blockName, partials := range pageDef.Blocks {
		blockData := htmlBlockData{
			FieldName: generator.ValidPublicIdentifier(blockName),
			Name:      blockName,
			Partials:  make([]*htmlBlockPartialData, 0, len(partials)),
		}
		blocks = append(blocks, &blockData)
		for _, partial := range partials {
			blockData.Partials = append(blockData.Partials, &htmlBlockPartialData{
				Path:     partial.Path,
				Name:     partial.Name,
				Fragment: partial.Fragment,
				Default:  partial.Default,
			})
		}
	}

	err = indexTemplate.Execute(sf, indexData{
		Title:     pageDef.Name,
		SiteLinks: links,
		Blocks:    blocks,
	})
	if err != nil {
		return fileName, err
	}

	return fileName, nil
}

func WriteTemplateFiles(dir string, pageDef *generator.PartialDef) ([]string, error) {
	var created []string
	return created, fmt.Errorf("Not Implemented")
}
