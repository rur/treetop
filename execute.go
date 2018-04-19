package treetop

import (
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

// implements TemplateExec
func DefaultTemplateExec(w io.Writer, templates []string, data interface{}) error {
	// trim strings and filter out empty
	filtered := make([]string, 0, len(templates))
	for _, templ := range templates {
		s := strings.TrimSpace(templ)
		if s != "" {
			filtered = append(filtered, s)
		}
	}

	t, err := template.New("__init__").ParseFiles(filtered...)
	if err != nil {
		return err
	}
	if err := t.ExecuteTemplate(w, filepath.Base(filtered[0]), data); err != nil {
		return err
	}
	return nil
}

func ExecutePartial(h Partial, handlerMap map[Block]Partial, resp http.ResponseWriter, r *http.Request) (Response, bool) {
	hw := partialWriter{
		ResponseWriter: resp,
		handler:        h,
		handlerMap:     handlerMap,
	}

	h.Func()(&hw, r)

	return Response{hw.data, hw.status}, !hw.wroteContent
}

func ExecuteFragment(h Fragment, dataMap map[string]interface{}, resp http.ResponseWriter, r *http.Request) (Response, bool) {
	hw := fragmentWriter{
		ResponseWriter: resp,
		handler:        h,
		datamap:        dataMap,
	}

	h.Func()(&hw, r)

	return Response{hw.data, hw.status}, !hw.wroteContent
}

type fragmentWriter struct {
	http.ResponseWriter
	handler      Fragment
	datamap      map[string]interface{}
	data         interface{}
	dataCalled   bool
	wroteContent bool
	status       int
}

func (fw *fragmentWriter) Write(b []byte) (int, error) {
	fw.wroteContent = true
	return fw.ResponseWriter.Write(b)
}

func (fw *fragmentWriter) WriteHeader(code int) {
	fw.wroteContent = true
	fw.ResponseWriter.WriteHeader(code)
}

func (fw *fragmentWriter) Data(d interface{}) {
	fw.data = d
	fw.dataCalled = true
}

func (fw *fragmentWriter) Status(code int) {
	fw.status = code
}

func (fw *fragmentWriter) PartialData(name string, r *http.Request) (interface{}, bool) {
	if fw.wroteContent {
		// response has already been written, nothing to do
		return nil, false
	}

	d, ok := fw.datamap[name]
	return d, ok
}

// implements treetop.go:DataWriter interface
type partialWriter struct {
	http.ResponseWriter
	handler      Partial
	handlerMap   map[Block]Partial
	data         interface{}
	dataCalled   bool
	wroteContent bool
	status       int
}

func (dw *partialWriter) Write(b []byte) (int, error) {
	dw.wroteContent = true
	return dw.ResponseWriter.Write(b)
}

func (dw *partialWriter) WriteHeader(code int) {
	dw.wroteContent = true
	dw.ResponseWriter.WriteHeader(code)
}

func (dw *partialWriter) Data(d interface{}) {
	dw.data = d
	dw.dataCalled = true
}

func (dw *partialWriter) Status(code int) {
	if code >= 100 && code < 600 && code > dw.status {
		dw.status = code
	}
}

func (dw *partialWriter) PartialData(name string, r *http.Request) (interface{}, bool) {
	if dw.wroteContent {
		// response has already been written, nothing to do
		return nil, false
	}

	block := dw.handler.Block(name)
	handler, ok := dw.handlerMap[block]
	if !ok {
		return nil, false
	}

	dw2 := partialWriter{
		ResponseWriter: dw.ResponseWriter,
		handler:        handler,
		handlerMap:     dw.handlerMap,
	}

	f := handler.Func()
	f(&dw2, r)

	if dw2.wroteContent {
		dw.wroteContent = true
		return nil, false
	}

	if dw2.dataCalled {
		return dw2.data, true
	} else {
		return nil, false
	}
}
