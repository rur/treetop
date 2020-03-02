package treetop

import (
	"html/template"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDeveloperExecutor_UpdateTemplate(t *testing.T) {
	keyed, err := NewKeyedStringExecutor(map[string]string{
		"test": "<p>Before {{ . }}</p>",
	})
	if err != nil {
		t.Error("Failed to create keyed executor", err)
		return
	}
	dev := DeveloperExecutor{keyed}
	handler := dev.NewViewHandler(NewView("test", Constant("from handler")))
	rec := httptest.NewRecorder()
	req := mockRequest("/some/path", "*/*")
	handler.ServeHTTP(rec, req)
	gotBefore := sDumpBody(rec)

	update := template.Must(template.New("test").Parse("<p>After {{ . }}</p>"))
	keyed.parsed["test"] = update.Tree
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	gotAfter := sDumpBody(rec)

	if gotBefore != `<p>Before from handler</p>` {
		t.Error("Expecting the before template, got", gotBefore)
	}
	if gotAfter != `<p>After from handler</p>` {
		t.Error("Expecting the before template, got", gotAfter)
	}
}

func TestDeveloperExecutor_RenderErrors(t *testing.T) {
	keyed, err := NewKeyedStringExecutor(map[string]string{
		"base.html": `
		<div>
			{{ template "test" . }}
		</div>
		`,
		"test.html": "<p>Test {{ .FAIL }}</p>",
	})
	if err != nil {
		t.Error("Failed to create keyed executor", err)
		return
	}
	dev := DeveloperExecutor{keyed}
	base := NewView("base.html", Delegate("test"))
	view := base.NewSubView("test", "test.html", Constant("data"))
	handler := dev.NewViewHandler(view)
	rec := httptest.NewRecorder()
	req := mockRequest("/some/path", "*/*")
	handler.ServeHTTP(rec, req)
	got := sDumpBody(rec)

	if !strings.Contains(
		stripIndent(got),
		stripIndent(`
		<pre>
			<code>
				template: test.html:1:11: executing &#34;test&#34; at &lt;.FAIL&gt;: can&#39;t evaluate field FAIL in type string
			</code>
		</pre>`),
	) {
		t.Error("Expecting Errors, got", got)
	}
}
