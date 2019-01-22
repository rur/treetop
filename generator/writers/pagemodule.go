package writers

import (
	"os"
	"path/filepath"
)

type contextsdata struct {
	Namespace string
}

func WriteContextFile(dir string, namespace string) (string, error) {
	fileName := "context.go"
	filePath := filepath.Join(dir, "context.go")
	sf, err := os.Create(filePath)
	if err != nil {
		return fileName, err
	}
	defer sf.Close()

	err = contextTemplate.Execute(sf, contextsdata{namespace})
	if err != nil {
		return fileName, err
	}

	return fileName, nil
}

func WriteMuxFile(dir string) (string, error) {
	fileName := "mux.go"
	filePath := filepath.Join(dir, "mux.go")
	sf, err := os.Create(filePath)
	if err != nil {
		return fileName, err
	}
	defer sf.Close()

	err = muxTemplate.Execute(sf, nil)
	if err != nil {
		return fileName, err
	}

	return fileName, nil
}
