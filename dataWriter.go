package treetop

import (
	"errors"
	"io"
	"net/http"
)

var errRespWritten = errors.New("Response headers have already been written")

// used to capture the status of the response
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusRecorder) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	return n, err
}

type dataWriter struct {
	writer          *statusRecorder
	responseId      uint32
	responseWritten bool
	dataCalled      bool
	data            interface{}
	status          int
	partial         *Partial
	done            <-chan struct{}
}

// Implement http.ResponseWriter interface by delegating to embedded instance
func (dw *dataWriter) Header() http.Header {
	return dw.writer.Header()
}

// Implement http.ResponseWriter interface by delegating to embedded instance
func (dw *dataWriter) Write(b []byte) (int, error) {
	dw.responseWritten = true
	return dw.writer.Write(b)
}

// Implement http.ResponseWriter interface by delegating to embedded instance
func (dw *dataWriter) WriteHeader(statusCode int) {
	dw.responseWritten = true
	dw.writer.WriteHeader(statusCode)
}

// Handler pass down data for template execution
// If called multiple times on the final call will to observed
func (dw *dataWriter) Data(d interface{}) {
	dw.dataCalled = true
	dw.data = d
}

// Indicate what the HTTP status should be for the response.
//
// Note that when different handlers indicate a different status
// the code with the greater numeric value is chosen.
//
// For example, given: Bad Request, Unauthorized and Internal Server Error.
// Status values are differentiated as follows, 400 < 401 < 500, so 'Internal Server Error' wins!
func (dw *dataWriter) Status(status int) {
	if status > dw.status {
		dw.status = status
	}
}

// Load data from a named child block handler.
//
// The second return value indicates whether the delegated handler called Data(...)
// or not. This is necessary to discern the meaning of a `nil` data value.
func (dw *dataWriter) BlockData(name string, req *http.Request) (interface{}, bool) {
	// don't do anything if a response has already been written
	if dw.responseWritten {
		return nil, false
	}
	var part *Partial
	// 1. loop children of the template
	for i := 0; i < len(dw.partial.Blocks); i++ {
		// 2. find a child-template which extends the named block
		if dw.partial.Blocks[i].Extends == name {
			part = &dw.partial.Blocks[i]
			break
		}
	}
	if part == nil {
		// a template which extends block name was not found, return nothing
		return nil, false
	}
	// 3. construct a sub dataWriter
	subWriter := dataWriter{
		writer:     dw.writer,
		responseId: dw.responseId,
		partial:    part,
		done:       dw.done,
	}
	// 4. invoke handler
	part.HandlerFunc(&subWriter, req)
	if subWriter.responseWritten {
		dw.responseWritten = true
		return nil, false
	}
	// 5. adopt status of sub handler (if applicable, see .Status doc)
	dw.Status(subWriter.status)
	// 6. return resulting data and flag indicating if .Data(...) was called
	if subWriter.dataCalled {
		return subWriter.data, true
	} else {
		return nil, false
	}
}

// context which allows request to be cancelled
func (dw *dataWriter) Done() <-chan int {
	status := make(chan int, 1)
	go func() {
		<-dw.done
		status <- dw.writer.status
		close(status)
	}()
	return status
}

// Locally unique ID for Treetop HTTP response. This is intended to be used to keep track of
// the request as is passes between handlers.
func (dw *dataWriter) ResponseId() uint32 {
	return dw.responseId
}

// Load data from handlers hierarchy and execute template. Body will be written to IO writer passed in.
func (dw *dataWriter) execute(body io.Writer, exec TemplateExec, req *http.Request) error {
	dw.partial.HandlerFunc(dw, req)
	if dw.responseWritten {
		// response headers were already sent by one of the handlers, nothing left to do
		return errRespWritten
	}

	// Topo-sort of templates connected via blocks. The order is important for how template inheritance is resolved.
	// TODO: The result should not change between requests so cache it when the handler instance is created.
	templates, err := dw.partial.TemplateList()
	if err != nil {
		return err
	}

	// execute the templates with data loaded from handlers
	if tplErr := exec(body, templates, dw.data); tplErr != nil {
		return tplErr
	}
	return nil
}
