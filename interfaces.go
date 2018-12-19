package treetop

import (
	"context"
	"io"
	"net/http"
)

// Response is the HandlerFunc facade for the treetop request response process.
// It can be used to delegate handling to block and to track the resolution of a
// single request across multiple handlers.
type Response interface {
	http.ResponseWriter
	Status(int) int
	Done() bool
	HandlePartial(string, *http.Request) interface{}
	ResponseID() uint32
	Context() context.Context
}

// TemplateExec interface is a signature of a function which can be configured to
// execute the templates with supplied data.
type TemplateExec func(io.Writer, []string, interface{}) error

// HandlerFunc is the interface for treetop handler functions that support hierarchical
// partial data loading.
type HandlerFunc func(Response, *http.Request) interface{}

// View is an interface for a hierarchical view configuration. Named child views can be
// added and a http.Handler instance can be derived.
type View interface {
	SubView(string, string, HandlerFunc) View
	DefaultSubView(string, string, HandlerFunc) View
	Handler() *Handler
}
