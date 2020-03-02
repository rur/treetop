package treetop

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFileSystemExecutor_NewViewHandler(t *testing.T) {
	standardHandler := func(exec ViewExecutor) ViewHandler {
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
		ps := base.NewSubView("ps", "ps.html", Constant("from ps to ps"))
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
			name:       "page only",
			getHandler: standardHandler,
			expectPage: stripIndent(`<html><body>
			<div id="content">
			<p>Given from base to content</p>
			<p id="sub">Given from base via content to sub</p>
			</div>
			
			<div id="ps">Given from base to ps</div>
			</body></html>`),
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
				`Failed to open template file 'notexists.html', ` +
					`error open testdata/notexists.html: no such file or directory`,
				`Failed to open template file 'notexists.html', ` +
					`error open testdata/notexists.html: no such file or directory`,
			},
		},
		{
			name: "malformed file",
			getHandler: func(exec ViewExecutor) ViewHandler {
				return exec.NewViewHandler(NewView("malformed.html", Noop))
			},
			expectPage:     "Not Acceptable\n",
			expectTemplate: "Not Acceptable\n",
			expectErrors: []string{
				`Failed to parse contents of template file 'malformed.html', ` +
					`error template: :2: function "functhatdoesnexist" not defined`,
				`Failed to parse contents of template file 'malformed.html', ` +
					`error template: :2: function "functhatdoesnexist" not defined`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exec := &FileSystemExecutor{
				FS: http.Dir("testdata"),
			}
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
				t.Errorf("Expecting partial body\n%s\nGot\n%s", tt.expectTemplate, gotTemplate)
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
