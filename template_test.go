package treetop

import (
	"html/template"
	"strconv"
	"strings"
	"testing"
)

func Test_viewQueue(t *testing.T) {
	// sanity check, queue works as expected
	queue := &viewQueue{}
	for i := 0; i < 10; i++ {
		queue.add(NewView(strconv.Itoa(i+1)+".html", Noop))
	}
	expect := []string{
		"1.html",
		"2.html",
		"3.html",
		"4.html",
		"5.html",
		"6.html",
		"7.html",
		"8.html",
		"9.html",
		"10.html",
	}
	var got []string
	for !queue.empty() {
		v, err := queue.next()
		if err != nil {
			t.Errorf("next returned an unexpected error %s", err)
			return
		}
		got = append(got, v.Template)
	}
	for len(got) < len(expect) {
		// pad got for diff purposes
		got = append(got, "")
	}
	for i := range got {
		if got[i] == "" {
			t.Errorf("Expecting template %s, got nothing", expect[i])
		} else if len(expect) <= i {
			t.Errorf("Unexpected template %s", got[i])
		} else if expect[i] != got[i] {
			t.Errorf("Expecting template %s, got %s", expect[i], got[i])
		}
	}
	if _, err := queue.next(); err != errEmptyViewQueue {
		t.Errorf("Expecting error '%s', got '%s'", errEmptyViewQueue, err)
	}
}

func Test_checkTemplateForBlockNames(t *testing.T) {
	type args struct {
		tmplString string
		subview    map[string]*View
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "empty template",
			args: args{
				tmplString: `with no blocks`,
			},
		},
		{
			name: "No declared subviews",
			args: args{
				tmplString: `before {{ block "test" . }}default{{ end }} after`,
			},
		},
		{
			name: "single declared subview",
			args: args{
				tmplString: `before {{ block "test" . }}default{{ end }} after`,
				subview: map[string]*View{
					"test": nil,
				},
			},
		},
		{
			name: "missing subview",
			args: args{
				tmplString: `before ?? after`,
				subview: map[string]*View{
					"test": nil,
				},
			},
			wantErr: `missing template declaration(s) for sub view blocks: "test"`,
		},
		{
			name: "template inside conditional",
			args: args{
				tmplString: `before  {{ if true }}{{ block "test" . }}default{{ end }} {{ end }}after`,
				subview: map[string]*View{
					"test": nil,
				},
			},
		},
		{
			name: "template inside range",
			args: args{
				tmplString: `before  {{ range $index, $item := .List }}{{ template "test" . }}{{ end }}after`,
				subview: map[string]*View{
					"test": nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := template.New(tt.name).Parse(tt.args.tmplString)
			if err != nil {
				t.Fatal("Failed to parse test template string", err)
			}

			err = checkTemplateForBlockNames(tmpl, tt.args.subview)

			if err != nil {
				if tt.wantErr == "" {
					t.Errorf("checkTemplateForBlockNames() unexpected error %v", err)
					return
				}
				// assert expected error is a substring of the actual err
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("checkTemplateForBlockNames() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if tt.wantErr != "" {
				t.Errorf("checkTemplateForBlockNames() expecting error that contains string = %v", tt.wantErr)
			}
		})
	}
}
