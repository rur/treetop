package treetop

type partialDefImpl struct {
	template string
	extends  *blockDefImpl
	handler  HandlerFunc
	blocks   []*blockDefImpl
}

func (t *partialDefImpl) Block(name string) BlockDef {
	for i := 0; i < len(t.blocks); i++ {
		if t.blocks[i].name == name {
			return t.blocks[i]
		}
	}
	block := &blockDefImpl{
		parent: t,
		name:   name,
	}
	t.blocks = append(t.blocks, block)
	return block
}

func (t *partialDefImpl) PartialHandler() *Handler {
	// TODO: implement this
	return &Handler{}
}

func (t *partialDefImpl) FragmentHandler() *Handler {
	// TODO: implement this
	return &Handler{
		FragmentOnly: true,
	}
}

type blockDefImpl struct {
	name           string
	parent         *partialDefImpl
	defaultpartial *partialDefImpl
}

func (b *blockDefImpl) Extend(string, HandlerFunc) PartialDef {
	return &partialDefImpl{
		extends: b,
	}
}
func (b *blockDefImpl) Default(string, HandlerFunc) PartialDef {
	return &partialDefImpl{
		extends: b,
	}
}
