package writers

import (
	"fmt"

	"github.com/rur/treetop/generator"
)

type blockpartialdata struct {
	Path     string
	Name     string
	Fragment bool
	Default  bool
}

type blockdata struct {
	FieldName  string
	Identifier string
	Name       string
	Partials   []*blockpartialdata
}

type templatedata struct {
	Path    string
	Extends string
	Name    string
	Blocks  []*blockdata
}

func WriteTemplateFiles(dir string, pageDef *generator.PartialDef) ([]string, error) {
	var created []string
	return created, fmt.Errorf("Not Implemented")
}
