package treetop

import (
	"io"
	"net/http"
)

const (
	PartialContentType  = "application/x.treetop-html-partial+xml"
	FragmentContentType = "application/x.treetop-html-fragment+xml"
)

type DataWriter interface {
	http.ResponseWriter
	Data(interface{})
	Status(int)
	BlockData(string, *http.Request) (interface{}, bool)
	ResponseId() uint32
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

type Renderer interface {
	Define(string, HandlerFunc) PartialDef
}
