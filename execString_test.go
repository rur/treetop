package treetop

import (
	"bytes"
	"html/template"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStringExecutor_constructTemplate(t *testing.T) {
	tests := []struct {
		name    string
		view    *View
		data    interface{}
		want    string
		wantErr string
	}{
		{
			name:    "basic",
			view:    NewView("<p>hello {{ . }}!</p>", Noop),
			data:    "world",
			want:    "<p>hello world!</p>",
			wantErr: "",
		},
		{
			name: "with subviews",
			view: func() *View {
				b := NewView(`<div> base, content: {{ block "content" . }} default here {{ end }} </div>`, Noop)
				b.NewDefaultSubView("content", `<p id="content">hello {{ . }}!</p>`, Noop)
				return b
			}(),
			data: "world",
			want: `<div> base, content: <p id="content">hello world!</p> </div>`,
		},
		{
			name: "template parse error",
			view: func() *View {
				b := NewView(`<div> base, content: {{ b.^^lock "content" . }} default here {{ end }} </div>`, Noop)
				b.NewDefaultSubView("content", `<p id="content">hello {{ . }}!</p>`, Noop)
				return b
			}(),
			data:    "world",
			wantErr: `template: :1: function "b" not defined`,
		},
		{
			name: "with nil subviews",
			view: func() *View {
				b := NewView(`<div> base, content: {{ block "content" . }} default here {{ end }} </div>`, Noop)
				b.NewSubView("content", `<p id="content">hello {{ . }}!</p>`, Noop)
				return b
			}(),
			data: "world",
			want: `<div> base, content:  default here  </div>`,
		},
		{
			name: "error, template missing a declared blockname",
			view: func() *View {
				b := NewView(`<div> base, content: </div>`, Noop)
				b.NewSubView("content", `<p id="content">hello {{ . }}!</p>`, Noop)
				return b
			}(),
			wantErr: `<div> base, content: </div> is missing template declaration(s) for sub view blocks: "content"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exec := StringExecutor{}
			got, ok := exec.NewViewHandler(tt.view).(*TemplateHandler)
			if !ok {
				t.Fatal("StringExecutor did not return a TemplateHandler")
			}
			err := exec.FlushErrors()

			if err != nil {
				if tt.wantErr == "" {
					t.Errorf("Unexpected error: %s", err)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Expecting error %s to contain: %s", err, tt.wantErr)
				}
				return
			} else if tt.wantErr != "" {
				t.Errorf("Expected error %s", tt.wantErr)
				return
			}

			buf := new(bytes.Buffer)
			got.PageTemplate.ExecuteTemplate(buf, tt.view.Defines, tt.data)
			gotString := buf.String()
			if gotString != tt.want {
				t.Errorf("StringExecutor.constructTemplate() got %v, want %v", gotString, tt.want)
			}
		})
	}
}

func TestStringExecutor_NewViewHandler(t *testing.T) {
	// setup
	standardHandler := func(exec ViewExecutor) ViewHandler {
		base := NewView(`<html><body>
				{{ template "content" .Content }}

				{{ block "ps" .PS }}
				<p id="ps">Default {{ . }}</p>
				{{ end }}
				</body></html>`, Constant(struct {
			Content interface{}
			PS      interface{}
		}{
			Content: struct {
				Message string
				Sub     interface{}
			}{
				Message: "from base to content",
				Sub:     "from base via content to sub",
			},
			PS: "from base to ps",
		}))
		content := base.NewSubView(
			"content",
			"<div id=\"content\">\n<p>Given {{ .Message }}</p>\n{{ template \"sub\" .Sub }}\n</div>",
			Constant(struct {
				Message string
				Sub     interface{}
			}{
				Message: "from content to content!",
				Sub:     "from content to sub",
			}),
		)
		content.NewDefaultSubView("sub", `<p id="sub">Given {{ . }}</p>`, Constant("from sub to sub"))
		ps := base.NewSubView("ps", `<div id="ps">Given {{ . }}</div>`, Constant("from ps to ps"))
		return exec.NewViewHandler(content, ps)
	}

	// tests
	tests := []struct {
		name           string
		getHandler     func(ViewExecutor) ViewHandler
		expectPage     string
		expectTemplate string
		expectErrors   []string
		pageOnly       bool
		templateOnly   bool
	}{
		{
			name:       "functional example",
			getHandler: standardHandler,
			expectPage: stripIndent(`<html><body>
			<div id="content">
			<p>Given from base to content</p>
			<p id="sub">Given from base via content to sub</p>
			</div>

			<div id="ps">Given from base to ps</div>
			</body></html>`),
			expectTemplate: stripIndent(`<template>
			<div id="content">
			<p>Given from content to content!</p>
			<p id="sub">Given from content to sub</p>
			</div>
			<div id="ps">Given from ps to ps</div>
			</template>`),
		},
		{
			name: "template parse errors",
			getHandler: func(exec ViewExecutor) ViewHandler {
				base := NewView(`{{ fail }}`, Constant(struct {
					Content interface{}
					PS      interface{}
				}{
					Content: struct {
						Message string
						Sub     interface{}
					}{
						Message: "from base to content",
						Sub:     "from base via content to sub",
					},
					PS: "from base to ps",
				}))
				content := base.NewSubView(
					"content",
					`{{ failcontent }}`,
					Constant(struct {
						Message string
						Sub     interface{}
					}{
						Message: "from content to content!",
						Sub:     "from content to sub",
					}),
				)
				content.NewDefaultSubView("sub", `{{ failsub }}`, Constant("from sub to sub"))
				ps := base.NewSubView("ps", `{{ failps }}`, Constant("from ps to ps"))
				return exec.NewViewHandler(content, ps)
			},
			expectPage:     "Not Acceptable",
			expectTemplate: "Not Acceptable",
			expectErrors: []string{
				`failed to parse template "{{ fail }}": template: :1: function "fail" not defined`,
				`failed to parse template "{{ failcontent }}": template: content:1: function "failcontent" not defined`,
				`failed to parse template "{{ failps }}": template: ps:1: function "failps" not defined`,
			},
		},
		{
			name:       "page only",
			getHandler: standardHandler,
			expectPage: stripIndent(`<html><body>
			<div id="content">
			<p>Given from base to content</p>
			<p id="sub">Given from base via content to sub</p>
			</div>

			<div id="ps">Given from base to ps</div>
			</body></html>`),
			expectTemplate: "Not Acceptable",
			pageOnly:       true,
		},
		{
			name:       "template only",
			getHandler: standardHandler,
			expectPage: "Not Acceptable",
			expectTemplate: stripIndent(`<template>
			<div id="content">
			<p>Given from content to content!</p>
			<p id="sub">Given from content to sub</p>
			</div>
			<div id="ps">Given from ps to ps</div>
			</template>`),
			templateOnly: true,
		},
		{
			name: "nil view",
			getHandler: func(exec ViewExecutor) ViewHandler {
				return exec.NewViewHandler(nil)
			},
			expectPage:     "Not Acceptable",
			expectTemplate: "Not Acceptable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exec := &StringExecutor{}
			handler := tt.getHandler(exec)
			if tt.pageOnly {
				handler = handler.PageOnly()
			}
			if tt.templateOnly {
				handler = handler.FragmentOnly()
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, mockRequest("/some/path", "*/*"))
			gotPage := stripIndent(sDumpBody(rec))

			if gotPage != tt.expectPage {
				t.Errorf("Expecting page body\n%s\nGot\n%s", tt.expectPage, gotPage)
			}

			rec = httptest.NewRecorder()
			handler.ServeHTTP(rec, mockRequest("/some/path", TemplateContentType))
			gotTemplate := stripIndent(sDumpBody(rec))

			if gotTemplate != tt.expectTemplate {
				t.Errorf("Expecting template body\n%s\nGot\n%s", tt.expectTemplate, gotTemplate)
			}

			gotErrors := exec.FlushErrors()
			for len(gotErrors) < len(tt.expectErrors) {
				gotErrors = append(gotErrors, nil)
			}

			for i, err := range gotErrors {
				if err == nil {
					t.Errorf("Expecting an error [%d]: %s", i, tt.expectErrors[i])
					continue
				}
				if i >= len(tt.expectErrors) {
					t.Errorf("Unexpected error [%d]: %s", i, err.Error())
					continue
				}
				if got := err.Error(); got != tt.expectErrors[i] {
					t.Errorf("Expecting error [%d]\n%s\ngot\n%s", i, tt.expectErrors[i], got)
				}
			}
		})
	}
}

func TestStringExecutor_FuncMap(t *testing.T) {
	exec := StringExecutor{
		Funcs: template.FuncMap{
			"title": strings.Title,
		},
	}
	v := NewView(`
	<div>
		<p>Input: {{printf "%q" .}}</p>
		<p>Output 0: {{title .}}</p>
		<p>Output 1: {{title . | printf "%q"}}</p>
		<p>Output 2: {{printf "%q" . | title}}</p>
	</div>
	`, Constant("the go programming language"))
	h := exec.NewViewHandler(v)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, mockRequest("/some/path", "*/*"))
	gotPage := stripIndent(sDumpBody(rec))
	if gotPage != stripIndent(`
	<div>
		<p>Input: &#34;the go programming language&#34;</p>
		<p>Output 0: The Go Programming Language</p>
		<p>Output 1: &#34;The Go Programming Language&#34;</p>
		<p>Output 2: &#34;The Go Programming Language&#34;</p>
	</div>
	`) {
		t.Errorf("Expecting title case, got\n%s", gotPage)
	}
}
