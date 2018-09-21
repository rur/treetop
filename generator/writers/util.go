package writers

import (
	"fmt"
	"regexp"

	"github.com/rur/treetop/generator"
)

// TODO: move this to generator module once it becomes clear how it should work
func SanitizeName(name string) (string, error) {
	re := regexp.MustCompile("(?i)^[a-z][a-z0-9-_]*$")
	if !re.MatchString(name) {
		return name, fmt.Errorf("Invalid name '%s'", name)
	}
	return generator.ValidIdentifier(name), nil
}
