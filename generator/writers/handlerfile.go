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
	// base page handler
	pageHandler := handlerData{
		Info:       pageName,
		Doc:        pageDef.Doc,
		Type:       "(page)",
		Blocks:     make([]*handlerBlockData, 0, len(pageDef.Blocks)),
		Identifier: pageName + "PageHandler",
	}
	handlers = append(handlers, &pageHandler)

	for rawName, partials := range pageDef.Blocks {
		blockIdent, err := SanitizeName(rawName)
		if err != nil {
			return fileName, fmt.Errorf("Invalid block name '%s'", rawName)
		}
		pageHandler.Blocks = append(pageHandler.Blocks, &handlerBlockData{
			Identifier: blockIdent + "Data",
			Name:       rawName,
			FieldName:  generator.ValidPublicIdentifier(rawName),
		})

		for _, partial := range partials {
			blockHandlers, err := processHandlersDef(blockIdent, &partial)
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

	// base page handler
	handler := handlerData{
		Info:       entryName,
		Doc:        def.Doc,
		Type:       entryType,
		Blocks:     make([]*handlerBlockData, 0, len(def.Blocks)),
		Identifier: entryName + "Handler",
	}
	handlers = append(handlers, &handler)

	for rawName, partials := range def.Blocks {
		blockIdent, err := SanitizeName(rawName)
		if err != nil {
			return handlers, fmt.Errorf("Invalid block name '%s'", rawName)
		}
		handler.Blocks = append(handler.Blocks, &handlerBlockData{
			Identifier: blockIdent + "Data",
			Name:       rawName,
			FieldName:  generator.ValidPublicIdentifier(rawName),
		})

		for _, partial := range partials {
			blockHandlers, err := processHandlersDef(blockIdent, &partial)
			if err != nil {
				return handlers, err
			}
			handlers = append(handlers, blockHandlers...)
		}
	}

	return handlers, nil
}
