package treetop

import (
	"reflect"
	"strings"
	"testing"
)

func basePartial() *viewImpl {
	return &viewImpl{
		template: "base.templ.html",
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
	p := part.SubView("test", "test.templ.html", Noop)

	got := PrintHandler(p.FragmentHandler())
	expecting := `FragmentHandler("test.templ.html", github.com/rur/treetop.Noop)`
	if !strings.Contains(got, expecting) {
		t.Errorf("Extended template, expecting: %s, got %s", expecting, got)
	}
}

func Test_fragment_with_blocks(t *testing.T) {
	part := basePartial()
	p := part.SubView("test", "test.templ.html", Noop)
	p.DefaultSubView("sub", "sub.templ.html", Noop)

	got, err := p.FragmentHandler().Partial.TemplateList()
	if err != nil {
		t.Errorf("Unexpected error while aggregating templates: %s", err.Error())
	}
	expecting := []string{"test.templ.html", "sub.templ.html"}
	if !reflect.DeepEqual(got, expecting) {
		t.Errorf("Extended template, expecting: %#v, got %#v", expecting, got)
	}
}

func Test_extend_block_partial(t *testing.T) {
	part := basePartial()
	p := part.SubView("test", "test.templ.html", Noop)

	handler := p.PartialHandler()
	expecting := `PartialHandler("test.templ.html", github.com/rur/treetop.Noop)`
	got := PrintHandler(handler)

	if !strings.Contains(got, expecting) {
		t.Errorf("Extended template, expecting: %s, got %s", expecting, got)
	}

	expectingTempl := []string{"base.templ.html", "test.templ.html"}
	if templates, err := handler.Page.TemplateList(); err != nil {
		t.Errorf("Failed to load page template, error: %s", err.Error())
	} else if !reflect.DeepEqual(templates, expectingTempl) {
		t.Errorf("Failed to list templates from root, expecting: %s got %s", expectingTempl, templates)
	}
}

func Test_extend_multiple_levels(t *testing.T) {
	base := basePartial()
	_ = base.DefaultSubView("test", "test.templ.html", Noop)

	test2 := base.SubView("test2", "test2.templ.html", Noop)

	_ = test2.DefaultSubView("A", "default_A.templ.html", Noop)

	test2_b := test2.SubView("B", "test2_B.templ.html", Noop)

	test2_b.DefaultSubView("B_plus", "test2_B_plus.templ.html", Noop)

	handler := test2_b.PartialHandler()

	var expectingTempl []string
	expectingTempl = []string{"test2_B.templ.html", "test2_B_plus.templ.html"}
	if templates, err := handler.Partial.TemplateList(); err != nil {
		t.Errorf("Failed to load partial template, error: %s", err.Error())
	} else if !reflect.DeepEqual(templates, expectingTempl) {
		t.Errorf("Failed to list templates from partial, expecting: %s got %s", expectingTempl, templates)
	}

	expectingTempl = []string{
		"base.templ.html",
		"test.templ.html",
		"test2.templ.html",
		"default_A.templ.html",
		"test2_B.templ.html",
		"test2_B_plus.templ.html",
	}
	if templates, err := handler.Page.TemplateList(); err != nil {
		t.Errorf("Failed to load page template, error: %s", err.Error())
	} else if !reflect.DeepEqual(templates, expectingTempl) {
		t.Errorf("Failed to list templates from root, expecting: %s got %s", expectingTempl, templates)
	}
}
