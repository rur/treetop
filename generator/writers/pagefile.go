package writers

import (
	"fmt"

	"github.com/rur/treetop/generator"
)

type pageBlockData struct {
	Identifier string
	Name       string
}

type pageEntryData struct {
	Identifier string
	Name       string
	Extends    string
	Handler    string
	Type       string
	Template   string
}

type pageRouteData struct {
	Identifier string
	Path       string
	Type       string
}

type pageData struct {
	Namespace string
	Name      string
	Template  string
	Handler   string
	Blocks    []pageBlockData
	Entries   []pageEntryData
	Routes    []pageRouteData
}

func WritePageFile(dir string, pageDef *generator.PartialDef, namespace string) (string, error) {
	var fullpath string
	return fullpath, fmt.Errorf("Not Implemented")
}
