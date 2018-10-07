package treetop

import (
	"testing"
)

func Test_define_basic_block(t *testing.T) {
	part := &partialDefImpl{
		template: "base.templ.html",
		extends:  nil,
		handler:  Noop,
		blocks:   []*blockDefImpl{},
	}

	_ = part.Block("test")

	if len(part.blocks) != 1 {
		t.Errorf("New block should have been added to the list of blocks")
		return
	}
	impl := part.blocks[0]
	if impl.parent.template != part.template {
		t.Errorf("Expected new block to refer back to the partial that defined it %v", block)
	}
	if impl.name != "test" {
		t.Errorf("Expected new blockname: %#v got %#v", "test", impl.name)
	}
}
func Test_retrieve_an_existing_block(t *testing.T) {
	// calling block with the same name should return an instance of the same block
	part := &partialDefImpl{
		template: "base.templ.html",
		extends:  nil,
		handler:  Noop,
		blocks:   []*blockDefImpl{},
	}

	_ = part.Block("test")
	_ = part.Block("test")

	if len(part.blocks) != 1 {
		t.Errorf("Only one new block should have been created")
		return
	}
	impl := part.blocks[0]
	if impl.parent.template != part.template {
		t.Errorf("Expected new block to refer back to the partial that defined it %v", block)
	}
	if impl.name != "test" {
		t.Errorf("Expected new blockname: %#v got %#v", "test", impl.name)
	}
}
