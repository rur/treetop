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
	writers "github.com/rur/treetop/generator/writers"
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
			fmt.Printf("Error loading sitemap file: %v", err)
			return
		}
		sitemap, err := generator.LoadSitemap(data)
		if err != nil {
			fmt.Printf("Error parsing sitemap YAML: %v", err)
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

		outfolder, createdFiles, err := generate(sitemap)
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

func generate(sitemap generator.Sitemap) (string, []string, error) {
	var files []string
	var file string
	created := make([]string, 0)
	outDir, err := ioutil.TempDir("", "")
	if err != nil {
		return outDir, created, err
	}

	pagesDir := filepath.Join(outDir, "pages")
	if err := os.Mkdir(pagesDir, os.ModePerm); err != nil {
		return outDir, created, err
	}

	for _, def := range sitemap.Pages {
		pageDir := filepath.Join(pagesDir, def.Name)
		if err := os.Mkdir(pageDir, os.ModePerm); err != nil {
			return outDir, created, err
		}
		templatesDir := filepath.Join(pageDir, "templates")
		if err := os.Mkdir(templatesDir, os.ModePerm); err != nil {
			return outDir, created, err
		}

		file, err = writers.WritePageFile(pageDir, &def, sitemap.Namespace)
		if err != nil {
			return outDir, created, err
		}
		created = append(created, file)
		files, err = writers.WriteHandlerFile(pageDir, &def, sitemap.Namespace)
		if err != nil {
			return outDir, created, err
		}
		created = append(created, file)
		files, err = writers.WriteTemplateFiles(templatesDir, &def)
		if err != nil {
			return outDir, created, err
		}
		created = append(created, files...)
	}

	file, err = writers.WriteContext(pagesDir)
	if err != nil {
		return outDir, created, err
	}
	created = append(created, file)

	file, err = writers.WriteStartFile(outDir, defs)
	if err != nil {
		return outDir, created, err
	}
	created = append(created, file)

	return outDir, created, nil
}
