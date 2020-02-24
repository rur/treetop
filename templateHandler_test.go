package treetop

import (
	"bytes"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupTemplateHandler() *TemplateHandler {
	baseTpl := strings.Join([]string{
		`<!DOCTYPE html>`,
		`<html>`,
		`<head>`,
		`	<title>base</title>`,
		`</head>`,
		`<body>`,
		`	<div class="top">	`,
		`		{{ block "nav" .Nav }}fallback block nav{{ end }}`,
		`	</div>`,
		`	<div class="content">	`,
		`		{{ block "content" .Content }}fallback block content{{ end }}`,
		`	</div>`,
		`</body>`,
		`</html>`,
	}, "\n")
	contentTpl := strings.Join([]string{
		`<div id="content">`,
		`    <p>hello {{ .Message }}</p>`,
		`    {{ block "sub-content" .SubContent }}fallback{{ end }}`,
		`</div>`,
	}, "\n")
	navTpl := `<div id="nav">hello {{ . }}</div>`
	subContentTpl := `<div id="sub-content">hello {{ . }}</div>`

	base := NewView("base.html", func(resp Response, req *http.Request) interface{} {
		return struct {
			Nav     interface{}
			Content interface{}
		}{
			Nav:     resp.HandleSubView("nav", req),
			Content: resp.HandleSubView("content", req),
		}
	})
	base.NewDefaultSubView("nav", "nav.html", Noop)
	content := base.NewSubView("content", "content.html", func(resp Response, req *http.Request) interface{} {
		return struct {
			Message    string
			SubContent interface{}
		}{
			Message:    "Content handler!!",
			SubContent: resp.HandleSubView("sub-content", req),
		}
	})
	content.NewDefaultSubView("sub-content", "sub-content.html", Constant("sub content!!"))

	page, part, ps := CompileViews(content, NewSubView("nav", "override-nav.html", Constant("override nav handler!!")))
	// parse HTML templates
	pageTemplate, _ := template.New("base").Parse(baseTpl)
	partTemplate, _ := template.New("content").Parse(contentTpl)
	subCTemplate, _ := template.New("sub-content").Parse(subContentTpl)
	navTemplate, _ := template.New("nav").Parse(navTpl)

	pageTemplate.AddParseTree("content", partTemplate.Tree)
	pageTemplate.AddParseTree("sub-content", subCTemplate.Tree)
	pageTemplate.AddParseTree("nav", navTemplate.Tree)

	partTemplate.AddParseTree("sub-content", subCTemplate.Tree)
	// create handler
	return &TemplateHandler{
		Page:            page,
		Partial:         part,
		Includes:        ps,
		PageTemplate:    pageTemplate,
		PartialTemplate: partTemplate,
		IncludeTemplates: []*template.Template{
			navTemplate,
		},
	}
}

func TestTemplateHandler_PartialRequest(t *testing.T) {
	th := setupTemplateHandler()
	rec := httptest.NewRecorder()
	th.ServeHTTP(rec, mockRequest("/some/path", TemplateContentType))

	// assertions:
	if rec.Code != http.StatusOK {
		t.Errorf("Expecting status 200 got %d", rec.Code)
	}
	if val := rec.Header().Get("Content-Type"); val != TemplateContentType {
		t.Errorf("Execting content type header value %s, got %s", TemplateContentType, val)
	}
	expecting := strings.Join([]string{
		`<template>`,
		`<div id="content">`,
		`    <p>hello Content handler!!</p>`,
		`    <div id="sub-content">hello sub content!!</div>`,
		`</div>`,
		`<div id="nav">hello override nav handler!!</div>`,
		`</template>`,
	}, "\n")

	buf := new(bytes.Buffer)
	buf.ReadFrom(rec.Body)

	if body := buf.String(); body != expecting {
		t.Errorf("Expecting body \n%s\nGOT\n%s", expecting, body)
	}
}

func TestTemplateHandler_PageRequestNotAcceptable(t *testing.T) {
	th := setupTemplateHandler()
	handler := th.FragmentOnly()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, mockRequest("/some/path", "*/*"))

	// assert that a not acceptable error message was recorded
	if rec.Code != http.StatusNotAcceptable {
		t.Errorf("Expecting status %d got %d", http.StatusNotAcceptable, rec.Code)
	}
	if val := rec.Header().Get("Content-Type"); strings.Contains("text/plain", val) {
		t.Errorf("Execting content type header value %s, got %s", "text/plain", val)
	}
	expecting := "Not Acceptable\n"

	buf := new(bytes.Buffer)
	buf.ReadFrom(rec.Body)

	if body := buf.String(); body != expecting {
		t.Errorf("Expecting body \n%s\ngot\n%s", expecting, body)
	}
}

func TestTemplateHandler_PageRequest(t *testing.T) {
	th := setupTemplateHandler()
	handler := th.PageOnly()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, mockRequest("/some/path", "*/*"))

	// assertions that a HTML document was successfully recorded
	if rec.Code != http.StatusOK {
		t.Errorf("Expecting status 200 got %d", rec.Code)
	}
	if val := rec.Header().Get("Content-Type"); val != "text/html" {
		t.Errorf("Execting content type header value %s, got %s", "text/html", val)
	}
	expecting := strings.Join([]string{
		`<!DOCTYPE html>`,
		`<html>`,
		`<head>`,
		`	<title>base</title>`,
		`</head>`,
		`<body>`,
		`	<div class="top">	`,
		`		<div id="nav">hello override nav handler!!</div>`,
		`	</div>`,
		`	<div class="content">	`,
		`		<div id="content">`,
		`    <p>hello Content handler!!</p>`,
		`    <div id="sub-content">hello sub content!!</div>`,
		`</div>`,
		`	</div>`,
		`</body>`,
		`</html>`,
	}, "\n")

	buf := new(bytes.Buffer)
	buf.ReadFrom(rec.Body)

	if body := buf.String(); body != expecting {
		t.Errorf("Expecting body \n%s\ngot\n%s", expecting, body)
	}
}

func TestTemplateHandler_TemplateRequestNotAcceptable(t *testing.T) {
	th := setupTemplateHandler()
	handler := th.PageOnly()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, mockRequest("/some/path", TemplateContentType))

	// assert that a not acceptable error message was recorded
	if rec.Code != http.StatusNotAcceptable {
		t.Errorf("Expecting status %d got %d", http.StatusNotAcceptable, rec.Code)
	}
	if val := rec.Header().Get("Content-Type"); strings.Contains("text/plain", val) {
		t.Errorf("Execting content type header value %s, got %s", "text/plain", val)
	}
	expecting := "Not Acceptable\n"

	buf := new(bytes.Buffer)
	buf.ReadFrom(rec.Body)

	if body := buf.String(); body != expecting {
		t.Errorf("Expecting body \n%s\ngot\n%s", expecting, body)
	}
}
