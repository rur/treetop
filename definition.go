package treetop

type partialDefImpl struct {
	template string
	extends  *blockDefImpl
	handler  HandlerFunc
	blocks   []*blockDefImpl
	renderer TemplateExec
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

func (t *partialDefImpl) PageHandler() *Handler {
	part := t.derivePartial(nil)
	page := part
	root := t
	for root.extends != nil && root.extends.parent != nil {
		root = root.extends.parent
		page = root.derivePartial(page)
	}
	handler := Handler{
		Page:     page,
		Renderer: t.renderer,
	}

	return &handler
}

func (t *partialDefImpl) PartialHandler() *Handler {
	part := t.derivePartial(nil)
	page := part
	root := t
	for root.extends != nil && root.extends.parent != nil {
		root = root.extends.parent
		page = root.derivePartial(page)
	}
	handler := Handler{
		Partial:  part,
		Page:     page,
		Renderer: t.renderer,
	}

	return &handler
}

func (t *partialDefImpl) FragmentHandler() *Handler {
	return &Handler{
		Partial:  t.derivePartial(nil),
		Renderer: t.renderer,
	}
}

func (t *partialDefImpl) derivePartial(override *Partial) *Partial {
	var extends string
	if t.extends != nil {
		extends = t.extends.name
	}

	p := Partial{
		Extends:     extends,
		Template:    t.template,
		HandlerFunc: t.handler,
	}

	var blP *Partial
	for i := 0; i < len(t.blocks); i++ {
		b := t.blocks[i]
		blP = nil
		if override != nil && override.Extends == b.name {
			blP = override
		} else if b.defaultpartial != nil {
			blP = b.defaultpartial.derivePartial(override)
		} else {
			// fallback when there is no default
			blP = &Partial{Extends: b.name, HandlerFunc: Noop}
		}

		p.Blocks = append(p.Blocks, Partial{
			Extends:     b.name,
			Template:    blP.Template,
			HandlerFunc: blP.HandlerFunc,
			Blocks:      blP.Blocks,
		})
	}
	return &p
}

type blockDefImpl struct {
	name           string
	parent         *partialDefImpl
	defaultpartial *partialDefImpl
}

func (b *blockDefImpl) Define(template string, handler HandlerFunc) PartialDef {
	return &partialDefImpl{
		extends:  b,
		template: template,
		handler:  handler,
		renderer: b.parent.renderer,
	}
}
func (b *blockDefImpl) Default(template string, handler HandlerFunc) PartialDef {
	p := &partialDefImpl{
		extends:  b,
		template: template,
		handler:  handler,
		renderer: b.parent.renderer,
	}
	b.defaultpartial = p
	return p
}
