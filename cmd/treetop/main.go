package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	generator "github.com/rur/treetop/generator"
)

var generateUsage = `
Usage: treetop generate site.yml [FLAGS...]
Create a temporary directory and generate templates and server code for given a site map.
By default the path to the new directory will be printed to stdout.

FLAGS:
--human	Human readable output

`

func main() {

	if len(os.Args) < 2 {
		log.Fatalln("Usage: treetop [cmd] args... \n")
		return
	}
	if os.Args[1] == "generate" {
		if len(os.Args) < 3 {
			log.Fatalln(generateUsage)
			return
		}
		config := os.Args[2]

		data, err := ioutil.ReadFile(config)
		if err != nil {
			log.Fatalf("Error loading YAML config: %v", err)
		}
		defs, err := generator.LoadPartialDef(data)
		if err != nil {
			log.Fatalf("Error parsing config: %v", err)
		}

		human := false
		for _, arg := range os.Args[3:] {
			if arg == "--human" {
				human = true
			} else {
				log.Fatalf("Unknown flag '%s'\n\n%s", arg, generateUsage)
			}
		}

		outfolder, createdFiles, err := generate(defs)
		if err != nil {
			log.Fatalf("Treetop: filed to generate data from sitemap %s\n%s\n", config, err.Error())
		}
		if human {
			fmt.Printf("Generated Treetop file in folder: %s\n\nTemplates:\n\t%s\n", outfolder, strings.Join(createdFiles, "\n\t"))
		} else {
			fmt.Print(outfolder)
		}

	} else {
		log.Fatalf("Treetop: unknown command %s\n\n", os.Args[1])
	}
}

func generate(defs []generator.PartialDef) (string, []string, error) {
	created := make([]string, 0)
	outDir, err := ioutil.TempDir("", "")
	if err != nil {
		return outDir, created, err
	}

	templatesDir := filepath.Join(outDir, "templates")
	if err := os.Mkdir(templatesDir, os.ModePerm); err != nil {
		return outDir, created, err
	}

	files, err := generator.CreateTemplateFiles(templatesDir, defs)
	if err != nil {
		return outDir, created, err
	}
	created = append(created, prependEach(files, "templates")...)

	files, err = generator.CreateServer(templatesDir, defs)
	if err != nil {
		return outDir, created, err
	}
	created = append(created, prependEach(files, "server")...)

	files, err = generator.CreateHandlers(templatesDir, defs)
	if err != nil {
		return outDir, created, err
	}
	created = append(created, prependEach(files, "server")...)
	return outDir, created, nil
}

func prependEach(files []string, prefix string) []string {
	out := make([]string, len(files))
	for i := 0; i < len(files); i++ {
		out[i] = filepath.Join(prefix, files[i])
	}
	return out
}
