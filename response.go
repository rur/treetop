package treetop

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
)

var (
	token               uint32
	ErrResponseHijacked = errors.New(
		"treetop response: cannot write, HTTP response has been hijacked by another handler")
)

// nextResponseID generates a token which can be used to identify treetop
// responses *locally*. The only uniqueness requirement
// is that concurrent active requests must not possess the same value.
func nextResponseID() uint32 {
	return atomic.AddUint32(&token, 1)
}

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

	// HandleSubView loads data from a named child subview handler. If no handler is available for the name,
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

// ResponseWrapper is the concrete implementation of the response writer wrapper
// supplied to view handler functions
type ResponseWrapper struct {
	http.ResponseWriter
	responseID       uint32
	context          context.Context
	status           int
	subViews         map[string]*View
	pageURL          string
	pageURLSpecified bool
	replaceURL       bool
	cancel           context.CancelFunc
	derivedFrom      *ResponseWrapper
	hijacked         bool
}

// BeginResponse initializes the context for a treetop request response
func BeginResponse(cxt context.Context, w http.ResponseWriter) *ResponseWrapper {
	rsp := ResponseWrapper{
		ResponseWriter: w,
		responseID:     nextResponseID(),
	}
	rsp.context, rsp.cancel = context.WithCancel(cxt)
	return &rsp
}

// WithSubViews creates a derived response wrapper for a different view, inheriting
// request
func (rsp *ResponseWrapper) WithSubViews(subViews map[string]*View) *ResponseWrapper {
	derived := ResponseWrapper{
		ResponseWriter: rsp.ResponseWriter,
		responseID:     rsp.responseID,
		subViews:       make(map[string]*View),
		context:        rsp.context,
		cancel:         rsp.cancel,
		derivedFrom:    rsp,
		hijacked:       rsp.hijacked,
	}
	if subViews != nil {
		// some defensive copying here
		for k, v := range subViews {
			derived.subViews[k] = v
		}
	}
	return &derived
}

// NewTemplateWriter will return a template Writer configured to add Treetop headers
// based up on the state of the response. If the request is not a template request
// the writer will be nil and the ok flag will be false
func (rsp *ResponseWrapper) NewTemplateWriter(req *http.Request) (Writer, bool) {
	if rsp.Finished() {
		return nil, false
	}
	ttW, ok := NewFragmentWriter(rsp.ResponseWriter, req)
	if !ok {
		return nil, false
	}
	if rsp.pageURLSpecified {
		if rsp.replaceURL {
			ttW.ReplacePageURL(rsp.pageURL)
		} else {
			ttW.DesignatePageURL(rsp.pageURL)
		}
	}
	if rsp.status > 0 {
		ttW.Status(rsp.status)
	}
	return ttW, true
}

// Cancel will teardown the treetop handing process
func (rsp *ResponseWrapper) Cancel() {
	rsp.cancel()
}

// markHijacked prevents this response wrapper instance from attempting to
// write to the underlying http.ResponseWriter. This will be called by derived
// response wrappers.
func (rsp *ResponseWrapper) markHijacked() {
	if rsp == nil {
		return
	}
	rsp.hijacked = true
	// mark all responses to the root handler
	rsp.derivedFrom.markHijacked()
}

// Write delegates to the underlying ResponseWriter while aborting the
// treetop executor handler.
func (rsp *ResponseWrapper) Write(b []byte) (int, error) {
	if rsp.hijacked {
		return 0, ErrResponseHijacked
	}
	rsp.Cancel()
	// prevent parent handler attempting to hijack the response
	rsp.derivedFrom.markHijacked()
	return rsp.ResponseWriter.Write(b)
}

// WriteHeader delegates to the underlying ResponseWriter while setting finished flag to true
func (rsp *ResponseWrapper) WriteHeader(statusCode int) {
	if rsp.Finished() {
		// ignore erroneous calls to WriteHeader if the response is finished
		return
	}
	rsp.Cancel()
	// prevent parent handler attempting to hijack the response
	rsp.derivedFrom.markHijacked()
	rsp.ResponseWriter.WriteHeader(statusCode)
}

// Status will set a status for the treetop response headers
// if a response status has been set previously, the larger
// code value will be adopted
func (rsp *ResponseWrapper) Status(status int) int {
	if rsp == nil {
		return 0
	}
	if status > rsp.status {
		rsp.status = status
	}
	// propegate status to root handler
	rsp.derivedFrom.Status(status)
	return rsp.status
}

// ReplacePageURL will instruct the client to replace the current
// history entry with the supplied URL
func (rsp *ResponseWrapper) ReplacePageURL(url string) {
	if rsp == nil {
		return
	}
	rsp.pageURL = url
	rsp.replaceURL = true
	rsp.pageURLSpecified = true

	// propegate url to root handler
	rsp.derivedFrom.ReplacePageURL(url)
}

// DesignatePageURL will result in a header being added to the response
// that will create a new history entry for the supplied URL
func (rsp *ResponseWrapper) DesignatePageURL(url string) {
	if rsp == nil {
		return
	}
	rsp.pageURL = url
	rsp.replaceURL = false
	rsp.pageURLSpecified = true

	// propegate url to root handler
	rsp.derivedFrom.DesignatePageURL(url)
}

// Finished will return true if the response headers have been written to the
// client, effectively cancelling the treetop view handler lifecycle
func (rsp *ResponseWrapper) Finished() bool {
	if rsp == nil {
		return true
	}
	select {
	case <-rsp.context.Done():
		return true
	default:
		return false
	}
}

// HandleSubView will execute the handler for a specified sub view of the current view
// if there is no match for the name, nil will be returned.
func (rsp *ResponseWrapper) HandleSubView(name string, req *http.Request) interface{} {
	// don't do anything if a response has already been written
	if rsp.Finished() || len(rsp.subViews) == 0 {
		return nil
	}

	sub, ok := rsp.subViews[name]
	if !ok || sub == nil {
		return nil
	}

	subResp := rsp.WithSubViews(sub.SubViews)

	// Invoke sub handler, collecting the response
	return sub.HandlerFunc(subResp, req)
}

// Context is getter for the treetop response context which will indicate when the request
// has been completed as was cancelled. This is derived from the request context so
// it can safely be used for cleanup.
func (rsp *ResponseWrapper) Context() context.Context {
	return rsp.context
}

// ResponseID is a getter which returns a locally unique ID for a Treetop HTTP response.
// This is intended to be used to keep track of the request as is passes between handlers.
// The ID will increment by one starting at zero, every time the server is restarted.
func (rsp *ResponseWrapper) ResponseID() uint32 {
	return rsp.responseID
}
