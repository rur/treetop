package treetop

import (
	"context"
	"io"
	"net/http"
)

// Response extends the http.ResponseWriter interface to give ViewHandelersFunc's limited
// ability to control the hierarchical request handling.
//
// Note that writing directly to the underlying ResponseWriter in the handler will cancel the
// treetop handling process. Taking control of response writing in this way is a very common and
// useful practice especially for error messages or redirects.
type Response interface {
	http.ResponseWriter

	// Status allows a handler to indicate (not determine) what the HTTP status
	// should be for the response.
	//
	// When different handlers indicate a different status,
	// the code with the greater numeric value is chosen.
	//
	// For example, given: Bad Request, Unauthorized and Internal Server Error.
	// Status values are differentiated as follows, 400 < 401 < 500,
	// 'Internal Server Error' is chosen for the response header.
	//
	// The resulting response status is returned. Getting the current status
	// without affecting the response can be done as follows
	//
	// 		status := rsp.Status(0)
	//
	Status(int) int

	// DesignatePageURL forces the response to be handled as a navigation event with a specified URL.
	// The browser will have a new history entry created for the supplied URL.
	DesignatePageURL(string)

	// ReplacePageURL forces the location bar in the web browser to be updated with the supplied
	// URL. This should be done by *replacing* the existing history entry. (not adding a new one)
	ReplacePageURL(string)

	// Finished will return true if a handler has taken direct responsibility for writing the
	// response.
	Finished() bool

	// HandleSubView loads data from a named child subview handler. If no handler is availabe for the name,
	// nil will be returned.
	//
	// NOTE: Since a sub handler may have returned nil, there is no way for the parent handler to determine
	//       whether the name resolved to a concrete view.
	HandleSubView(string, *http.Request) interface{}

	// ResponseID returns the ID treetop has associated with this request.
	// Since multiple handlers may be involved, the ID is useful for logging and caching.
	//
	// Response IDs avoid potential pitfalls around Request instance comparison that can affect middleware.
	//
	// NOTE: This is *not* a UUID, response IDs are incremented from zero when the server is started
	ResponseID() uint32

	// Context returns the context associated with the treetop process.
	// This is a child of the http Request context.
	Context() context.Context
}

// responseImpl is the API that treetop request handlers interact with
// through the treetop.ResponseWriter interface.
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

// Write delegates to the underlying ResponseWriter while setting finished flag to true
func (rsp *responseImpl) Write(b []byte) (int, error) {
	rsp.finished = true
	return rsp.ResponseWriter.Write(b)
}

// WriteHeader delegates to the underlying ResponseWriter while setting finished flag to true
func (rsp *responseImpl) WriteHeader(statusCode int) {
	rsp.finished = true
	rsp.ResponseWriter.WriteHeader(statusCode)
}

func (rsp *responseImpl) Status(status int) int {
	if status > rsp.status {
		rsp.status = status
	}
	return rsp.status
}

func (rsp *responseImpl) ReplacePageURL(url string) {
	rsp.pageURL = url
	rsp.replaceURL = true
	rsp.pageURLSpecified = true
}

func (rsp *responseImpl) DesignatePageURL(url string) {
	rsp.pageURL = url
	rsp.replaceURL = false
	rsp.pageURLSpecified = true
}

func (rsp *responseImpl) Finished() bool {
	return rsp.finished
}

func (rsp *responseImpl) HandleSubView(name string, req *http.Request) interface{} {
	// don't do anything if a response has already been written
	if rsp.finished {
		return nil
	}
	var part *Partial

	// 1. loop through direct subviews
	for i := range rsp.partial.Blocks {
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
