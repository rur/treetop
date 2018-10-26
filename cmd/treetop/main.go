package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
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
--temp-dir [DIRECTORY_PATH]	Path to directory that should be used as 'temp'

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
		skip := 0
		tmpDir := ""
		for i, arg := range os.Args[3:] {
			if skip > 0 {
				skip = skip - 1
				continue
			} else if arg == "--human" {
				human = true
			} else if arg == "--temp-dir" {
				tmpDir = os.Args[i+4]
				skip = 1
			} else {
				fmt.Printf("Unknown flag '%s'\n\n%s", arg, generateUsage)
				return
			}
		}

		outfolder, err := ioutil.TempDir(tmpDir, "")
		if err != nil {
			fmt.Printf("Error creating temp dir: %s", err)
			return
		}

		createdFiles, err := generate(outfolder, sitemap)
		if err != nil {
			fmt.Printf("Treetop: Failed to build scaffold for sitemap %s\nGenerator Error: %s\n", config, err.Error())
			if err := os.RemoveAll(outfolder); err != nil {
				fmt.Printf("Scaffold failed but temp directory was not cleaned up: %s\n", err.Error())
			}
			return
		} else {
			// attempt to format the go code
			// this should not cause the generate command to fail if go fmt fails for some reason
			var fmtError []string
			for i := range createdFiles {
				if strings.HasSuffix(createdFiles[i], ".go") {
					cmd := exec.Command("go", "fmt", path.Join(outfolder, createdFiles[i]))
					output, err := cmd.CombinedOutput()
					if err != nil {
						fmtError = append(fmtError, fmt.Sprintf("%s Error: %s\nOutput: %s", createdFiles[i], err, string(output)))
					}
				}
			}
			if len(fmtError) > 0 {
				log.Fatalf(
					"Generated folder %s but `go fmt` failed for the following files:\n\t%s",
					outfolder,
					strings.Join(fmtError, "\n\t"),
				)
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

func generate(outDir string, sitemap generator.Sitemap) ([]string, error) {
	var file string
	var err error
	created := make([]string, 0)

	// check that sitemap namespace is a uri looking thing (without protocol, creds, etc...)
	// it will typically be something like "github.com/example/project"
	var nsReg = regexp.MustCompile(`(?i)^[A-Z][A-Z0-9-_]*(\.[A-Z][A-Z0-9-_]*)*(/[A-Z][A-Z0-9-_]*(\.[A-Z][A-Z0-9-_]*)*)*$`)
	if !nsReg.MatchString(sitemap.Namespace) {
		return created, fmt.Errorf("Invalid site namespace in config: %s", sitemap.Namespace)
	}

	pageDir := filepath.Join(outDir, "pages")
	if err := os.Mkdir(pageDir, os.ModePerm); err != nil {
		return created, fmt.Errorf("Error creating 'page' dir in temp directory. %s", err)
	}

	for _, def := range sitemap.Pages {
		pageName, err := writers.SanitizeName(def.Name)
		if err != nil {
			return created, err
		}
		pageDir := filepath.Join(pageDir, pageName)
		if err := os.Mkdir(pageDir, os.ModePerm); err != nil {
			return created, fmt.Errorf("Error creating dir for page '%s'. %s", def.Name, err)
		}
		templatesDir := filepath.Join(pageDir, "templates")
		if err := os.Mkdir(templatesDir, os.ModePerm); err != nil {
			return created, fmt.Errorf("Error creating template dir for page '%s'. %s", def.Name, err)
		}

		file, err = writers.WritePageFile(pageDir, &def, sitemap.Namespace)
		if err != nil {
			return created, fmt.Errorf("Error creating page.go file for '%s'. %s", def.Name, err)
		}
		created = append(created, path.Join("pages", pageName, file))

		file, err = writers.WriteHandlerFile(pageDir, &def, sitemap.Namespace)
		if err != nil {
			return created, fmt.Errorf("Error creating handler.go file for page '%s'. %s", def.Name, err)
		}
		created = append(created, path.Join("pages", pageName, file))

		file, err = writers.WriteIndexFile(templatesDir, &def, sitemap.Pages)
		if err != nil {
			return created, fmt.Errorf("Error creating index.templ.html file for page '%s'. %s", def.Name, err)
		}
		created = append(created, path.Join("pages", pageName, "templates", file))

		files, err := writers.WriteTemplateBlock(templatesDir, def.Blocks)
		if err != nil {
			return created, fmt.Errorf("Error creating HTML partials for page '%s'. %s", def.Name, err)
		}
		for _, file = range files {
			created = append(created, path.Join("pages", pageName, "templates", file))
		}
	}

	file, err = writers.WriteContextFile(pageDir)
	if err != nil {
		return created, fmt.Errorf("Error creating context.go file. %s", err)
	}
	created = append(created, path.Join("pages", file))

	file, err = writers.WriteStartFile(outDir, sitemap.Pages, sitemap.Namespace)
	if err != nil {
		return created, fmt.Errorf("Error creating start.go file. %s", err)
	}
	created = append(created, file)

	return created, nil
}
