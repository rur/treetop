package treetop

import (
	"reflect"
	"strings"
	"testing"
)

func basePartial() *viewImpl {
	return &viewImpl{
		template: "base.html.tmpl",
		extends:  nil,
		handler:  Noop,
		blocks:   []*blockImpl{},
	}
}

func Test_define_basic_block(t *testing.T) {
	part := basePartial()

	_ = part.SubView("test", "test.html", Noop)

	if len(part.blocks) != 1 {
		t.Errorf("New block should have been added to the list of blocks")
		return
	}
	impl := part.blocks[0]
	if impl.parent.template != part.template {
		t.Errorf("Expected new block to refer back to the partial that defined it %v", impl)
	}
	if impl.name != "test" {
		t.Errorf("Expected new blockname: %#v got %#v", "test", impl.name)
	}
}
func Test_retrieve_an_existing_block(t *testing.T) {
	// calling block with the same name should return an instance of the same block
	part := basePartial()

	_ = part.SubView("test", "test.html", Noop)
	_ = part.SubView("test", "test.html", Noop)

	if len(part.blocks) != 1 {
		t.Errorf("Only one new block should have been created")
		return
	}
	impl := part.blocks[0]
	if impl.parent.template != part.template {
		t.Errorf("Expected new block to refer back to the partial that defined it %v", impl)
	}
	if impl.name != "test" {
		t.Errorf("Expected new blockname: %#v got %#v", "test", impl.name)
	}
}

func Test_extend_block_basic(t *testing.T) {
	part := basePartial()
	p := part.SubView("test", "test.html.tmpl", Noop)

	got := PrintHandler(p.Handler().FragmentOnly())
	expecting := `FragmentHandler("test.html.tmpl", github.com/rur/treetop.Noop)`
	if !strings.Contains(got, expecting) {
		t.Errorf("Extended template, expecting: %s, got %s", expecting, got)
	}
}

func Test_fragment_with_blocks(t *testing.T) {
	part := basePartial()
	p := part.SubView("test", "test.html.tmpl", Noop)
	p.DefaultSubView("sub", "sub.html.tmpl", Noop)

	got, err := p.Handler().FragmentOnly().Fragment.TemplateList()
	if err != nil {
		t.Errorf("Unexpected error while aggregating templates: %s", err.Error())
	}
	expecting := []string{"test.html.tmpl", "sub.html.tmpl"}
	if !reflect.DeepEqual(got, expecting) {
		t.Errorf("Extended template, expecting: %#v, got %#v", expecting, got)
	}
}

func Test_extend_block_partial(t *testing.T) {
	part := basePartial()
	p := part.SubView("test", "test.html.tmpl", Noop)

	handler := p.Handler()
	expecting := `PartialHandler("test.html.tmpl", github.com/rur/treetop.Noop)`
	got := PrintHandler(handler)

	if !strings.Contains(got, expecting) {
		t.Errorf("Extended template, expecting: %s, got %s", expecting, got)
	}

	expectingTempl := []string{"base.html.tmpl", "test.html.tmpl"}
	if templates, err := handler.Page.TemplateList(); err != nil {
		t.Errorf("Failed to load page template, error: %s", err.Error())
	} else if !reflect.DeepEqual(templates, expectingTempl) {
		t.Errorf("Failed to list templates from root, expecting: %s got %s", expectingTempl, templates)
	}
}

func Test_extend_multiple_levels(t *testing.T) {
	base := basePartial()

	_ = base.DefaultSubView("test", "test.html.tmpl", Noop)
	test2 := base.SubView("test2", "test2.html.tmpl", Noop)
	_ = test2.DefaultSubView("A", "default_A.html.tmpl", Noop)
	test2B := test2.SubView("B", "test2_B.html.tmpl", Noop)
	test2B.DefaultSubView("B_plus", "test2_B_plus.html.tmpl", Noop)

	handler := test2B.Handler()

	var expectingTempl []string
	expectingTempl = []string{"test2_B.html.tmpl", "test2_B_plus.html.tmpl"}
	if templates, err := handler.Fragment.TemplateList(); err != nil {
		t.Errorf("Failed to load partial template, error: %s", err.Error())
	} else if !reflect.DeepEqual(templates, expectingTempl) {
		t.Errorf("Failed to list templates from partial, expecting: %s got %s", expectingTempl, templates)
	}

	expectingTempl = []string{
		"base.html.tmpl",
		"test.html.tmpl",
		"test2.html.tmpl",
		"default_A.html.tmpl",
		"test2_B.html.tmpl",
		"test2_B_plus.html.tmpl",
	}
	if templates, err := handler.Page.TemplateList(); err != nil {
		t.Errorf("Failed to load page template, error: %s", err.Error())
	} else if !reflect.DeepEqual(templates, expectingTempl) {
		t.Errorf("Failed to list templates from root, expecting: %s got %s", expectingTempl, templates)
	}
}
