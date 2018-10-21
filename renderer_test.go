package treetop

import (
	"reflect"
	"testing"
)

func Test_renderer_Define(t *testing.T) {
	renderer := NewRenderer(DefaultTemplateExec)

	page := renderer.Define("base.templ.html", Noop)

	a := page.Block("A")
	b := page.Block("B")

	a.Default("a.templ.html", Noop)
	def := b.Extend("b.templ.html", Noop)

	handler := def.PartialHandler()

	expects := []string{"b.templ.html"}
	if files, err := handler.Partial.TemplateList(); err != nil {
		t.Errorf("renderer.Define() Partial = unexpected error %s", err.Error())
	} else if !reflect.DeepEqual(files, expects) {
		t.Errorf("renderer.Define() = %v, want %v", files, expects)
	}

	expects = []string{"base.templ.html", "a.templ.html", "b.templ.html"}
	if files, err := handler.Page.TemplateList(); err != nil {
		t.Errorf("renderer.Define() Page = unexpected error %s", err.Error())
	} else if !reflect.DeepEqual(files, expects) {
		t.Errorf("renderer.Define() = %v, want %v", files, expects)
	}
}
