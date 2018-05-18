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
func snakify(name string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9-_]+")
	if err != nil {
		log.Fatal(err)
	}
	return strings.ToLower(reg.ReplaceAllString(name, "-"))
}

func delims(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsNumber(r)
}

// convert a name to a valid identifier, leading digits will be stripped
func validIdentifier(name string) string {
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
func validPublicIdentifier(name string) string {
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

// create a new identifier that is unique relative to u instance
func (u *uniqueIdentifiers) new(name string, qualifier []string) string {
	var found bool
	// this is a statefull method so I'm using a channel as a locking mechanism
	ref := <-u.ref
	defer func() {
		u.ref <- ref
	}()

	ident := validIdentifier(name)

	if len(qualifier) > 0 {
		ident = validIdentifier(strings.Join(qualifier, " ")) + "_" + strings.Title(ident)
	}

	if _, found = ref[ident]; !found {
		ref[ident] = true
		return ident
	}
	i := 1
	var identI string
	for {
		identI = fmt.Sprintf("%s%v", ident, i)
		if _, found = ref[identI]; !found {
			ref[identI] = true
			return identI
		}
		i += 1
	}
}

func (u *uniqueIdentifiers) copy() *uniqueIdentifiers {
	ref := <-u.ref
	defer func() {
		u.ref <- ref
	}()
	newIdent := newIdentifiers()
	newRef := <-newIdent.ref
	for k, v := range ref {
		newRef[k] = v
	}
	newIdent.ref <- newRef
	return &newIdent
}

func (u *uniqueIdentifiers) reserve(ident string) string {
	ref := <-u.ref
	defer func() {
		u.ref <- ref
	}()
	ref[ident] = true
	return ident
}

func (u *uniqueIdentifiers) exists(ident string) bool {
	// this is a statefull method so I'm using a channel as a locking mechanism
	ref := <-u.ref
	defer func() {
		u.ref <- ref
	}()
	_, found := ref[ident]
	return found
}
