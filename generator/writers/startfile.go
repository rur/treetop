package writers

import (
	"fmt"

	"github.com/rur/treetop/generator"
)

type startdata struct {
	Namespace string
	Pages     []string
}

func WriteStartFile(dir string, pages []generator.PartialDef, namespace string) (string, error) {
	var fullpath string
	return fullpath, fmt.Errorf("Not Implemented")
}
