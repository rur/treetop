package treetop

import "net/http"

type dataWriter struct {
	writer          http.ResponseWriter
	localToken      uint32
	responseWritten bool
	dataCalled      bool
	data            interface{}
	status          int
	partial         *Partial
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
			part = dw.partial.Blocks[i]
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
		localToken: dw.localToken,
		partial:    part,
	}
	// 4. invoke handler
	part.HandlerFunc(&subWriter, req)
	// 5. adopt status of sub handler (if applicable, see .Status doc)
	dw.Status(subWriter.status)
	// 6. return resulting data and flag indicating if .Data(...) was called
	if subWriter.dataCalled {
		return subWriter.data, true
	} else {
		return nil, false
	}
}

// Locally unique ID for Treetop HTTP response. This is intended to be used to keep track of
// the request as is passes between handlers.
func (dw *dataWriter) LocalToken() uint32 {
	return dw.localToken
}
