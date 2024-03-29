package treetop

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFileExecutor_NewViewHandler(t *testing.T) {
	standardHandler := func(exec ViewExecutor) ViewHandler {
		base := NewView("testdata/base.html", Constant(struct {
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
		content := base.NewSubView("content", "testdata/content.html", Constant(struct {
			Message string
			Sub     interface{}
		}{
			Message: "from content to content!",
			Sub:     "from content to sub",
		}))
		content.NewDefaultSubView("sub", "testdata/sub.html", Constant("from sub to sub"))
		ps := base.NewSubView("ps", "testdata/ps.html", Constant("from ps to ps"))
		return exec.NewViewHandler(content, ps)
	}

	tests := []struct {
		name           string
		getHandler     func(exec ViewExecutor) ViewHandler
		expectPage     string
		expectTemplate string
		expectErrors   []string
		pageOnly       bool
		templateOnly   bool
	}{
		{
			name:       "functional example",
			getHandler: standardHandler,
			expectPage: stripIndent(`<html>
			<body>
			<div id="content">
			<p>Given from base to content</p>
			<p id="sub">Given from base via content to sub</p>
			</div>
			<div id="ps">Given from base to ps</div>
			</body>
			</html>`),
			expectTemplate: stripIndent(`<template>
			<div id="content">
			<p>Given from content to content!</p>
			<p id="sub">Given from content to sub</p>
			</div>

			<div id="ps">Given from ps to ps</div>
			</template>`),
		},
		{
			name:       "page only",
			getHandler: standardHandler,
			expectPage: stripIndent(`<html>
			<body>
			<div id="content">
			<p>Given from base to content</p>
			<p id="sub">Given from base via content to sub</p>
			</div>

			<div id="ps">Given from base to ps</div>
			</body>
			</html>`),
			expectTemplate: "Not Acceptable\n",
			pageOnly:       true,
		},
		{
			name:       "template only",
			getHandler: standardHandler,
			expectPage: "Not Acceptable\n",
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
			name: "file not found",
			getHandler: func(exec ViewExecutor) ViewHandler {
				return exec.NewViewHandler(NewView("notexists.html", Noop))
			},
			expectPage:     "Not Acceptable\n",
			expectTemplate: "Not Acceptable\n",
			expectErrors: []string{
				`failed to load template "notexists.html": open notexists.html: no such file or directory`,
				`failed to load template "notexists.html": open notexists.html: no such file or directory`,
			},
		},
		{
			name: "error template file",
			getHandler: func(exec ViewExecutor) ViewHandler {
				return exec.NewViewHandler(NewView("testdata/missingFunc.html", Noop))
			},
			expectPage:     "Not Acceptable\n",
			expectTemplate: "Not Acceptable\n",
			expectErrors: []string{
				`failed to parse template "testdata/missingFunc.html": template: :2: function "func_that_does_not_exist" not defined`,
				`failed to parse template "testdata/missingFunc.html": template: :2: function "func_that_does_not_exist" not defined`,
			},
		},
		{
			name: "with nil subview",
			getHandler: func(exec ViewExecutor) ViewHandler {
				base := NewView("testdata/base.html", Constant(struct {
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
				content := base.NewSubView("content", "testdata/content.html", Constant(struct {
					Message string
					Sub     interface{}
				}{
					Message: "from content to content!",
					Sub:     "from content to sub",
				}))
				content.NewDefaultSubView("sub", "testdata/sub.html", Constant("from sub to sub"))
				content.NewSubView("never", "never.html", Noop) // nil subview
				ps := base.NewSubView("ps", "testdata/ps.html", Constant("from ps to ps"))
				return exec.NewViewHandler(content, ps)
			},
			expectPage: stripIndent(`<html>
			<body>
			<div id="content">
			<p>Given from base to content</p>
			<p id="sub">Given from base via content to sub</p>
			</div>

			<div id="ps">Given from base to ps</div>
			</body>
			</html>
			`),
			expectTemplate: stripIndent(`<template>
			<div id="content">
			<p>Given from content to content!</p>
			<p id="sub">Given from content to sub</p>
			</div>

			<div id="ps">Given from ps to ps</div>
			</template>`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exec := &FileExecutor{}
			handler := tt.getHandler(exec)

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

			if tt.pageOnly {
				handler = handler.PageOnly()
			}
			if tt.templateOnly {
				handler = handler.FragmentOnly()
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, mockRequest("/some/path", "*/*"))
			gotPage := stripIndent(sDumpBody(rec))

			if gotPage != stripIndent(tt.expectPage) {
				t.Errorf("Expecting page body\n%s\nGot\n%s", tt.expectPage, gotPage)
			}

			rec = httptest.NewRecorder()
			handler.ServeHTTP(rec, mockRequest("/some/path", TemplateContentType))
			gotTemplate := stripIndent(sDumpBody(rec))

			if gotTemplate != stripIndent(tt.expectTemplate) {
				t.Errorf("Expecting partial body\n%s\nGot\n%s", tt.expectTemplate, gotTemplate)
			}

		})
	}
}

func TestFileExecutor_FuncMap(t *testing.T) {
	exec := FileExecutor{
		Funcs: template.FuncMap{
			"title": strings.Title,
		},
	}
	v := NewView("./testdata/titles.html", Constant("the go programming language"))
	h := exec.NewViewHandler(v)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, mockRequest("/some/path", "*/*"))
	gotPage := stripIndent(sDumpBody(rec))
	if gotPage != stripIndent(strings.TrimSpace(`
	<div>
		<p>Input: &#34;the go programming language&#34;</p>
		<p>Output 0: The Go Programming Language</p>
		<p>Output 1: &#34;The Go Programming Language&#34;</p>
		<p>Output 2: &#34;The Go Programming Language&#34;</p>
	</div>
	`)) {
		t.Errorf("Expecting title case, got\n%s", gotPage)
	}
}

func TestFileExecutor_KeyedString(t *testing.T) {
	exec := FileExecutor{
		KeyedString: map[string]string{
			"local://titles.html": `<h1>{{ . }}</h1>`,
		},
	}
	v := NewView("local://titles.html", Constant("Test Title"))
	h := exec.NewViewHandler(v)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, mockRequest("/some/path", "*/*"))
	gotPage := stripIndent(sDumpBody(rec))
	if gotPage != stripIndent(strings.TrimSpace("<h1>Test Title</h1>")) {
		t.Errorf("Expecting title case, got\n%s", gotPage)
	}
}

func TestFileExecutor_UsingSameTemplate(t *testing.T) {
	exec := FileExecutor{
		KeyedString: map[string]string{
			"local://common.html": `<h1>Common {{ . }}</h1>`,
		},
	}
	v := NewView("./testdata/base.html", func(rsp Response, req *http.Request) interface{} {
		return struct {
			Content interface{}
			PS      interface{}
		}{
			Content: rsp.HandleSubView("content", req),
			PS:      rsp.HandleSubView("ps", req),
		}
	})
	v.NewDefaultSubView("content", "local://common.html", Constant("Content Data"))
	v.NewDefaultSubView("ps", "local://common.html", Constant("Postscript Data"))
	h := exec.NewViewHandler(v)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, mockRequest("/some/path", "*/*"))
	gotPage := strings.TrimSpace(stripIndent(sDumpBody(rec)))
	expectPage := strings.TrimSpace(stripIndent(`
	<html>
	<body>
	<h1>Common Content Data</h1>
	<h1>Common Postscript Data</h1>
	</body>
	</html>
	`))
	if gotPage != expectPage {
		t.Errorf("Expecting\n%s\ngot\n%s", expectPage, gotPage)
	}
}
