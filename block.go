package treetop

import "fmt"

type blockInternal struct {
	name           string
	defaultHandler Handler
	container      Handler
}

func (b *blockInternal) String() string {
	return fmt.Sprintf("<Block Name: '%s'>", b.name)
}

func (b *blockInternal) Name() string {
	return b.name
}

func (b *blockInternal) WithDefault(template string, handlerFunc HandlerFunc) Block {
	b.defaultHandler = b.Extend(template, handlerFunc)
	return b
}

func (b *blockInternal) SetDefault(h Handler) Block {
	b.defaultHandler = h
	return b
}

func (b *blockInternal) Default() Handler {
	return b.defaultHandler
}

func (b *blockInternal) Container() Handler {
	return b.container
}

func (b *blockInternal) Extend(template string, handlerFunc HandlerFunc) Partial {
	h := partialInternal{
		template:    template,
		handlerFunc: handlerFunc,
		extends:     b,
		includes:    make(map[Block]Handler),
		blocks:      make(map[string]Block),
	}
	return &h
}
