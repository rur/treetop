package generator

import (
	"gopkg.in/yaml.v2"
)

type Sitemap struct {
	Namespace string       `yaml:"namespace"`
	Pages     []PartialDef `yaml:"pages"`
}

type PartialDef struct {
	Name     string                  `yaml:"name"`     // The unique name for this view
	Fragment bool                    `yaml:"fragment"` // Is this a 'FragmetOnly' route
	Default  bool                    `yaml:"default"`  // Is this a default subview
	Path     string                  `yaml:"path"`     // determine if a HTTP route should be associated with this view
	Includes []string                `yaml:"includes"` // list of other partails that should be included in the route
	Handler  string                  `yaml:"handler"`  // explicit handler declaration
	Template string                  `yaml:"template"` // explicit template path
	Merge    string                  `yaml:"merge"`    // treetop-merge attribute of a partial root element
	Method   string                  `yaml:"method"`   // HTTP request method for a route, default "GET"
	Doc      string                  `yaml:"doc"`      // Optional doc string to include with the generated handler
	Blocks   map[string][]PartialDef `yaml:"blocks"`   // List of subviews from this view
	URI      string                  `yaml:"uri"`      // the entrypoint URL for the top level view
}

func LoadSitemap(data []byte) (Sitemap, error) {
	var config Sitemap
	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, err
	}
	return config, nil
}
