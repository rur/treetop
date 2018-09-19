package writers

import (
	"fmt"
	"regexp"

	"github.com/rur/treetop/generator"
)

func sanitizeName(name string) (string, error) {
	re := regexp.MustCompile("(?i)^[a-z]{3}[a-z0-9-_]*$")
	if !re.MatchString(name) {
		return name, fmt.Errorf("Invalid name '%s'", name)
	}
	return generator.ValidIdentifier(name), nil
}
