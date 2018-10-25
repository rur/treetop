package treetop

import (
	"context"
	"io"
	"net/http"
)

type DataWriter interface {
	http.ResponseWriter
	Data(interface{})
	Status(int)
	BlockData(string, *http.Request) (interface{}, bool)
	ResponseId() uint32
	Context() context.Context
}

type TemplateExec func(io.Writer, []string, interface{}) error
type HandlerFunc func(DataWriter, *http.Request)

type View interface {
	SubView(string, string, HandlerFunc) View
	DefaultSubView(string, string, HandlerFunc) View
	PageHandler() *Handler
	PartialHandler() *Handler
	FragmentHandler() *Handler
}