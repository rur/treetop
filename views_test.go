package treetop

import (
	"fmt"
	"reflect"
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
	sub := base.NewSubView("sub-block", "sub.html", Constant("sub!"))

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
	base.NewDefaultSubView("sub-block", "sub.html", Constant("sub!"))

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
	base.NewDefaultSubView("sub-block", "sub.html", Constant("sub!"))
	copy := base.Copy()
	copy.Template = "copy.html"
	copy.HandlerFunc = Constant("copy!")
	copy.NewDefaultSubView("sub-block", "subCopy.html", Constant("subCopy!"))

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
	base := NewView("base.html", Constant("github.com/rur/treetop.Constant.func1"))
	sub := base.NewSubView("test", "sub.html", Constant("sub!"))
	copy := sub.Copy()
	if copy.Defines != "test" {
		t.Errorf("Expecting copy to have defines of 'test', got %s", copy.Defines)
	}
	err := assertViewDetails(copy.Parent, "base.html", "github.com/rur/treetop.Constant.func1")
	if err != nil {
		t.Error(err)
	}
}

func TestCompileViews(t *testing.T) {
	type TestCase struct {
		name           string
		view           *View
		includes       []*View
		expectPage     string
		expectView     string
		expectIncludes []string
	}
	cases := []TestCase{
		TestCase{
			name:           "basic",
			view:           NewView("base.html", Constant("github.com/rur/treetop.Constant.func1")),
			expectPage:     `- View("base.html", github.com/rur/treetop.Constant.func1)`,
			expectView:     `- View("base.html", github.com/rur/treetop.Constant.func1)`,
			expectIncludes: []string{},
		},
		TestCase{
			name: "with includes",
			view: NewView("base.html", Constant("github.com/rur/treetop.Constant.func1")),
			includes: []*View{
				NewView("other.html", Constant("other!")),
			},
			expectPage: `- View("base.html", github.com/rur/treetop.Constant.func1)`,
			expectView: `- View("base.html", github.com/rur/treetop.Constant.func1)`,
			expectIncludes: []string{
				`- View("other.html", github.com/rur/treetop.Constant.func1)`,
			},
		},
	}
	for _, tCase := range cases {
		t.Run(tCase.name, func(tt *testing.T) {
			page, view, incl := CompileViews(tCase.view, tCase.includes...)
			pageStr := SprintViewTree(page)
			if tCase.expectPage != pageStr {
				tt.Errorf("Expecting page %s, got %s", tCase.expectPage, pageStr)
				return
			}
			viewStr := SprintViewTree(view)
			if tCase.expectView != viewStr {
				tt.Errorf("Expecting view %s, got %s", tCase.expectView, viewStr)
				return
			}
			inclS := make([]string, len(incl))
			for i, inc := range incl {
				inclS[i] = SprintViewTree(inc)
			}
			if !reflect.DeepEqual(tCase.expectIncludes, inclS) {
				tt.Errorf("Expecting includes:\n%v\nGot:\n%v", tCase.expectIncludes, inclS)
				return
			}
		})
	}
}

// ---------
// Helpers
// ---------

// assertViewDetails is used for asserting that a view matches an expected template and
// data handler return value. This is for tests because it is expecting that the view handler
// will return string and not require use a the request object to do so.
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
