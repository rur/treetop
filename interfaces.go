package treetop

import (
	"context"
	"io"
	"net/http"
)

type Response interface {
	http.ResponseWriter
	Status(int) int
	Done() bool
	HandlePartial(string, *http.Request) interface{}
	ResponseID() uint32
	Context() context.Context
}

type TemplateExec func(io.Writer, []string, interface{}) error
type HandlerFunc func(Response, *http.Request) interface{}

type View interface {
	SubView(string, string, HandlerFunc) View
	DefaultSubView(string, string, HandlerFunc) View
	PageHandler() *Handler
	PartialHandler() *Handler
	FragmentHandler() *Handler
}
