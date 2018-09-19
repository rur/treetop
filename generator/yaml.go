package generator

import (
	"gopkg.in/yaml.v2"
)

type Sitemap struct {
	Namespace string       `yaml:"namespace"`
	Pages     []PartialDef `yaml:"pages"`
}

type PartialDef struct {
	Name     string                  `yaml:"name"`
	Fragment bool                    `yaml:"fragment"`
	Default  bool                    `yaml:"default"`
	Path     string                  `yaml:"path"`
	Doc      string                  `yaml:"doc"`
	Blocks   map[string][]PartialDef `yaml:"blocks"`
	URI      string                  `yaml:"uri"`
}

func LoadSitemap(data []byte) (Sitemap, error) {
	var config Sitemap
	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, err
	}
	return config, nil
}
