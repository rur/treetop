package writers

import (
	"fmt"
	"os"
	"path"
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
	Path     string
	Extends  string
	Fragment bool
	Name     string
	Blocks   []*htmlBlockData
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
	fileName := "index.templ.html"
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

	blockList, err := iterateSortedBlocks(pageDef.Blocks)
	if err != nil {
		return fileName, err
	}
	blocks := make([]*htmlBlockData, 0, len(blockList))
	for _, block := range blockList {
		blockData := htmlBlockData{
			FieldName:  generator.ValidPublicIdentifier(block.name),
			Identifier: block.ident,
			Name:       block.name,
			Partials:   make([]*htmlBlockPartialData, 0, len(block.partials)),
		}
		blocks = append(blocks, &blockData)
		for _, partial := range block.partials {
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

func WriteTemplateBlock(dir string, blocks map[string][]generator.PartialDef) ([]string, error) {
	var created []string
	blockList, err := iterateSortedBlocks(blocks)
	if err != nil {
		return created, err
	}
	for _, block := range blockList {
		blockTemplDir := path.Join(dir, block.ident)
		if _, err := os.Stat(blockTemplDir); os.IsNotExist(err) {
			if err := os.Mkdir(blockTemplDir, os.ModePerm); err != nil {
				return created, fmt.Errorf("Error creating template dir '%s': %s", blockTemplDir, err)
			}
		}
		for _, def := range block.partials {
			files, err := writePartialTemplate(blockTemplDir, &def, block.name)
			if err != nil {
				return created, err
			}
			for _, file := range files {
				created = append(created, path.Join(block.ident, file))
			}
		}
	}
	return created, nil
}

func writePartialTemplate(dir string, def *generator.PartialDef, extends string) ([]string, error) {
	var created []string
	name, err := SanitizeName(def.Name)
	if err != nil {
		return created, fmt.Errorf("Invalid Partial name: '%s'", def.Name)
	}

	if def.Template == "" {
		// a template is not already defined, generate one
		partial := partialData{
			Path:     def.Path,
			Extends:  extends,
			Fragment: def.Fragment,
			Name:     def.Name,
			Blocks:   make([]*htmlBlockData, 0, len(def.Blocks)),
		}
		blockList, err := iterateSortedBlocks(def.Blocks)
		if err != nil {
			return created, err
		}
		for _, block := range blockList {
			blockData := htmlBlockData{
				FieldName:  generator.ValidPublicIdentifier(block.name),
				Identifier: block.ident,
				Name:       block.name,
				Partials:   make([]*htmlBlockPartialData, 0, len(block.partials)),
			}
			for _, bPartial := range block.partials {
				blockData.Partials = append(blockData.Partials, &htmlBlockPartialData{
					Path:     bPartial.Path,
					Name:     bPartial.Name,
					Fragment: bPartial.Fragment,
					Default:  bPartial.Default,
				})
			}
			partial.Blocks = append(partial.Blocks, &blockData)
		}

		fileName := fmt.Sprintf("%s.templ.html", name)
		filePath := filepath.Join(dir, fileName)
		sf, err := os.Create(filePath)
		if err != nil {
			return created, err
		}
		defer sf.Close()
		created = append(created, fileName)

		err = partialTemplate.Execute(sf, partial)
		if err != nil {
			return created, fmt.Errorf("Error executing partial template '%s': %s", fileName, err)
		}
	}

	// writer nested templates
	files, err := WriteTemplateBlock(dir, def.Blocks)
	if err != nil {
		return created, nil
	}
	created = append(created, files...)

	return created, nil
}
