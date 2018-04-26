package generator

import (
	"gopkg.in/yaml.v2"
)

type PartialDef struct {
	Name     string                  `json:"name"`
	Default  bool                    `json:"default"`
	Path     string                  `json:"path"`
	Template string                  `json:"template"`
	Blocks   map[string][]PartialDef `json:"blocks"`
}

func LoadPartialDef(data []byte) ([]PartialDef, error) {
	var defs []PartialDef
	if err := yaml.Unmarshal(data, &defs); err != nil {
		return defs, err
	}
	return defs, nil
}
