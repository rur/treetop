package treetop

import (
	"bytes"
	"html/template"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestKeyedStringExecutor_NewViewHandler(t *testing.T) {
	base := NewView("base.html", Constant(struct {
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
	content := base.NewSubView("content", "content.html", Constant(struct {
		Message string
		Sub     interface{}
	}{
		Message: "from content to content!",
		Sub:     "from content to sub",
	}))
	content.NewDefaultSubView("sub", "sub.html", Constant("from sub to sub"))
	content.NewSubView("never", "never.html", Noop)
	ps := base.NewSubView("ps", "ps.html", Constant("from ps to ps"))

	tests := []struct {
		name           string
		exec           *KeyedStringExecutor
		expectPage     string
		expectTemplate string
		expectErrors   []string
		pageOnly       bool
		templateOnly   bool
	}{
		{
			name: "functional example",
			exec: NewKeyedStringExecutor(map[string]string{
				"base.html": `<html><body>
				{{ template "content" .Content }}

				{{ block "ps" .PS }}
				<p id="ps">Default {{ . }}</p>
				{{ end }}
				</body></html>`,
				"content.html": "<div id=\"content\">\n<p>Given {{ .Message }}</p>\n{{ template \"sub\" .Sub }}\n</div>{{ block \"never\" . }}{{ end }}",
				"sub.html":     `<p id="sub">Given {{ . }}</p>`,
				"ps.html":      `<div id="ps">Given {{ . }}</div>`,
			}),
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
			name:           "missing base template error",
			exec:           NewKeyedStringExecutor(map[string]string{}),
			expectPage:     "Not Acceptable",
			expectTemplate: "Not Acceptable",
			expectErrors: []string{
				`no key found for template 'base.html'`,
				`no key found for template 'content.html'`,
				`no key found for template 'ps.html'`,
			},
		},
		{
			name: "page only",
			exec: NewKeyedStringExecutor(map[string]string{
				"base.html": `<html><body>
				{{ template "content" .Content }}

				{{ block "ps" .PS }}
				<p id="ps">Default {{ . }}</p>
				{{ end }}
				</body></html>`,
				"content.html": "<div id=\"content\">\n<p>Given {{ .Message }}</p>\n{{ template \"sub\" .Sub }}\n</div>{{ block \"never\" . }}{{ end }}",
				"sub.html":     `<p id="sub">Given {{ . }}</p>`,
				"ps.html":      `<div id="ps">Given {{ . }}</div>`,
			}),
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
			name: "template only",
			exec: NewKeyedStringExecutor(map[string]string{
				"base.html": `<html><body>
				{{ template "content" .Content }}

				{{ block "ps" .PS }}
				<p id="ps">Default {{ . }}</p>
				{{ end }}
				{{ block "never" . }}{{ end }}
				</body></html>`,
				"content.html": "<div id=\"content\">\n<p>Given {{ .Message }}</p>\n{{ template \"sub\" .Sub }}\n</div>{{ block \"never\" . }}{{ end }}",
				"sub.html":     `<p id="sub">Given {{ . }}</p>`,
				"ps.html":      `<div id="ps">Given {{ . }}</div>`,
			}),
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
			name: "error, template missing a declared blockname",
			exec: NewKeyedStringExecutor(map[string]string{
				"base.html": `{{ template "content" .Content }}
				{{ block "ps" .PS }}<p id="ps">Default {{ . }}</p>{{ end }}`,
				"content.html": "MISSING TEMPLATE BLOCK 'sub' and 'never'",
				"sub.html":     `Sub`,
				"ps.html":      `Ps`,
			}),
			expectErrors: []string{
				`template content.html: missing template declaration(s) for sub view blocks: "never", "sub"`,
				`template content.html: missing template declaration(s) for sub view blocks: "never", "sub"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.exec.NewViewHandler(content, ps)
			err := tt.exec.FlushErrors()
			if err != nil {
				if len(tt.expectErrors) == 0 {
					t.Errorf("Unexpected error: %s", err)
				}
				for _, expect := range tt.expectErrors {
					if !strings.Contains(err.Error(), expect) {
						t.Errorf("Expecting error %s to contain: %s", err, expect)
					}
				}
				return
			} else if len(tt.expectErrors) > 0 {
				t.Errorf("Expected errors: %v", tt.expectErrors)
				return
			}

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
				t.Errorf("Expecting partial body\n%s\nGot\n%s", tt.expectTemplate, gotTemplate)
			}
		})
	}
}

func TestKeyedStringExecutor_NewViewHandler_NilView(t *testing.T) {
	exec := NewKeyedStringExecutor(nil)
	handler := exec.NewViewHandler(nil)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, mockRequest("/some/path", "*/*"))
	gotPage := stripIndent(sDumpBody(rec))

	expected := "Not Acceptable"

	if gotPage != expected {
		t.Errorf("Expecting page body\n%s\nGot\n%s", expected, gotPage)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, mockRequest("/some/path", TemplateContentType))
	gotTemplate := stripIndent(sDumpBody(rec))

	if gotTemplate != expected {
		t.Errorf("Expecting partial body\n%s\nGot\n%s", expected, gotTemplate)
	}

	tErs := exec.FlushErrors()
	if len(tErs) != 0 {
		t.Errorf("Unexpected errors: %v", tErs)
	}
}

func sDumpBody(rec *httptest.ResponseRecorder) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(rec.Body)
	return buf.String()
}

func TestKeyedStringExecutor_constructTemplate(t *testing.T) {
	tests := []struct {
		name    string
		exec    *KeyedStringExecutor
		view    *View
		data    interface{}
		want    string
		wantErr string
	}{
		{
			name: "basic",
			exec: NewKeyedStringExecutor(map[string]string{
				"test.key": "<p>hello {{ . }}!</p>",
			}),
			view:    NewView("test.key", Noop),
			data:    "world",
			want:    "<p>hello world!</p>",
			wantErr: "",
		},
		{
			name: "with subviews",
			exec: NewKeyedStringExecutor(map[string]string{
				"base.html":    `<div> base, content: {{ block "content" . }} default here {{ end }} </div>`,
				"content.html": `<p id="content">hello {{ . }}!</p>`,
			}),
			view: func() *View {
				b := NewView("base.html", Noop)
				b.NewDefaultSubView("content", "content.html", Noop)
				return b
			}(),
			data:    "world",
			want:    `<div> base, content: <p id="content">hello world!</p> </div>`,
			wantErr: "",
		},
		{
			name: "key not found",
			exec: NewKeyedStringExecutor(map[string]string{
				"base.html":    `<div> base, content: {{ block "content" . }} default here {{ end }} </div>`,
				"content.html": `<p id="content">hello {{ . }}!</p>`,
			}),
			view: func() *View {
				b := NewView("base.html", Noop)
				b.NewDefaultSubView("content", "content-other.html", Noop)
				return b
			}(),
			data:    "world",
			wantErr: "no key found for template 'content-other.html'",
		},
		{
			name: "multi level default children",
			exec: NewKeyedStringExecutor(map[string]string{
				"base.html": `<div> base, content: {{ block "content" . }} default here {{ end }} </div>`,
				"content.html": `<p id="content">
						<h2>hello {{ . }}!</h2>
						{{ template "sub" .}}
					</p>`,
				"sub.html": `<p id="sub">hello {{ . }}!</p>`,
			}),
			view: func() *View {
				b := NewView("base.html", Noop)
				c := b.NewDefaultSubView("content", "content.html", Noop)
				c.NewDefaultSubView("sub", "sub.html", Noop)
				return b
			}(),
			data: "world",
			want: stripIndent(`<div> base, content: <p id="content">
				<h2>hello world!</h2>
				<p id="sub">hello world!</p>
			</p> </div>`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tt.exec.NewViewHandler(tt.view).(*TemplateHandler)
			if !ok {
				t.Fatal("StringExecutor did not return a TemplateHandler")
			}
			err := tt.exec.FlushErrors()
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
			if stripIndent(gotString) != stripIndent(tt.want) {
				t.Errorf("KeyedStringExecutor.constructTemplate() got %v, want %v", gotString, tt.want)
			}
		})
	}
}

func TestNewKeyedStringExecutor(t *testing.T) {
	tests := []struct {
		name      string
		templates map[string]string
		key       string
		want      string
		wantErr   string
	}{
		{
			name: "",
			templates: map[string]string{
				"test.html": `<p>{{ . }}</p>`,
			},
			want: `<p>some data here</p>`,
		},
		{
			name: "",
			templates: map[string]string{
				"err.html": `<p>{{ .$TEST }}</p>`,
			},
			wantErr: `template: err.html:1: unexpected bad character U+0024 '$' in command`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewKeyedStringExecutor(tt.templates)
			if tt.key == "" {
				return
			}
			buf := new(bytes.Buffer)
			tpl, err := template.New("test").Parse(got.Templates[tt.key])
			if err != nil {
				t.Error(err)
			}

			err = tpl.Execute(buf, "some data here")
			if err != nil {
				t.Error(err)
			}

			gotStr := buf.String()
			if gotStr != tt.want {
				t.Errorf("NewKeyedStringExecutor() = %v, want %v", gotStr, tt.want)
			}
		})
	}
}

// stripIndent removes all whitespace at the beginning of every line
func stripIndent(s string) string {
	return strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(s, " "))
}

func TestKeyedStringExecutor_FuncMap(t *testing.T) {
	exec := KeyedStringExecutor{
		Templates: map[string]string{
			"test.html": `
			<div>
				<p>Input: {{printf "%q" .}}</p>
				<p>Output 0: {{title .}}</p>
				<p>Output 1: {{title . | printf "%q"}}</p>
				<p>Output 2: {{printf "%q" . | title}}</p>
			</div>
			`,
		},
		Funcs: template.FuncMap{
			"title": strings.Title,
		},
	}
	v := NewView("test.html", Constant("the go programming language"))
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
