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

type handlerdata struct {
	Info       string
	Doc        string
	Blocks     []*handlerBlockData
	Identifier string
}

type handlersdata struct {
	Namespace string
	PageName  string
	Handlers  []*handlerdata
}

func WriteHandlerFile(dir string, pageDef *generator.PartialDef, namespace string) (string, error) {
	pageName, err := sanitizeName(pageDef.Name)
	if err != nil {
		return "", fmt.Errorf("Invalid page name '%s'.", err)
	}

	fileName := filepath.Join("pages", pageName, "handlers.go")
	filePath := filepath.Join(dir, "handlers.go")
	sf, err := os.Create(filePath)
	if err != nil {
		return fileName, err
	}
	defer sf.Close()

	var handlers []*handlerdata
	// base page handler
	pageHandler := handlerdata{
		Info:       pageName,
		Doc:        pageDef.Doc,
		Blocks:     make([]*handlerBlockData, 0, len(pageDef.Blocks)),
		Identifier: pageName + "Handler",
	}
	handlers = append(handlers, &pageHandler)

	for rawName, partials := range pageDef.Blocks {
		blockIdent, err := sanitizeName(rawName)
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

func processHandlersDef(blockName string, def *generator.PartialDef) ([]*handlerdata, error) {
	return []*handlerdata{}, nil
}
