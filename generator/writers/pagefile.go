package writers

import (
	"fmt"

	"github.com/rur/treetop/generator"
)

type blockdata struct {
	Identifier string
	Name       string
}

type entrydata struct {
	Identifier string
	Name       string
	Extends    string
	Handler    string
	Type       string
	Template   string
}

type routedata struct {
	Identifier string
	Path       string
	Type       string
}

type pagedata struct {
	Namespace string
	Name      string
	Template  string
	Handler   string
	Blocks    []blockdata
	Entries   []entrydata
	Routes    []routedata
}

func WritePageFile(dir string, pageDef *generator.PartialDef, namespace string) (string, error) {
	var fullpath string
	return fullpath, fmt.Errorf("Not Implemented")
}
