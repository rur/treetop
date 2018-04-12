package treetop

import (
	"fmt"
	"io"
	"net/http"
)

const (
	PartialContentType  = "application/x.treetop-html-partial+xml"
	FragmentContentType = "application/x.treetop-html-fragment+xml"
)

type DataWriter interface {
	http.ResponseWriter
	// render the template with the supplied data
	Data(interface{})
	// load data from specified child handler
	Delegate(string, *http.Request) (interface{}, bool)
}

type Block interface {
	fmt.Stringer
	WithDefault(string, HandlerFunc) Block
	SetDefault(Partial) Block
	Default() Partial
	Extend(string, HandlerFunc) Partial
	Container() Partial
}

type Handler interface {
	fmt.Stringer
	http.Handler
	Func() HandlerFunc
	Template() string
}

type Partial interface {
	Handler
	DefineBlock(string) Block
	Extends() Block
	GetBlocks() map[string]Block
	Includes(...Partial) Partial
	GetIncludes() map[Block]Partial
}

type Renderer interface {
	NewPage(string, HandlerFunc) Partial
	NewFragment(string, HandlerFunc) Handler
}

type HandlerFunc func(DataWriter, *http.Request)

type TemplateExec func(io.Writer, []string, interface{}) error
