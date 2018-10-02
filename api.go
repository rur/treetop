package treetop

import "net/http"

type DataWriter interface {
	http.ResponseWriter
	Data(interface{})
	Status(int)
	BlockData(string, *http.Request) (interface{}, bool)
	ResponseToken() string
}

type RenderFunc func([]string, interface{}) error
type HandlerFunc func(DataWriter, *http.Request)

type TemplateDef interface {
	Block(string) BlockDef
	PartialHandler() *Handler
	FragmentHandler() *Handler
}
type BlockDef interface {
	Extend(string, HandlerFunc) TemplateDef
	Default(string, HandlerFunc) TemplateDef
}

type Template struct {
	Extends     string
	Content     string
	HandlerFunc HandlerFunc
	Parent      *Template
	Blocks      []*Template
}
