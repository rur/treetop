package treetop

type blockInternal struct {
	name           string
	defaultHandler Handler
	container      Handler
}

func (b *blockInternal) WithDefault(template string, handlerFunc HandlerFunc) Block {
	b.defaultHandler = b.Extend(template, handlerFunc)
	return b
}

func (b *blockInternal) Default() Handler {
	return b.defaultHandler
}

func (b *blockInternal) Container() Handler {
	return b.container
}

func (b *blockInternal) Extend(template string, handlerFunc HandlerFunc) Handler {
	h := handlerInternal{
		template: template,
		handlerFunc:     handlerFunc,
		extends:  b,
		includes: make(map[Block]Handler),
		blocks: make(map[string]Block),
	}
	return &h
}
