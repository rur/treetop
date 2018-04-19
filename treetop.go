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
	// Set the data to be passed to the template
	Data(interface{})
	// Indicate the http status of the response, since there are potentially
	// multiple handling functions per request this will not necessarily be the final
	// status code for the response. Given multiple valid calls to Status during a request,
	// the upper bound of the ordered set (Statuses, ≤) will be chosen for the response.
	// If no valid status is specified, the http.ResponseWriter default is 200 OK.
	Status(int)
	// load data from specified child handler
	PartialData(string, *http.Request) (interface{}, bool)
}

type Response struct {
	Data   interface{}
	Status int
}

type Block interface {
	fmt.Stringer
	Default() Partial
	SetDefault(Partial) Block
	Fragment(string, HandlerFunc) Fragment
	Partial(string, HandlerFunc) Partial
	DefaultPartial(string, HandlerFunc) Partial
	Container() Partial
}

type Fragment interface {
	fmt.Stringer
	http.Handler
	Func() HandlerFunc
	Template() string
	Extends() Block
}

type Partial interface {
	Fragment
	Block(string) Block
	GetBlocks() map[string]Block
	Includes(...Partial) Partial
	GetIncludes() map[Block]Partial
}

type Renderer interface {
	Page(string, HandlerFunc) Partial
	Fragment(string, HandlerFunc) Fragment
	Append(Partial, ...Fragment) http.Handler
}

type HandlerFunc func(DataWriter, *http.Request)

type TemplateExec func(io.Writer, []string, interface{}) error
