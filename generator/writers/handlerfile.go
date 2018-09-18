package writers

import (
	"fmt"

	"github.com/rur/treetop/generator"
)

type blockdata struct {
	Identifier string
	Name       string
	FieldName  string
}

type handlerdata struct {
	Info       string
	Doc        string
	Blocks     []blockdata
	Identifier string
}

type handlersdata struct {
	Namespace string
	PageName  string
	Handlers  []handlerdata
}

func WriteHandlerFile(dir string, pageDef *generator.PartialDef, namespace string) (string, error) {
	var fullpath string
	return fullpath, fmt.Errorf("Not Implemented")
}
