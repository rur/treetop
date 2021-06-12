package internal

import (
	"fmt"
)

func LoadStringTemplate(tmpl string) (string, error) {
	return tmpl, nil
}

type KeyedString map[string]string

func (ks KeyedString) LoadTemplate(key string) (string, error) {
	tmpl, ok := ks[key]
	if !ok {
		return "", fmt.Errorf("no key found for template '%s'", key)
	}
	return tmpl, nil
}

// func LoadTemplateFile(name string) (string, error) {
// 	tmpl, ok := ks[key]
// 	if !ok {
// 		return "", fmt.Errorf("no key found for template '%s'", key)
// 	}
// 	return tmpl, nil
// }

// // readStringAndClose ensures that the supplied read closer is closed
// func readStringAndClose(buffer *bytes.Buffer, rc io.ReadCloser) (string, error) {
// 	defer rc.Close()
// 	_, err := buffer.ReadFrom(rc)
// 	if err != nil {
// 		return "", err
// 	}
// 	return buffer.String(), nil
// }
