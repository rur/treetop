package generator

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"unicode"
)

type uniqueIdentifiers struct {
	ref chan map[string]bool
}

func newIdentifiers() uniqueIdentifiers {
	idents := uniqueIdentifiers{make(chan map[string]bool, 1)}
	idents.ref <- make(map[string]bool)
	return idents
}

// pretty lo-fi slugify, remove all non-alphanum and lowercase first character
func namelike(name string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9-]+")
	if err != nil {
		log.Fatal(err)
	}
	stripped := reg.ReplaceAllString(name, "")
	mutable := []rune(stripped)
	mutable[0] = unicode.ToLower(mutable[0])
	return string(mutable)
}

// convert a name to a valid identifier, leading digits will be stripped
func validIdentifier(name string) string {
	nme := namelike(name)
	parts := strings.Split(nme, "-")
	fixed := make([]string, len(parts))

	word := []rune(parts[0])
	for {
		if len(word) == 0 {
			fixed[0] = "var"
			break
		} else if unicode.IsDigit(word[0]) {
			word = word[1:]
		} else {
			word[0] = unicode.ToLower(word[0])
			fixed[0] = string(word)
			break
		}
	}

	for i := 1; i < len(parts); i++ {
		mutable := []rune(parts[i])
		mutable[0] = unicode.ToUpper(mutable[0])
		fixed[i] = string(mutable)
	}
	return strings.Join(fixed, "")
}

// create a new identifier that is unique relative to uI instance
func (u *uniqueIdentifiers) new(name, qualifier string) string {
	var found bool
	// this is a statefull method so I'm using a channel as a locking mechanism
	ref := <-u.ref
	defer func() {
		u.ref <- ref
	}()

	ident := validIdentifier(name)

	if _, found = ref[ident]; !found {
		ref[ident] = true
		return ident
	}
	identQlf := ident + strings.Title(qualifier)
	if _, found = ref[identQlf]; !found {
		ref[identQlf] = true
		return identQlf
	}
	i := 1
	var identI string
	for {
		identI = fmt.Sprintf("%s%v", identQlf, i)
		if _, found = ref[identI]; !found {
			ref[identI] = true
			return identI
		}
		i += 1
	}
}
