package treetop

import (
	"context"
	"errors"
	"io"
	"net/http"
)

var errRespWritten = errors.New("Response headers have already been written")

type responseImpl struct {
	http.ResponseWriter
	responseId      uint32
	context         context.Context
	responseWritten bool
	dataCalled      bool
	data            interface{}
	status          int
	partial         *Partial
}

// Implement http.ResponseWriter interface by delegating to embedded instance
func (rsp *responseImpl) Header() http.Header {
	return rsp.ResponseWriter.Header()
}

// Implement http.ResponseWriter interface by delegating to embedded instance
func (rsp *responseImpl) Write(b []byte) (int, error) {
	rsp.responseWritten = true
	return rsp.ResponseWriter.Write(b)
}

// Implement http.ResponseWriter interface by delegating to embedded instance
func (rsp *responseImpl) WriteHeader(statusCode int) {
	rsp.responseWritten = true
	rsp.ResponseWriter.WriteHeader(statusCode)
}

// Handler pass down data for template execution
// If called multiple times on the final call will to observed
func (rsp *responseImpl) Data(d interface{}) {
	rsp.dataCalled = true
	rsp.data = d
}

// Indicate what the HTTP status should be for the response.
//
// Note that when different handlers indicate a different status
// the code with the greater numeric value is chosen.
//
// For example, given: Bad Request, Unauthorized and Internal Server Error.
// Status values are differentiated as follows, 400 < 401 < 500, so 'Internal Server Error' wins!
func (rsp *responseImpl) Status(status int) {
	if status > rsp.status {
		rsp.status = status
	}
}

// Load data from a named child block handler.
//
// The second return value indicates whether the delegated handler called Data(...)
// or not. This is necessary to discern the meaning of a `nil` data value.
func (rsp *responseImpl) Delegate(name string, req *http.Request) (interface{}, bool) {
	// don't do anything if a response has already been written
	if rsp.responseWritten {
		return nil, false
	}
	var part *Partial
	// 1. loop children of the template
	for i := 0; i < len(rsp.partial.Blocks); i++ {
		// 2. find a child-template which extends the named block
		if rsp.partial.Blocks[i].Extends == name {
			part = &rsp.partial.Blocks[i]
			break
		}
	}
	if part == nil {
		// a template which extends block name was not found, return nothing
		return nil, false
	}
	// 3. construct a sub responseImpl
	subWriter := responseImpl{
		ResponseWriter: rsp.ResponseWriter,
		responseId:     rsp.responseId,
		context:        rsp.context,
		partial:        part,
	}
	// 4. invoke handler
	part.HandlerFunc(&subWriter, req)
	if subWriter.responseWritten {
		rsp.responseWritten = true
		return nil, false
	}
	// 5. adopt status of sub handler (if applicable, see .Status doc)
	rsp.Status(subWriter.status)
	// 6. return resulting data and flag indicating if .Data(...) was called
	if subWriter.dataCalled {
		return subWriter.data, true
	} else {
		return nil, false
	}
}

func (rsp *responseImpl) DelegateWithDefault(name string, req *http.Request, dfl interface{}) interface{} {
	if data, ok := rsp.Delegate(name, req); !ok {
		return dfl
	} else {
		return data
	}
}

// context which allows request to be cancelled
func (rsp *responseImpl) Context() context.Context {
	return rsp.context
}

// Locally unique ID for Treetop HTTP response. This is intended to be used to keep track of
// the request as is passes between handlers.
func (rsp *responseImpl) ResponseId() uint32 {
	return rsp.responseId
}

// Load data from handlers hierarchy and execute template. Body will be written to IO writer passed in.
func (rsp *responseImpl) execute(body io.Writer, exec TemplateExec, req *http.Request) error {
	rsp.partial.HandlerFunc(rsp, req)
	if rsp.responseWritten {
		// response headers were already sent by one of the handlers, nothing left to do
		return errRespWritten
	}

	// Topo-sort of templates connected via blocks. The order is important for how template inheritance is resolved.
	// TODO: The result should not change between requests so cache it when the handler instance is created.
	templates, err := rsp.partial.TemplateList()
	if err != nil {
		return err
	}

	// execute the templates with data loaded from handlers
	if tplErr := exec(body, templates, rsp.data); tplErr != nil {
		return tplErr
	}
	return nil
}
