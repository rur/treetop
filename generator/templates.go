package generator

import (
	"html/template"
	templ "html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	partialTemplate *templ.Template
)

var partial = `
{{ block "[[ .Extends ]]" . }}
<div id="[[ .Extends ]]" style="background-color: rgba(0, 0, 0, 0.1)">
    <p>View named [[.Name]]</p>
[[ range $blockName, $partials := .Blocks ]]
    <ul>
    [[ range $index, $def := $partials ]]
        <li style="display: inline;"><a href="[[ $def.Path ]]" treetop>[[ $def.Name ]]</a></li>
    [[ end ]]
    </ul>
    {{ block "[[ $blockName ]]" . }}
    <div id="[[ $blockName  ]]" style="background-color: lightsalmon;">
        <p>default for block named [[$blockName ]]</p>
    </div>
    {{ end }}
[[ end ]]
</div>
{{ end }}
`

func init() {
	var err error
	partialTemplate, err = template.New("partial").Delims("[[", "]]").Parse(partial)
	if err != nil {
		log.Fatal(err)
	}
}

func CreateTemplateFiles(dir string, defs []PartialDef) ([]string, error) {
	created := make([]string, 0, len(defs))
	for _, def := range defs {
		files, err := handleDef(def, "", dir, "root")
		if err != nil {
			return created, err
		}
		created = append(created, files...)
	}
	return created, nil
}

func handleDef(def PartialDef, prefix string, folder string, extends string) ([]string, error) {
	created := make([]string, 0, 1+len(def.Blocks))
	var filename string
	if prefix != "" {
		filename = strings.Join([]string{prefix, def.Name}, "_")
	} else {
		filename = def.Name
	}

	for blockName, defs := range def.Blocks {
		for _, d := range defs {
			files, err := handleDef(d, filename, folder, blockName)
			if err != nil {
				return created, err
			}
			created = append(created, files...)
		}
	}

	filename = filename + ".templ.html"
	path := filepath.Join(folder, filename)
	f, err := os.Create(path)
	if err != nil {
		return created, err
	}
	created = append(created, filename)
	defer f.Close()
	err = partialTemplate.Execute(f, struct {
		Name     string
		Default  bool
		Path     string
		Template string
		Blocks   map[string][]PartialDef
		Extends  string
	}{
		Name:     def.Name,
		Default:  def.Default,
		Path:     def.Path,
		Template: def.Template,
		Blocks:   def.Blocks,
		Extends:  extends,
	})
	if err != nil {
		return created, err
	}
	return created, nil
}
