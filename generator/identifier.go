package generator

import (
	"strings"
	"unicode"
)

func delims(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsNumber(r)
}

// convert a name to a valid identifier, leading digits will be stripped
func ValidIdentifier(name string) string {
	parts := strings.FieldsFunc(name, delims)
	fixed := make([]string, len(parts))

	leading := strings.TrimLeft(parts[0], "0123456789")
	if len(leading) == 0 {
		fixed[0] = "var"
	} else {
		mutable := []rune(parts[0])
		mutable[0] = unicode.ToLower(mutable[0])
		fixed[0] = string(mutable)
	}

	for i := 1; i < len(parts); i++ {
		mutable := []rune(parts[i])
		mutable[0] = unicode.ToUpper(mutable[0])
		fixed[i] = string(mutable)
	}
	return strings.Join(fixed, "")
}

// convert a name to a valid public identifier, leading digits will be stripped
func ValidPublicIdentifier(name string) string {
	parts := strings.FieldsFunc(name, delims)
	fixed := make([]string, len(parts))

	leading := strings.TrimLeft(parts[0], "0123456789")
	if len(leading) == 0 {
		fixed[0] = "var"
	} else {
		fixed[0] = strings.Title(leading)
	}

	for i := 1; i < len(parts); i++ {
		mutable := []rune(parts[i])
		mutable[0] = unicode.ToUpper(mutable[0])
		fixed[i] = string(mutable)
	}
	return strings.Join(fixed, "")
}
