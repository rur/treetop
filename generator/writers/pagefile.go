package writers

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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

type pageTemplateData struct {
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
	pageName := strings.Trim(pageDef.Name, " ")
	var nsReg = regexp.MustCompile(`(?i)^[A-Z][A-Z0-9-_]*$`)
	if !nsReg.MatchString(pageName) {
		return "", fmt.Errorf("Invalid page name '%s'.", pageName)
	}

	fileName := filepath.Join("pages", pageName, "page.go")
	filePath := filepath.Join(dir, "page.go")
	sf, err := os.Create(filePath)
	if err != nil {
		return fileName, err
	}
	defer sf.Close()

	page := pageData{
		Namespace: namespace,
		Name:      pageName,
		Template:  filepath.Join("pages", pageName, "templates", "index.html"),
		Handler:   pageName + "Handler",
		Blocks:    []pageBlockData{},
		Entries:   []pageEntryData{},
		Routes:    []pageRouteData{},
	}

	err = pageGoTemplate.Execute(sf, page)
	if err != nil {
		return fileName, err
	}

	return fileName, nil
}
