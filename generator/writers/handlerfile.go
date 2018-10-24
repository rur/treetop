package writers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rur/treetop/generator"
)

type handlerBlockData struct {
	Identifier string
	Name       string
	FieldName  string
}

type handlerData struct {
	Info       string
	Type       string
	Doc        string
	Extends    string
	Blocks     []*handlerBlockData
	Identifier string
}

type handlersdata struct {
	Namespace string
	PageName  string
	Handlers  []*handlerData
}

func WriteHandlerFile(dir string, pageDef *generator.PartialDef, namespace string) (string, error) {
	pageName, err := SanitizeName(pageDef.Name)
	if err != nil {
		return "", fmt.Errorf("Invalid page name '%s'.", err)
	}

	fileName := "handlers.go"
	filePath := filepath.Join(dir, "handlers.go")
	sf, err := os.Create(filePath)
	if err != nil {
		return fileName, err
	}
	defer sf.Close()

	var handlers []*handlerData
	var pageHandler *handlerData
	if pageDef.Handler == "" {
		// base page handler
		pageHandler = &handlerData{
			Info:       pageName,
			Doc:        pageDef.Doc,
			Type:       "(page)",
			Blocks:     make([]*handlerBlockData, 0, len(pageDef.Blocks)),
			Identifier: pageName + "PageHandler",
		}
		handlers = append(handlers, pageHandler)
	}

	blocks, err := iterateSortedBlocks(pageDef.Blocks)
	if err != nil {
		return fileName, err
	}

	for _, block := range blocks {
		if pageHandler != nil {
			pageHandler.Blocks = append(pageHandler.Blocks, &handlerBlockData{
				Identifier: block.ident + "Data",
				Name:       block.name,
				FieldName:  generator.ValidPublicIdentifier(block.name),
			})
		}

		for _, partial := range block.partials {
			blockHandlers, err := processHandlersDef(block.ident, &partial)
			if err != nil {
				return fileName, err
			}
			handlers = append(handlers, blockHandlers...)
		}
	}

	err = handlerTemplate.Execute(sf, handlersdata{
		Namespace: namespace,
		PageName:  pageName,
		Handlers:  handlers,
	})
	if err != nil {
		return fileName, err
	}

	return fileName, nil
}

func processHandlersDef(blockName string, def *generator.PartialDef) ([]*handlerData, error) {
	var handlers []*handlerData
	var entryType string
	if def.Fragment {
		entryType = "(fragment)"
	} else if def.Default {
		entryType = "(default partial)"
	} else {
		entryType = "(partial)"
	}

	entryName, err := SanitizeName(def.Name)
	if err != nil {
		return handlers, fmt.Errorf("Invalid name '%s'", def.Name)
	}

	var handler *handlerData

	if def.Handler == "" {
		// base page handler
		handler = &handlerData{
			Info:       entryName,
			Extends:    blockName,
			Doc:        def.Doc,
			Type:       entryType,
			Blocks:     make([]*handlerBlockData, 0, len(def.Blocks)),
			Identifier: entryName + "Handler",
		}
		handlers = append(handlers, handler)
	}

	blocks, err := iterateSortedBlocks(def.Blocks)
	if err != nil {
		return handlers, err
	}

	for _, block := range blocks {
		if handler != nil {
			handler.Blocks = append(handler.Blocks, &handlerBlockData{
				Identifier: block.ident + "Data",
				Name:       block.name,
				FieldName:  generator.ValidPublicIdentifier(block.name),
			})
		}

		for _, partial := range block.partials {
			blockHandlers, err := processHandlersDef(block.ident, &partial)
			if err != nil {
				return handlers, err
			}
			handlers = append(handlers, blockHandlers...)
		}
	}

	return handlers, nil
}
