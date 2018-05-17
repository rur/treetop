package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
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
		fmt.Printf("Usage: treetop [cmd] args... \n")
		return
	}
	if os.Args[1] == "generate" {
		if len(os.Args) < 3 {
			fmt.Printf(generateUsage)
			return
		}
		config := os.Args[2]

		data, err := ioutil.ReadFile(config)
		if err != nil {
			fmt.Printf("Error loading YAML config: %v", err)
			return
		}
		defs, err := generator.LoadPartialDef(data)
		if err != nil {
			fmt.Printf("Error parsing config: %v", err)
			return
		}

		human := false
		for _, arg := range os.Args[3:] {
			if arg == "--human" {
				human = true
			} else {
				fmt.Printf("Unknown flag '%s'\n\n%s", arg, generateUsage)
				return
			}
		}

		outfolder, createdFiles, err := generate(defs)
		if err != nil {
			fmt.Printf("Treetop: filed to generate data from sitemap %s\n%s\n", config, err.Error())
			return
		} else {
			// attempt to format the go code
			// this should not cause the generate command to fail if go fmt fails for some reason
			for i := range createdFiles {
				if strings.HasSuffix(createdFiles[i], ".go") {
					cmd := exec.Command("go", "fmt", path.Join(outfolder, createdFiles[i]))
					err := cmd.Run()
					if err != nil {
						log.Fatal(err)
					}
				}
			}

		}

		if human {
			fmt.Printf("Generated Treetop file in folder: %s\n\nFiles:\n\t%s\n", outfolder, strings.Join(createdFiles, "\n\t"))
		} else {
			fmt.Print(outfolder)
		}

	} else {
		fmt.Printf("Treetop: unknown command %s\n\n", os.Args[1])
		return
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

	serverDir := filepath.Join(outDir, "server")
	if err := os.Mkdir(serverDir, os.ModePerm); err != nil {
		return outDir, created, err
	}

	files, err := generator.CreateSeverFiles(outDir, defs)
	if err != nil {
		return outDir, created, err
	}
	created = append(created, files...)

	return outDir, created, nil
}
