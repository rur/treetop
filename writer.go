package treetop

import "net/http"

type dataWriter struct {
	writer          http.ResponseWriter
	responseToken   string
	responseWritten bool
	dataCalled      bool
	data            interface{}
	status          int
	template        *Template
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

// Indicate what the response status should be for the request.
//
// if the new status code is a larger value than current status
// then the greater numeric status is chosen.
//
// eg. Given: Bad Request, Unauthorized and Internal Server Error (400 < 401 < 500), server error wins!
func (dw *dataWriter) Status(status int) {
	if status > dw.status {
		dw.status = status
	}
}

// Load data from a named child block handler
func (dw *dataWriter) BlockData(name string, req *http.Request) (interface{}, bool) {
	// don't do anything if a response has already been written
	if dw.responseWritten {
		return nil, false
	}
	var templ *Template
	// 1. loop children of the template
	for i := 0; i < len(dw.template.Blocks); i++ {
		// 2. find a child-template which extends the named block
		if dw.template.Blocks[i].Extends == name {
			templ = dw.template.Blocks[i]
			break
		}
	}
	if templ == nil {
		// a template which extends block name was not found, return nothing
		return nil, false
	}
	// 3. construct a sub dataWriter
	subWriter := dataWriter{
		writer:        dw.writer,
		responseToken: dw.responseToken,
		template:      templ,
	}
	// 4. invoke handler
	templ.HandlerFunc(&subWriter, req)
	// 5. adopt status of sub handler (if applicable, see .Status doc)
	dw.Status(subWriter.status)
	// 6. return resulting data and flag indicating if .Data(...) was called
	if subWriter.dataCalled {
		return subWriter.data, true
	} else {
		return nil, false
	}
}

// unique ID for Treetop HTTP response, can be used to keep track of the request
// as is passes between handlers
func (dw *dataWriter) ResponseToken() string {
	return dw.responseToken
}
