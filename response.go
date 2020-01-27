package treetop

import (
	"context"
	"io"
	"net/http"
)

type responseImpl struct {
	http.ResponseWriter
	responseID       uint32
	context          context.Context
	finished         bool
	status           int
	partial          *Partial
	pageURL          string
	pageURLSpecified bool
	replaceURL       bool
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

// ReplacePageURL forces the location bar in the web browser to be updated with the supplied
// URL. This should be done by *replacing* the existing history entry. (not adding a new one)
func (rsp *responseImpl) ReplacePageURL(url string) {
	rsp.pageURL = url
	rsp.replaceURL = true
	rsp.pageURLSpecified = true
}

// DesignatePageURL forces the response to be handled as a full or part of a full page.
// the web browser will have a new history entry for the supplied URL.
func (rsp *responseImpl) DesignatePageURL(url string) {
	rsp.pageURL = url
	rsp.replaceURL = false
	rsp.pageURLSpecified = true
}

// Done allows a handler to check if the request has already been satisfied.
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

	// 1. loop through direct subviews
	for i := 0; i < len(rsp.partial.Blocks); i++ {
		// find a subview which extends the named block
		if rsp.partial.Blocks[i].Extends == name {
			part = &rsp.partial.Blocks[i]
			break
		}
	}
	if part == nil {
		// a template which extends block name was not found, return nothing
		return nil
	}

	// 2. Construct a Response instance for the sub handler and inherit
	//    properties from the current response
	subResp := responseImpl{
		ResponseWriter: rsp.ResponseWriter,
		responseID:     rsp.responseID,
		context:        rsp.context,
		partial:        part,
	}

	// 3. invoke sub handler, collecting the response
	data := part.HandlerFunc(&subResp, req)
	if subResp.finished {
		// sub handler took over the writing of the response, all handlers should be halted.
		//
		// NOTE: It seems like this should invole the Golang 'Context' pattern in some way.
		//       It is not obvious to me but there is probably a better way to tear the
		//       whole process down. A simple 'finished' flag is fine for the time being.
		//
		rsp.finished = true
		return nil
	}
	subResp.finished = true

	// 4. Adopt status and page URL of sub handler (as applicable)
	rsp.Status(subResp.status)
	if subResp.pageURLSpecified {
		// adopt pageURL if the child handler specified one
		if subResp.replaceURL {
			rsp.ReplacePageURL(subResp.pageURL)
		} else {
			rsp.DesignatePageURL(subResp.pageURL)
		}
	}

	// 5. return resulting data for parent handler to use
	return data
}

// Context is getter for the treetop response context which will indicate when the request
// has been completed as was cancelled. This is derived from the request context so
// it can safely be used for cleanup.
func (rsp *responseImpl) Context() context.Context {
	return rsp.context
}

// ResponseID is a getter which returns a locally unique ID for a Treetop HTTP response.
// This is intended to be used to keep track of the request as is passes between handlers.
// The ID will increment by one starting at zero, every time the server is restarted.
func (rsp *responseImpl) ResponseID() uint32 {
	return rsp.responseID
}

// execute loads data from handlers hierarchy and executes the aggregated template list.
// Body will be written to IO writer passed in.
func (rsp *responseImpl) execute(body io.Writer, exec TemplateExec, req *http.Request) error {
	data := rsp.partial.HandlerFunc(rsp, req)
	if rsp.finished {
		// response headers were already sent by one of the handlers, nothing left to do
		return nil
	}

	// Toposort of templates connected via blocks. The order is important for how template inheritance is resolved.
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
