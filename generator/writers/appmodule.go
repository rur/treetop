package writers

import (
	"os"
	"path/filepath"
)

func WriteServerFile(dir string) (string, error) {
	fileName := "server.go"
	filePath := filepath.Join(dir, "server.go")
	sf, err := os.Create(filePath)
	if err != nil {
		return fileName, err
	}
	defer sf.Close()

	err = serverTemplate.Execute(sf, nil)
	if err != nil {
		return fileName, err
	}

	return fileName, nil
}

func WriteResourcesFile(dir string) (string, error) {
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
