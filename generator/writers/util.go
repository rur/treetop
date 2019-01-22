package writers

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/rur/treetop/generator"
)

func SanitizeName(name string) (string, error) {
	re := regexp.MustCompile("(?i)^[a-z][a-z0-9-_]*$")
	if !re.MatchString(name) {
		return name, fmt.Errorf("Invalid name '%s'", name)
	}
	return generator.ValidIdentifier(name), nil
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
