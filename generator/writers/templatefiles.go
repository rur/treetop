package writers

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"

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
}

// NOTE: Make fragment data a duplicate of partial template data for now,
//       if they do not diverge as we refine the templates then they
//       should be merged again.
type fragmentData struct {
	Path    string
	Extends string
	Name    string
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

type blockDef struct {
	name     string
	ident    string
	partials []generator.PartialDef
}

func iterateSortedBlocks(blocks map[string][]generator.PartialDef) ([]blockDef, error) {
	output := make([]blockDef, 0, len(blocks))
	var keys []string
	for k := range blocks {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		ident, err := SanitizeName(k)
		if err != nil {
			return output, fmt.Errorf("Invalid block name '%s'", k)
		}
		output = append(output, blockDef{
			name:     k,
			ident:    ident,
			partials: blocks[k],
		})
	}
	return output, nil
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

func WriteTemplateBlock(dir string, blocks map[string][]generator.PartialDef) ([]string, error) {
	var created []string
	blockList, err := iterateSortedBlocks(blocks)
	if err != nil {
		return created, err
	}
	for _, block := range blockList {
		blockTemplDir := path.Join(dir, block.ident)
		if err := os.Mkdir(blockTemplDir, os.ModePerm); err != nil {
			return created, fmt.Errorf("Error creating template dir '%s': %s", blockTemplDir, err)
		}
		for _, def := range block.partials {
			if def.Fragment {
				file, err := writeFragmentFile(blockTemplDir, &def, block.name)
				if err != nil {
					return created, fmt.Errorf("Error creating fragment %s for block %s", def.Name, block.name)
				}
				created = append(created, path.Join(block.ident, file))
			} else {
				files, err := writePartialFiles(blockTemplDir, &def, block.name)
				if err != nil {
					return created, err
				}
				for _, file := range files {
					created = append(created, path.Join(block.ident, file))
				}
			}
		}
	}
	return created, nil
}

func writePartialFiles(dir string, def *generator.PartialDef, extends string) ([]string, error) {
	var created []string
	name, err := SanitizeName(def.Name)
	if err != nil {
		return created, fmt.Errorf("Invalid Partial name: '%s'", def.Name)
	}

	partial := partialData{
		Path:    def.Path,
		Extends: extends,
		Name:    def.Name,
		Blocks:  make([]*htmlBlockData, 0, len(def.Blocks)),
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

	err = partialTemplate.Execute(sf, partial)
	if err != nil {
		return created, fmt.Errorf("Error executing partial template '%s': %s", fileName, err)
	}

	// writer nested templates
	files, err := WriteTemplateBlock(dir, def.Blocks)
	if err != nil {
		return created, nil
	}
	created = append(created, files...)

	return created, nil
}

func writeFragmentFile(dir string, def *generator.PartialDef, extends string) (string, error) {
	name, err := SanitizeName(def.Name)
	if err != nil {
		return "", fmt.Errorf("Invalid Fragment name: '%s'", def.Name)
	}

	fragment := fragmentData{
		Path:    def.Path,
		Extends: extends,
		Name:    def.Name,
	}

	fileName := fmt.Sprintf("%s.templ.html", name)
	filePath := filepath.Join(dir, fileName)
	sf, err := os.Create(filePath)
	if err != nil {
		return fileName, err
	}
	defer sf.Close()

	err = fragmentTemplate.Execute(sf, fragment)
	if err != nil {
		return fileName, fmt.Errorf("Error executing fragment template '%s': %s", fileName, err)
	}
	return fileName, nil
}
