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

type PartialDef interface {
	Block(string) BlockDef
	PageHandler() *Handler
	PartialHandler() *Handler
	FragmentHandler() *Handler
}
type BlockDef interface {
	Extend(string, HandlerFunc) PartialDef
	Default(string, HandlerFunc) PartialDef
}
