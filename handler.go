package treetop

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// pool of buffers used for executing HTML templates to completion
// before writing response headers and content
var buffers = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// obtain exclusive use of a buffer resource
func getBuffer() *bytes.Buffer {
	v := buffers.Get()
	buf, ok := v.(*bytes.Buffer)
	if !ok {
		panic(fmt.Sprintf("Expecting a Buffer but got %v", v))
	}
	return buf
}

// reset a buffer and make it eligible for either being returned to the
// pool or being deallocated
func releaseBuffer(buf *bytes.Buffer) {
	buf.Reset()
	buffers.Put(buf)
}

// ViewHandlerFunc is the interface for treetop handler functions that support hierarchical
// partial data loading.
type ViewHandlerFunc func(Response, *http.Request) interface{}

// ViewHandler is an extension of the http.Handler interface with methods added
// for extra treetop endpoint configuration
type ViewHandler interface {
	http.Handler
	FragmentOnly() ViewHandler
	PageOnly() ViewHandler
}

// Errors used by the TemplateHandler.
var (
	// ErrNotAcceptable is produced by ServeHTTP when a request
	// does not contain an accept header that can be handled by this endpoint
	ErrNotAcceptable = errors.New(
		"treetop template handler: server cannot produce a response matching the list of acceptable values")
)

// Template is an interface so that the concrete template implementation can be changed
type Template interface {
	ExecuteTemplate(io.Writer, string, interface{}) error
}

// TemplateHandler implements the treetop.ViewHandler interface for endpoints that support the treetop protocol
type TemplateHandler struct {
	Page             *View
	PageTemplate     Template
	Partial          *View
	PartialTemplate  Template
	Includes         []*View
	IncludeTemplates []Template
	// optional developer defined error handler
	ServeTemplateError func(error, Response, *http.Request)
}

// FragmentOnly creates a new Handler that only responds to fragment requests
func (h *TemplateHandler) FragmentOnly() ViewHandler {
	return &TemplateHandler{
		Partial:          h.Partial,
		Includes:         h.Includes,
		PartialTemplate:  h.PartialTemplate,
		IncludeTemplates: h.IncludeTemplates,
	}
}

// PageOnly create a new handler that will only respond to non-fragment (full page) requests
func (h *TemplateHandler) PageOnly() ViewHandler {
	return &TemplateHandler{
		Page:         h.Page,
		PageTemplate: h.PageTemplate,
	}
}

// ServeHTTP is responsible for directing the handing of an incoming request.
// Implements the procedure through which views functions and templates
// are to be executed.
func (h *TemplateHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	resp := BeginResponse(req.Context(), w)
	defer resp.Cancel()

	if IsTemplateRequest(req) {
		if h.Page != nil {
			// since a page view exists for this handler, use the request
			// URI as the designated page URL
			resp.DesignatePageURL(req.URL.RequestURI())
		}
		// render treetop template, application/x.treetop-html-template+xml
		h.serveTemplateRequest(resp, req)
	} else {
		// render a HTML document, text/html
		h.servePageRequest(resp, req)
	}
}

// servePageRequest will render a HTML document using hierarchial handlers and templates
func (h *TemplateHandler) servePageRequest(resp *ResponseWrapper, req *http.Request) {
	errlog := h.newResponseErrorLog(resp, req)

	// response body will be buffered before being written to the connection
	// to avoid torn writes as a result of errors
	buf := getBuffer()
	defer releaseBuffer(buf)

	// This is a full page request,
	// execute the page view handlers and templates
	if h.Page == nil {
		errlog(ErrNotAcceptable)
		return
	}
	data := h.Page.HandlerFunc(resp.WithSubViews(h.Page.SubViews), req)
	if resp.Finished() {
		return
	}
	err := h.PageTemplate.ExecuteTemplate(buf, h.Page.Defines, data)
	if err != nil {
		errlog(err)
		return
	}

	// set content length from write buffer
	resp.Header().Set("Content-Length", strconv.Itoa(buf.Len()))

	if h.Partial != nil {
		// inform cache that another content type is possible for this endpoint
		resp.Header().Set("Vary", "Accept")
	}

	// set content type as standard html mimetype
	resp.Header().Set("Content-Type", "text/html")

	if status := resp.Status(0); status > 0 {
		// response instance was given a status code,
		// write the status, finalizing the headers
		resp.WriteHeader(status)
	}

	// copy from buffer to the connection writer
	if _, err := io.Copy(resp, buf); err != nil {
		// It is likely that the header has been written at this stage,
		// hence there is no ability to notify the client of this error.
		log.Printf("treetop: page write error %s", err)

		// This will be ignored if the header was sent
		http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	return
}

// serverTemplateRequest will execute the partial along with each postscript handler in order
// then append the postscript HTML to partial HTML
func (h *TemplateHandler) serveTemplateRequest(resp *ResponseWrapper, req *http.Request) {
	errlog := h.newResponseErrorLog(resp, req)

	// This is a template request,
	if h.Partial == nil {
		errlog(ErrNotAcceptable)
		return
	}

	// response body will be buffered before being written to the connection
	// to avoid torn writes as a result of errors
	buf := getBuffer()
	defer releaseBuffer(buf)

	var (
		views = append([]*View{h.Partial}, h.Includes...)
		data  = make([]interface{}, len(views))
		tmpls = append([]Template{h.PartialTemplate}, h.IncludeTemplates...)
	)

	// call handler for partial and each postscript view. Collect template data.
	for i, view := range views {
		if view == nil {
			continue
		}
		data[i] = view.HandlerFunc(resp.WithSubViews(view.SubViews), req)
		if resp.Finished() {
			return
		}
	}
	// write opening template tag
	buf.WriteString("<template>\n")

	// execute partial and postscript templates with data collected from handler funcs
	for i, tmpl := range tmpls {
		if i > 0 {
			buf.WriteByte('\n')
		}
		err := tmpl.ExecuteTemplate(buf, views[i].Defines, data[i])
		if err != nil {
			if h.ServeTemplateError != nil {
				h.ServeTemplateError(err, resp, req)
			} else {
				// use log pkg standard logger
				log.Printf("treetop: partial template execute error %s", err)
				http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}
	}
	// write closing template tag
	buf.WriteString("\n</template>")

	// use treetop writer so that headers will assigned for XHR client
	ttW, ok := resp.NewTemplateWriter(req)
	if !ok {
		http.Error(resp, http.StatusText(http.StatusNotAcceptable), http.StatusNotAcceptable)
		return
	}

	// set content length from write buffer
	resp.Header().Set("Content-Length", strconv.Itoa(buf.Len()))

	// copy from buffer to the connection
	if _, err := io.Copy(ttW, buf); err != nil {
		// It is likely that the header has been written at this stage,
		// hence there is no ability to notify the client of this error.
		log.Printf("treetop: partial write error %s", err)

		// This will be ignored if the header was sent
		http.Error(resp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// newResponseErrorLog create an error handler for a supplied response and request instance
func (h *TemplateHandler) newResponseErrorLog(rsp Response, req *http.Request) func(err error) {
	return func(err error) {
		if h.ServeTemplateError != nil {
			if err == ErrNotAcceptable {
				rsp.Status(http.StatusNotAcceptable)
			}
			// delegate handing to user defined function
			h.ServeTemplateError(err, rsp, req)
			return
		}

		if err == ErrNotAcceptable {
			http.Error(rsp, http.StatusText(http.StatusNotAcceptable), http.StatusNotAcceptable)
			return
		}
		// use log pkg standard logger
		log.Printf("treetop template handler error, %s", err)
		http.Error(rsp, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
