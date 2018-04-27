package generator

import (
	"gopkg.in/yaml.v2"
)

type PartialDef struct {
	Name     string                  `yaml:"name"`
	Default  bool                    `yaml:"default"`
	Path     string                  `yaml:"path"`
	Template string                  `yaml:"template"`
	Blocks   map[string][]PartialDef `yaml:"blocks"`
}

func LoadPartialDef(data []byte) ([]PartialDef, error) {
	var defs []PartialDef
	if err := yaml.Unmarshal(data, &defs); err != nil {
		return defs, err
	}
	return defs, nil
}
