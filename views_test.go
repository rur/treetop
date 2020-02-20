package treetop

import (
	"fmt"
	"testing"
)

func TestNewView(t *testing.T) {
	base := NewView("base.html", Constant("ok!"))
	err := assertViewDetails(base, "base.html", "ok!")
	if err != nil {
		t.Error(err)
	}
}

func TestSubView(t *testing.T) {
	base := NewView("base.html", Constant("base!"))
	sub := base.SubView("sub-block", "sub.html", Constant("sub!"))

	var err error

	err = assertViewDetails(sub, "sub.html", "sub!")
	if err != nil {
		t.Error(err)
	}
	err = assertViewDetails(sub.Parent, "base.html", "base!")
	if err != nil {
		t.Error(err)
	}
}

func TestDefaultSubView(t *testing.T) {
	base := NewView("base.html", Constant("base!"))
	base.DefaultSubView("sub-block", "sub.html", Constant("sub!"))

	sub, ok := base.SubViews["sub-block"]
	if !ok {
		t.Error("Failed to register block with base")
	}
	err := assertViewDetails(sub, "sub.html", "sub!")
	if err != nil {
		t.Error(err)
	}
}

func TestCopyView(t *testing.T) {
	// create a base view then copy it and make changes
	base := NewView("base.html", Constant("base!"))
	base.DefaultSubView("sub-block", "sub.html", Constant("sub!"))
	copy := base.Copy()
	copy.Template = "copy.html"
	copy.HandlerFunc = Constant("copy!")
	copy.DefaultSubView("sub-block", "subCopy.html", Constant("subCopy!"))

	// assert that changes to the copy did not affect the original view
	var err error
	err = assertViewDetails(base, "base.html", "base!")
	if err != nil {
		t.Error(err)
	}
	basesub, ok := base.SubViews["sub-block"]
	if !ok {
		t.Error("Failed to register block with base")
	}
	err = assertViewDetails(basesub, "sub.html", "sub!")
	if err != nil {
		t.Error(err)
	}

	err = assertViewDetails(copy, "copy.html", "copy!")
	if err != nil {
		t.Error(err)
	}
	copysub, ok := copy.SubViews["sub-block"]
	if !ok {
		t.Error("Failed to register block with copy")
	}
	err = assertViewDetails(copysub, "subCopy.html", "subCopy!")
	if err != nil {
		t.Error(err)
	}
}

func TestCopySubView(t *testing.T) {
	base := NewView("base.html", Constant("base!"))
	sub := base.SubView("test", "sub.html", Constant("sub!"))
	copy := sub.Copy()
	if copy.Defines != "test" {
		t.Errorf("Expecting copy to have defines of 'test', got %s", copy.Defines)
	}
	err := assertViewDetails(copy.Parent, "base.html", "base!")
	if err != nil {
		t.Error(err)
	}
}

// ---------
// Helpers
// ---------

func assertViewDetails(v *View, t string, data string) error {
	if v.Template != t {
		return fmt.Errorf("expecting template %s got %s", t, v.Template)
	}

	switch got := v.HandlerFunc(&ResponseWrapper{}, nil).(type) {
	case string:
		if got != data {
			return fmt.Errorf("expecting %s got %s", data, got)
		}
		return nil
	default:
		return fmt.Errorf("unexpected return value from base handler %v", got)
	}
}
