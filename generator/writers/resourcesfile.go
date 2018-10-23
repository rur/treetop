package writers

import (
	"os"
	"path/filepath"
)

type resourcesdata struct {
	Namespace string
}

func WriteContextFile(dir string) (string, error) {
	fileName := "resources.go"
	filePath := filepath.Join(dir, "resources.go")
	sf, err := os.Create(filePath)
	if err != nil {
		return fileName, err
	}
	defer sf.Close()

	err = resourcesTemplate.Execute(sf, nil)
	if err != nil {
		return fileName, err
	}

	return fileName, nil
}
