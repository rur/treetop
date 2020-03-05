package treetop

import (
	"net/http"
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

	keyed.Templates["test"] = "<p>After {{ . }}</p>"
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
				template: test:1:11: executing &#34;test&#34; at &lt;.FAIL&gt;: can&#39;t evaluate field FAIL in type string
			</code>
		</pre>`),
	) {
		t.Error("Expecting Errors, got", got)
	}
}

func TestDeveloperExecutor_PageOnly(t *testing.T) {
	keyed, err := NewKeyedStringExecutor(map[string]string{
		"base.html": `
		<div>
			{{ template "test" . }}
		</div>
		`,
		"test.html": "<p>Test {{ . }}</p>",
	})
	if err != nil {
		t.Error("Failed to create keyed executor", err)
		return
	}
	dev := DeveloperExecutor{keyed}
	base := NewView("base.html", Delegate("test"))
	view := base.NewSubView("test", "test.html", Constant("data"))
	handler := dev.NewViewHandler(view).PageOnly()
	rec := httptest.NewRecorder()
	req := mockRequest("/some/path", "*/*")
	handler.ServeHTTP(rec, req)
	gotPage := sDumpBody(rec)
	if !strings.Contains(gotPage, "<p>Test data</p>") {
		t.Error("Expecting page, got", gotPage)
	}

	rec = httptest.NewRecorder()
	req = mockRequest("/some/path", TemplateContentType)
	handler.ServeHTTP(rec, req)
	gotTemplate := sDumpBody(rec)
	if !strings.Contains(
		stripIndent(gotTemplate),
		stripIndent(`
		<pre>
			<code>
				treetop template handler: server cannot produce a response matching the list of acceptable values
			</code>
		</pre>`),
	) {
		t.Error("Expecting Errors, got", gotTemplate)
	}
	if rec.Code != http.StatusNotAcceptable {
		t.Errorf("Expecting status %d, got %d", http.StatusNotAcceptable, rec.Code)
	}
}

func TestDeveloperExecutor_FragmentOnly(t *testing.T) {
	keyed, err := NewKeyedStringExecutor(map[string]string{
		"base.html": `
		<div>
			{{ template "test" . }}
		</div>
		`,
		"test.html": "<p>Test {{ . }}</p>",
	})
	if err != nil {
		t.Error("Failed to create keyed executor", err)
		return
	}
	dev := DeveloperExecutor{keyed}
	base := NewView("base.html", Delegate("test"))
	view := base.NewSubView("test", "test.html", Constant("data"))
	handler := dev.NewViewHandler(view).FragmentOnly()
	rec := httptest.NewRecorder()
	req := mockRequest("/some/path", "*/*")
	handler.ServeHTTP(rec, req)
	gotPage := sDumpBody(rec)
	if !strings.Contains(
		stripIndent(gotPage),
		stripIndent(`
		<pre>
			<code>
				treetop template handler: server cannot produce a response matching the list of acceptable values
			</code>
		</pre>`),
	) {
		t.Error("Expecting Errors, got", gotPage)
	}
	if rec.Code != http.StatusNotAcceptable {
		t.Errorf("Expecting status %d, got %d", http.StatusNotAcceptable, rec.Code)
	}

	rec = httptest.NewRecorder()
	req = mockRequest("/some/path", TemplateContentType)
	handler.ServeHTTP(rec, req)
	gotTemplate := sDumpBody(rec)

	if !strings.Contains(gotTemplate, "<p>Test data</p>") {
		t.Error("Expecting template, got", gotPage)
	}
}
