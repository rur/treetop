package treetop

import (
	"reflect"
	"testing"
)

func Test_renderer_new_page(t *testing.T) {
	renderer := NewRenderer(DefaultTemplateExec)

	page := renderer.NewPageView("base.templ.html", Noop)

	page.DefaultSubView("A", "a.templ.html", Noop)
	def := page.SubView("B", "b.templ.html", Noop)

	handler := def.PartialHandler()

	expects := []string{"b.templ.html"}
	if files, err := handler.Partial.TemplateList(); err != nil {
		t.Errorf("treetop.Define() Partial = unexpected error %s", err.Error())
	} else if !reflect.DeepEqual(files, expects) {
		t.Errorf("treetop.Define() = %v, want %v", files, expects)
	}

	expects = []string{"base.templ.html", "a.templ.html", "b.templ.html"}
	if files, err := handler.Page.TemplateList(); err != nil {
		t.Errorf("treetop.Define() Page = unexpected error %s", err.Error())
	} else if !reflect.DeepEqual(files, expects) {
		t.Errorf("treetop.Define() = %v, want %v", files, expects)
	}
}
