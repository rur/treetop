package treetop

import (
	"fmt"
	"reflect"
	"strings"
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

func TestHasSubview(t *testing.T) {
	base := NewView("base.html", Constant("base!"))
	base.HasSubView("sub-block")

	sub, ok := base.SubViews["sub-block"]
	if !ok {
		t.Error("Failed to register block with base")
	}
	if sub != nil {
		t.Errorf("HasSubView unexpected subview: %v", sub)
	}
}

func TestHasSubviewNoEffectOnDefault(t *testing.T) {
	base := NewView("base.html", Constant("base!"))
	base.NewDefaultSubView("sub-block", "sub.html", Constant("sub!"))
	base.HasSubView("sub-block")

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
		{
			name:       "nil case",
			view:       nil,
			expectPage: "- nil",
			expectView: "- nil",
		},
		{
			name:           "basic",
			view:           NewView("base.html", Constant("github.com/rur/treetop.Constant.func1")),
			expectPage:     `- View("base.html", github.com/rur/treetop.Constant.func1)`,
			expectView:     `- View("base.html", github.com/rur/treetop.Constant.func1)`,
			expectIncludes: []string{},
		}, {
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
		}, {
			name: "with parent view",
			view: func() *View {
				base := NewView("base.html", Noop)
				return base.NewSubView("test", "test.html", Constant("test!"))
			}(),
			expectPage: `
			- View("base.html", github.com/rur/treetop.Noop)
			  '- test: SubView("test", "test.html", github.com/rur/treetop.Constant.func1)
			`,
			expectView: `- SubView("test", "test.html", github.com/rur/treetop.Constant.func1)`,
		}, {
			name: "with parent and overrideing includes",
			view: func() *View {
				base := NewView("base.html", Noop)
				base.NewSubView("other", "never_used.html", Noop)
				return base.NewSubView("test", "test.html", Constant("test!"))
			}(),
			expectPage: `
			- View("base.html", github.com/rur/treetop.Noop)
			  |- other: SubView("other", "other.html", github.com/rur/treetop.Constant.func1)
			  '- test: SubView("test", "test.html", github.com/rur/treetop.Constant.func1)
			`,
			expectView: `- SubView("test", "test.html", github.com/rur/treetop.Constant.func1)`,
			expectIncludes: []string{
				`- SubView("other", "other.html", github.com/rur/treetop.Constant.func1)`,
			},
			includes: []*View{
				NewSubView("other", "other.html", Constant("other!")),
			},
		}, {
			name: "include that overrides block in view",
			view: func() *View {
				base := NewView("base.html", Noop)
				base.NewSubView("other", "never_used.html", Noop)
				t := base.NewSubView("test", "test.html", Constant("test!"))
				t.NewSubView("subother", "never_used.html", Noop)
				return t
			}(),
			expectPage: `
			- View("base.html", github.com/rur/treetop.Noop)
			  |- other: SubView("other", "other.html", github.com/rur/treetop.Constant.func1)
			  '- test: SubView("test", "test.html", github.com/rur/treetop.Constant.func1)
			     '- subother: SubView("subother", "subother.html", github.com/rur/treetop.Constant.func1)
			`,
			expectView: `
			- SubView("test", "test.html", github.com/rur/treetop.Constant.func1)
			  '- subother: SubView("subother", "subother.html", github.com/rur/treetop.Constant.func1)
			`,
			expectIncludes: []string{
				`- SubView("other", "other.html", github.com/rur/treetop.Constant.func1)`,
			},
			includes: []*View{
				NewSubView("other", "other.html", Constant("other!")),
				NewSubView("subother", "subother.html", Constant("subother!")),
			},
		}, {
			name: "page include a chain of children",
			view: func() *View {
				base := NewView("base.html", Noop)
				c := base.NewSubView("content", "content.html", Noop)
				s := c.NewSubView("sub", "sub.html", Noop)
				return s
			}(),
			expectPage: `
			- View("base.html", github.com/rur/treetop.Noop)
			  '- content: SubView("content", "content.html", github.com/rur/treetop.Noop)
			     '- sub: SubView("sub", "sub.html", github.com/rur/treetop.Noop)
			`,
			expectView: `- SubView("sub", "sub.html", github.com/rur/treetop.Noop)`,
		},
	}
	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			page, view, incl := CompileViews(tCase.view, tCase.includes...)
			pageStr := normalizeTreePrint(SprintViewTree(page))
			expectPage := normalizeTreePrint(tCase.expectPage)
			if expectPage != pageStr {
				t.Errorf("Expecting Page =\n%s\nwant\n%s", pageStr, expectPage)
				return
			}
			viewStr := normalizeTreePrint(SprintViewTree(view))
			expectView := normalizeTreePrint(tCase.expectView)
			if expectView != viewStr {
				t.Errorf("Expecting View =\n%s\nwant\n%s", viewStr, expectView)
				return
			}
			inclS := make([]string, len(incl))
			for i, inc := range incl {
				inclS[i] = normalizeTreePrint(SprintViewTree(inc))
			}
			expectIncl := make([]string, len(tCase.expectIncludes))
			for i, inc := range tCase.expectIncludes {
				expectIncl[i] = normalizeTreePrint(inc)
			}
			if !reflect.DeepEqual(expectIncl, inclS) {
				t.Errorf("Expecting Include =\n%v\nwant\n%v", inclS, expectIncl)
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

func Test_insertView(t *testing.T) {
	tests := []struct {
		name      string
		view      *View
		child     *View
		want      string
		wantFound bool
	}{
		{
			name:      "basic",
			view:      NewView("test.html", Noop),
			child:     nil,
			want:      `- View("test.html", github.com/rur/treetop.Noop)`,
			wantFound: false,
		},
		{
			name: "found",
			view: func() *View {
				v := NewView("base.html", Noop)
				v.NewDefaultSubView("test", "gone.html", Noop)
				return v
			}(),
			child: NewSubView("test", "inserted.html", Noop),
			want: `
			- View("base.html", github.com/rur/treetop.Noop)
			  '- test: SubView("test", "inserted.html", github.com/rur/treetop.Noop)
			`,
			wantFound: true,
		},
		{
			name: "found, without interfering with other views",
			view: func() *View {
				v := NewView("base.html", Noop)
				v.NewDefaultSubView("test", "gone.html", Noop)
				v.NewDefaultSubView("test_other", "unaffected.html", Noop)
				return v
			}(),
			child: func() *View {
				b := NewView("different_base.html", Noop)
				incl := b.NewDefaultSubView("test", "included.html", Noop)
				incl.NewDefaultSubView("test_sub", "inclSub.html", Noop)
				return incl
			}(),
			want: `
			- View("base.html", github.com/rur/treetop.Noop)
			  |- test: SubView("test", "included.html", github.com/rur/treetop.Noop)
			  |  '- test_sub: SubView("test_sub", "inclSub.html", github.com/rur/treetop.Noop)
			  |
			  '- test_other: SubView("test_other", "unaffected.html", github.com/rur/treetop.Noop)
			`,
			wantFound: true,
		},
		{
			name: "keep nil",
			view: func() *View {
				v := NewView("base.html", Noop)
				v.NewDefaultSubView("test", "gone.html", Noop)
				v.NewSubView("another", "gone.html", Noop)
				return v
			}(),
			child: NewSubView("test", "inserted.html", Noop),
			want: `
			- View("base.html", github.com/rur/treetop.Noop)
			  |- another: nil
			  '- test: SubView("test", "inserted.html", github.com/rur/treetop.Noop)
			`,
			wantFound: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view, found := insertView(tt.view, tt.child)
			if found != tt.wantFound {
				t.Errorf("insetView() got found %v, want %v", found, tt.wantFound)
			}
			expect := normalizeTreePrint(tt.want)
			got := SprintViewTree(view)
			if expect != got {
				t.Errorf("insertView() got \n%s\nexpecting\n%s", got, expect)
			}
		})
	}
}

func TestViewBaseCopied(t *testing.T) {
	// test that views altered by compiling have been made immutable
	base := NewView("base.html", Noop)
	firstChild := base.NewSubView("test", "firstChild.html", Noop)
	p, _, _ := CompileViews(firstChild)
	base.NewDefaultSubView("test2", "secondChild.html", Noop)
	if strings.Contains(SprintViewTree(p), "test2") {
		t.Errorf("Changing base cause compiled view to change: got\n%s", SprintViewTree(p))
	}
}

func TestViewUnchangedUntilAfterCompile(t *testing.T) {
	// have a child of the base which is mutated AFTER compiling the views
	base := NewView("base.html", Noop)
	test2 := base.NewDefaultSubView("test2", "secondChild.html", Noop)
	firstChild := base.NewSubView("test", "firstChild.html", Noop)
	p, _, _ := CompileViews(firstChild)
	test2.NewDefaultSubView("test2_sub", "shouldNotBeInFirstChild.html", Noop)
	if strings.Contains(SprintViewTree(p), "test2_sub") {
		t.Errorf("Changing sibling view cause compiled view page to change: got\n%s", SprintViewTree(p))
	}
}
