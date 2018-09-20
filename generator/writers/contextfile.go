package writers

import (
	"os"
	"path/filepath"
)

type contextdata struct {
	Namespace string
}

func WriteContextFile(dir string) (string, error) {
	fileName := "context.go"
	filePath := filepath.Join(dir, "context.go")
	sf, err := os.Create(filePath)
	if err != nil {
		return fileName, err
	}
	defer sf.Close()

	err = contextTemplate.Execute(sf, nil)
	if err != nil {
		return fileName, err
	}

	return fileName, nil
}
