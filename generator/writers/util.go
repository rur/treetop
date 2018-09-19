package writers

import "github.com/rur/treetop/generator"

func sanitizeName(name string) (string, error) {
	return generator.ValidIdentifier(name), nil
}
