package treetop

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var handlerTemplateTestTemplateMap = map[string]string{
	"base.html": strings.Join([]string{
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
	}, "\n"),
	"content.html": strings.Join([]string{
		`<div id="content">`,
		`    <p>hello {{ .Message }}</p>`,
		`    {{ block "sub-content" .SubContent }}fallback{{ end }}`,
		`</div>`,
	}, "\n"),
	"override-nav.html": `<div id="nav">hello {{ . }}</div>`,
	"sub-content.html":  `<div id="sub-content">hello {{ . }}</div>`,
}

func setupTemplateHandler() ViewHandler {
	exec := NewKeyedStringExecutor(handlerTemplateTestTemplateMap)

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
		resp.Header().Add("Vary", "Cookie")
		return struct {
			Message    string
			SubContent interface{}
		}{
			Message:    "Content handler!!",
			SubContent: resp.HandleSubView("sub-content", req),
		}
	})
	content.NewDefaultSubView("sub-content", "sub-content.html", Constant("sub content!!"))

	handler := exec.NewViewHandler(content, NewSubView("nav", "override-nav.html", Constant("override nav handler!!")))
	errs := exec.FlushErrors()
	if len(errs) > 0 {
		panic(errs)
	}
	return handler
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

	gotVary := strings.Join(rec.Header().Values("Vary"), ", ")
	expectVary := "Cookie, Accept"
	if gotVary != expectVary {
		t.Errorf("Expecting Vary header: [%s], got: [%s]", expectVary, gotVary)
	}

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
	handler := setupTemplateHandler().PageOnly()
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

	gotVary := strings.Join(rec.Header().Values("Vary"), ", ")
	expectVary := "Cookie"
	if gotVary != expectVary {
		t.Errorf("Expecting Vary header: [%s], got: [%s]", expectVary, gotVary)
	}

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

func TestTemplateHandler_DesignatePageURL(t *testing.T) {
	expecting := "/designated/in/handler"
	exec := NewKeyedStringExecutor(handlerTemplateTestTemplateMap)
	v := NewSubView("sub-content", "sub-content.html", func(resp Response, req *http.Request) interface{} {
		resp.DesignatePageURL(expecting)
		return "testing"
	})

	th := exec.NewViewHandler(v)
	rec := httptest.NewRecorder()
	th.ServeHTTP(rec, mockRequest("/some/path", TemplateContentType))
	pageURL := rec.Header().Get("X-Page-URL")
	if pageURL != expecting {
		t.Errorf("Expecting X-Page-URL header to be %s, got %s", expecting, pageURL)
	}
}

func TestTemplateHandler_AdditionalVaryHeaders(t *testing.T) {
	exec := NewKeyedStringExecutor(handlerTemplateTestTemplateMap)
	v := NewSubView("sub-content", "sub-content.html", func(resp Response, req *http.Request) interface{} {
		resp.Header().Add("Vary", "Cookie")
		return "testing"
	})

	th := exec.NewViewHandler(v)
	rec := httptest.NewRecorder()
	th.ServeHTTP(rec, mockRequest("/some/path", TemplateContentType))
	expecting := "Cookie, Accept"
	varyHeader := rec.Header().Values("Vary")
	if strings.Join(varyHeader, ", ") != expecting {
		t.Errorf("Expecting Vary header to be [%s], got %v", expecting, varyHeader)
	}
}
