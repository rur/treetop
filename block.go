package treetop

import "fmt"

type blockInternal struct {
	name           string
	defaultPartial Partial
	container      Partial
	execute        TemplateExec
}

func (b *blockInternal) String() string {
	return fmt.Sprintf("<Block Name: '%s'>", b.name)
}

func (b *blockInternal) Name() string {
	return b.name
}

func (b *blockInternal) SetDefault(h Partial) Block {
	b.defaultPartial = h
	return b
}

func (b *blockInternal) Default() Partial {
	return b.defaultPartial
}

func (b *blockInternal) Container() Partial {
	return b.container
}

func (b *blockInternal) Partial(template string, handlerFunc HandlerFunc) Partial {
	h := partialInternal{
		template:    template,
		handlerFunc: handlerFunc,
		extends:     b,
		includes:    make(map[Block]Partial),
		blocks:      make(map[string]Block),
		execute:     b.execute,
	}
	return &h
}

func (b *blockInternal) DefaultPartial(template string, handlerFunc HandlerFunc) Partial {
	b.defaultPartial = b.Partial(template, handlerFunc)
	return b.defaultPartial
}

func (b *blockInternal) Fragment(template string, handlerFunc HandlerFunc) Fragment {
	f := fragmentInternal{
		template:    template,
		handlerFunc: handlerFunc,
		execute:     b.execute,
		extends:     b,
	}
	return &f
}
