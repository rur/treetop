package treetop

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
)

// implements TemplateExec
func DefaultTemplateExec(w io.Writer, templates []string, data interface{}) error {
	t, err := template.New("__init__").ParseFiles(templates...)
	if err != nil {
		return err
	}
	if err := t.ExecuteTemplate(w, filepath.Base(templates[0]), data); err != nil {
		return err
	}
	return nil
}

func ExecutePartial(h Partial, handlerMap map[Block]Partial, resp http.ResponseWriter, r *http.Request) (interface{}, bool) {
	hw := partialWriter{
		ResponseWriter: resp,
		handler:        h,
		handlerMap:     handlerMap,
	}

	h.Func()(&hw, r)

	if hw.wroteHeader {
		// response headers have already been written in one of the handlers, do not proceed
		return hw.data, false
	} else {
		return hw.data, true
	}
}

func ExecuteFragment(h Handler, dataMap map[string]interface{}, resp http.ResponseWriter, r *http.Request) (interface{}, bool) {
	hw := fragmentWriter{
		ResponseWriter: resp,
		handler:        h,
		datamap:        dataMap,
	}

	h.Func()(&hw, r)

	if hw.wroteHeader {
		// response headers have already been written in one of the handlers, do not proceed
		return hw.data, false
	} else {
		return hw.data, true
	}
}

type fragmentWriter struct {
	http.ResponseWriter
	handler     Handler
	datamap     map[string]interface{}
	data        interface{}
	dataCalled  bool
	wroteHeader bool
}

func (fw *fragmentWriter) Write(b []byte) (int, error) {
	fw.wroteHeader = true
	return fw.ResponseWriter.Write(b)
}

func (fw *fragmentWriter) WriteHeader(code int) {
	fw.wroteHeader = true
	fw.ResponseWriter.WriteHeader(code)
}

func (fw *fragmentWriter) Data(d interface{}) {
	fw.data = d
	fw.dataCalled = true
}

func (fw *fragmentWriter) Delegate(name string, r *http.Request) (interface{}, bool) {
	if fw.wroteHeader {
		// response has already been written, nothing to do
		return nil, false
	}

	d, ok := fw.datamap[name]
	return d, ok
}

// implements treetop.go:DataWriter interface
type partialWriter struct {
	http.ResponseWriter
	handler     Partial
	handlerMap  map[Block]Partial
	data        interface{}
	dataCalled  bool
	wroteHeader bool
}

func (dw *partialWriter) Write(b []byte) (int, error) {
	dw.wroteHeader = true
	return dw.ResponseWriter.Write(b)
}

func (dw *partialWriter) WriteHeader(code int) {
	dw.wroteHeader = true
	dw.ResponseWriter.WriteHeader(code)
}

func (dw *partialWriter) Data(d interface{}) {
	dw.data = d
	dw.dataCalled = true
}

func (dw *partialWriter) Delegate(name string, r *http.Request) (interface{}, bool) {
	if dw.wroteHeader {
		// response has already been written, nothing to do
		return nil, false
	}

	block, ok := dw.handler.GetBlocks()[name]
	if !ok {
		// TODO: Add better error logging/handling and make sure this wont cause issues elsewhere!!!
		http.Error(dw, fmt.Sprintf("Unable to delegate to a handler that has not been defined '%s'", name), 500)
		return nil, false
	}

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

	if dw2.wroteHeader {
		dw.wroteHeader = true
		return nil, false
	}

	if dw2.dataCalled {
		return dw2.data, true
	} else {
		return nil, false
	}
}
