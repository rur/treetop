package treetop

import (
	"context"
	"io"
	"net/http"
)

type Response interface {
	http.ResponseWriter
	Data(interface{})
	Status(int)
	Delegate(string, *http.Request) (interface{}, bool)
	DelegateWithDefault(string, *http.Request, interface{}) interface{}
	ResponseId() uint32
	Context() context.Context
}

type TemplateExec func(io.Writer, []string, interface{}) error
type HandlerFunc func(Response, *http.Request)

type View interface {
	SubView(string, string, HandlerFunc) View
	DefaultSubView(string, string, HandlerFunc) View
	PageHandler() *Handler
	PartialHandler() *Handler
	FragmentHandler() *Handler
}
