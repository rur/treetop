package treetop

import (
	"reflect"
	"strings"
	"testing"
)

func basePartial() *partialDefImpl {
	return &partialDefImpl{
		template: "base.templ.html",
		extends:  nil,
		handler:  Noop,
		blocks:   []*blockDefImpl{},
	}
}

func Test_define_basic_block(t *testing.T) {
	part := basePartial()

	_ = part.Block("test")

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

	_ = part.Block("test")
	_ = part.Block("test")

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
	b := part.Block("test")
	p := b.Extend("test.templ.html", Noop)

	got := PrintHandler(p.FragmentHandler())
	expecting := `FragmentHandler("test.templ.html", github.com/rur/treetop.Noop)`
	if !strings.Contains(got, expecting) {
		t.Errorf("Extended template, expecting: %s, got %s", expecting, got)
	}
}

func Test_extend_block_partial(t *testing.T) {
	part := basePartial()
	b := part.Block("test")
	p := b.Extend("test.templ.html", Noop)

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
	part := basePartial()
	b := part.Block("test")
	_ = b.Default("test.templ.html", Noop)

	b2 := part.Block("test2")
	p2 := b2.Extend("test2.templ.html", Noop)

	a := p2.Block("A")
	_ = a.Default("deflt_A.templ.html", Noop)

	a_b := p2.Block("B")
	b2_bp := a_b.Extend("ext_B.templ.html", Noop)

	b2_bp.Block("B_plus").Default("ext_B_plus.templ.html", Noop)

	handler := b2_bp.PartialHandler()

	var expectingTempl []string
	expectingTempl = []string{"ext_B.templ.html", "ext_B_plus.templ.html"}
	if templates, err := handler.Partial.TemplateList(); err != nil {
		t.Errorf("Failed to load partial template, error: %s", err.Error())
	} else if !reflect.DeepEqual(templates, expectingTempl) {
		t.Errorf("Failed to list templates from partial, expecting: %s got %s", expectingTempl, templates)
	}

	expectingTempl = []string{
		"base.templ.html",
		"test.templ.html",
		"test2.templ.html",
		"deflt_A.templ.html",
		"ext_B.templ.html",
		"ext_B_plus.templ.html",
	}
	if templates, err := handler.Page.TemplateList(); err != nil {
		t.Errorf("Failed to load page template, error: %s", err.Error())
	} else if !reflect.DeepEqual(templates, expectingTempl) {
		t.Errorf("Failed to list templates from root, expecting: %s got %s", expectingTempl, templates)
	}
}
