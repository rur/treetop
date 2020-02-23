package treetop

import (
	"strings"
	"testing"
)

func Test_previewTemplate(t *testing.T) {
	type args struct {
		str    string
		before int
		after  int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "basic",
			args: args{
				str:    "some/path/on/fs.html",
				before: 10,
				after:  10,
			},
			want: `"some/path/on/fs.html"`,
		},
		{
			name: "realistic html",
			args: args{
				str:    fromExampleDotCom,
				before: 10,
				after:  10,
			},
			want: `"<!doctype……y></html>"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := previewTemplate(tt.args.str, tt.args.before, tt.args.after); got != tt.want {
				t.Errorf("previewTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSprintViewTree(t *testing.T) {
	tests := []struct {
		name string
		v    *View
		want string
	}{
		{
			name: "single view usage",
			v:    NewView("test.html", Constant("test!")),
			want: `- View("test.html", github.com/rur/treetop.Constant.func1)`,
		},
		{
			name: "view with sub views",
			v: func() *View {
				v := NewView("base.html", Constant("base!"))
				v.NewDefaultSubView("A", "A.html", Constant("A!"))
				v.NewDefaultSubView("B", "B.html", Constant("B!"))
				return v
			}(),
			want: `
			- View("base.html", github.com/rur/treetop.Constant.func1)
			  |- A: SubView("A", "A.html", github.com/rur/treetop.Constant.func1)
			  '- B: SubView("B", "B.html", github.com/rur/treetop.Constant.func1)
			`,
		},
		{
			name: "view with sub sub views",
			v: func() *View {
				v := NewView("base.html", Constant("base!"))
				a := v.NewDefaultSubView("A", "A.html", Constant("A!"))
				a.NewDefaultSubView("A1", "A1.html", Constant("A1!"))
				a.NewDefaultSubView("A2", "A2.html", Constant("A2!"))
				b := v.NewDefaultSubView("B", "B.html", Constant("B!"))
				b.NewDefaultSubView("B1", "B1.html", Constant("B1!"))
				b.NewDefaultSubView("B2", "B2.html", Constant("B2!"))
				return v
			}(),
			want: `
				- View("base.html", github.com/rur/treetop.Constant.func1)
				  |- A: SubView("A", "A.html", github.com/rur/treetop.Constant.func1)
				  |  |- A1: SubView("A1", "A1.html", github.com/rur/treetop.Constant.func1)
				  |  '- A2: SubView("A2", "A2.html", github.com/rur/treetop.Constant.func1)
				  |
				  '- B: SubView("B", "B.html", github.com/rur/treetop.Constant.func1)
				     |- B1: SubView("B1", "B1.html", github.com/rur/treetop.Constant.func1)
				     '- B2: SubView("B2", "B2.html", github.com/rur/treetop.Constant.func1)
			`,
		},
		{
			name: "view with sub sub sub views",
			v: func() *View {
				v := NewView("base.html", Constant("base!"))
				a := v.NewDefaultSubView("A", "A.html", Constant("A!"))
				a1 := a.NewDefaultSubView("A1", "A1.html", Constant("A1!"))
				a2 := a.NewDefaultSubView("A2", "A2.html", Constant("A2!"))
				a1.NewDefaultSubView("A11", "A11.html", Constant("A11!"))
				a2.NewDefaultSubView("A21", "A21.html", Constant("A21!"))
				b := v.NewDefaultSubView("B", "B.html", Constant("B!"))
				b1 := b.NewDefaultSubView("B1", "B1.html", Constant("B1!"))
				b1.NewDefaultSubView("B11", "B11.html", Constant("B11!"))
				b1.NewDefaultSubView("B12", "B12.html", Constant("B12!"))
				b2 := b.NewDefaultSubView("B2", "B2.html", Constant("B2!"))
				b2.NewDefaultSubView("B21", "B21.html", Constant("B21!"))
				b2.NewDefaultSubView("B22", "B22.html", Constant("B22!"))
				return v
			}(),
			want: `
            - View("base.html", github.com/rur/treetop.Constant.func1)
              |- A: SubView("A", "A.html", github.com/rur/treetop.Constant.func1)
              |  |- A1: SubView("A1", "A1.html", github.com/rur/treetop.Constant.func1)
              |  |  '- A11: SubView("A11", "A11.html", github.com/rur/treetop.Constant.func1)
              |  |
              |  '- A2: SubView("A2", "A2.html", github.com/rur/treetop.Constant.func1)
              |     '- A21: SubView("A21", "A21.html", github.com/rur/treetop.Constant.func1)
              |
              '- B: SubView("B", "B.html", github.com/rur/treetop.Constant.func1)
                 |- B1: SubView("B1", "B1.html", github.com/rur/treetop.Constant.func1)
                 |  |- B11: SubView("B11", "B11.html", github.com/rur/treetop.Constant.func1)
                 |  '- B12: SubView("B12", "B12.html", github.com/rur/treetop.Constant.func1)
                 |
                 '- B2: SubView("B2", "B2.html", github.com/rur/treetop.Constant.func1)
                    |- B21: SubView("B21", "B21.html", github.com/rur/treetop.Constant.func1)
                    '- B22: SubView("B22", "B22.html", github.com/rur/treetop.Constant.func1)
			`,
		},
		{
			name: "view with nil sub views",
			v: func() *View {
				v := NewView("base.html", Constant("base!"))
				a := v.NewDefaultSubView("A", "A.html", Constant("A!"))
				a.NewSubView("A1", "a1.html", Noop)
				a.NewDefaultSubView("A2", "A2.html", Constant("A2!"))
				b := v.NewDefaultSubView("B", "B.html", Constant("B!"))
				b.NewSubView("B1", "b1.html", Noop)
				return v
			}(),
			want: `
            - View("base.html", github.com/rur/treetop.Constant.func1)
              |- A: SubView("A", "A.html", github.com/rur/treetop.Constant.func1)
              |  |- A1: nil
              |  '- A2: SubView("A2", "A2.html", github.com/rur/treetop.Constant.func1)
              |
              '- B: SubView("B", "B.html", github.com/rur/treetop.Constant.func1)
                 '- B1: nil
			`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expecting := sanitizeExpectedTreePrint(tt.want)
			if got := SprintViewTree(tt.v); got != expecting {
				t.Errorf("SprintViewTree() =\n%s\nwant\n%s", got, expecting)
			}
		})
	}
}

// -----------
// helpers
// -----------

// sanitizeExpectedTreePrint will trim whitespace from test assertions
// and account for indentation in multiline raw strings.
func sanitizeExpectedTreePrint(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	var trimLeader int
FirstLineScan:
	for i, line := range lines {
		for j := range line {
			if line[j] != ' ' && line[j] != '\t' {
				// this is the first line
				// leading whitespace on first line should be
				// removed from all subsequent lines
				trimLeader = j
				lines = lines[i:]
				break FirstLineScan
			}
		}
	}
	for _, line := range lines {
		if len(line) <= trimLeader {
			continue
		}
		out = append(out, line[trimLeader:])
	}
	return strings.Join(out, "\n")
}

// fromExampleDotCom is used for testing template abbreviation
var fromExampleDotCom = `
<!doctype html>
<html>
<head>
    <title>Example Domain</title>

    <meta charset="utf-8" />
    <meta http-equiv="Content-type" content="text/html; charset=utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style type="text/css">
    body {
        background-color: #f0f0f2;
        margin: 0;
        padding: 0;
        font-family: "Open Sans", "Helvetica Neue", Helvetica, Arial, sans-serif;

    }
    div {
        width: 600px;
        margin: 5em auto;
        padding: 50px;
        background-color: #fff;
        border-radius: 1em;
    }
    a:link, a:visited {
        color: #38488f;
        text-decoration: none;
    }
    @media (max-width: 700px) {
        body {
            background-color: #fff;
        }
        div {
            width: auto;
            margin: 0 auto;
            border-radius: 0;
            padding: 1em;
        }
    }
    </style>
</head>

<body>
<div>
    <h1>Example Domain</h1>
    <p>This domain is established to be used for illustrative examples in documents. You may use this
    domain in examples without prior coordination or asking for permission.</p>
    <p><a href="http://www.iana.org/domains/example">More information...</a></p>
</div>
</body>
</html>
`
