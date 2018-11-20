package treetop

import (
	"context"
	"io"
	"net/http"
)

type responseImpl struct {
	http.ResponseWriter
	responseID uint32
	context    context.Context
	finished   bool
	status     int
	partial    *Partial
}

// Implement http.ResponseWriter interface by delegating to embedded instance
func (rsp *responseImpl) Header() http.Header {
	return rsp.ResponseWriter.Header()
}

// Implement http.ResponseWriter interface by delegating to embedded instance
func (rsp *responseImpl) Write(b []byte) (int, error) {
	rsp.finished = true
	return rsp.ResponseWriter.Write(b)
}

// Implement http.ResponseWriter interface by delegating to embedded instance
func (rsp *responseImpl) WriteHeader(statusCode int) {
	rsp.finished = true
	rsp.ResponseWriter.WriteHeader(statusCode)
}

// Indicate what the HTTP status should be for the response.
//
// Note that when different handlers indicate a different status
// the code with the greater numeric value is chosen.
//
// For example, given: Bad Request, Unauthorized and Internal Server Error.
// Status values are differentiated as follows, 400 < 401 < 500, so 'Internal Server Error' wins!
func (rsp *responseImpl) Status(status int) int {
	if status > rsp.status {
		rsp.status = status
	}
	return rsp.status
}

func (rsp *responseImpl) Done() bool {
	return rsp.finished
}

// Load data from a named child block handler.
//
// The second return value indicates whether the delegated handler called Data(...)
// or not. This is necessary to discern the meaning of a `nil` data value.
func (rsp *responseImpl) HandlePartial(name string, req *http.Request) interface{} {
	// don't do anything if a response has already been written
	if rsp.finished {
		return nil
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
		return nil
	}
	// 3. construct a sub responseImpl
	subResp := responseImpl{
		ResponseWriter: rsp.ResponseWriter,
		responseID:     rsp.responseID,
		context:        rsp.context,
		partial:        part,
	}
	// 4. invoke handler
	data := part.HandlerFunc(&subResp, req)
	if subResp.finished {
		rsp.finished = true
		return nil
	} else {
		subResp.finished = true
	}
	// 5. adopt status of sub handler (if applicable, see .Status doc)
	rsp.Status(subResp.status)
	// 6. return resulting data and flag indicating if .Data(...) was called
	return data
}

// context which allows request to be cancelled
func (rsp *responseImpl) Context() context.Context {
	return rsp.context
}

// Locally unique ID for Treetop HTTP response. This is intended to be used to keep track of
// the request as is passes between handlers.
func (rsp *responseImpl) ResponseID() uint32 {
	return rsp.responseID
}

// Load data from handlers hierarchy and execute template. Body will be written to IO writer passed in.
func (rsp *responseImpl) execute(body io.Writer, exec TemplateExec, req *http.Request) error {
	data := rsp.partial.HandlerFunc(rsp, req)
	if rsp.finished {
		// response headers were already sent by one of the handlers, nothing left to do
		return nil
	}

	// Topo-sort of templates connected via blocks. The order is important for how template inheritance is resolved.
	// TODO: The result should not change between requests so cache it when the handler instance is created.
	templates, err := rsp.partial.TemplateList()
	if err != nil {
		return err
	}

	// execute the templates with data loaded from handlers
	if tplErr := exec(body, templates, data); tplErr != nil {
		return tplErr
	}
	return nil
}
