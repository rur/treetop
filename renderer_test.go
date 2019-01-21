package treetop

import (
	"reflect"
	"testing"
)

func Test_renderer_new_page(t *testing.T) {
	renderer := NewRenderer(DefaultTemplateExec)

	page := renderer.NewView("base.html.tmpl", Noop)

	page.DefaultSubView("A", "a.html.tmpl", Noop)
	def := page.SubView("B", "b.html.tmpl", Noop)

	handler := def.Handler()

	expects := []string{"b.html.tmpl"}
	if files, err := handler.Fragment.TemplateList(); err != nil {
		t.Errorf("treetop.Define() Fragment = unexpected error %s", err.Error())
	} else if !reflect.DeepEqual(files, expects) {
		t.Errorf("treetop.Define() = %v, want %v", files, expects)
	}

	expects = []string{"base.html.tmpl", "a.html.tmpl", "b.html.tmpl"}
	if files, err := handler.Page.TemplateList(); err != nil {
		t.Errorf("treetop.Define() Page = unexpected error %s", err.Error())
	} else if !reflect.DeepEqual(files, expects) {
		t.Errorf("treetop.Define() = %v, want %v", files, expects)
	}
}
