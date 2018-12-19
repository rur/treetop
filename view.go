package treetop

type viewImpl struct {
	template string
	extends  *blockImpl
	handler  HandlerFunc
	blocks   []*blockImpl
	renderer TemplateExec
}

type blockImpl struct {
	name           string
	parent         *viewImpl
	defaultpartial *viewImpl
}

func (t *viewImpl) SubView(blockName, template string, handler HandlerFunc) View {
	var block *blockImpl
	for i := 0; i < len(t.blocks); i++ {
		if t.blocks[i].name == blockName {
			block = t.blocks[i]
		}
	}
	if block == nil {
		block = &blockImpl{
			parent: t,
			name:   blockName,
		}
		t.blocks = append(t.blocks, block)
	}
	return &viewImpl{
		extends:  block,
		template: template,
		handler:  handler,
		renderer: t.renderer,
	}
}

func (t *viewImpl) DefaultSubView(blockName, template string, handler HandlerFunc) View {
	var block *blockImpl
	for i := 0; i < len(t.blocks); i++ {
		if t.blocks[i].name == blockName {
			block = t.blocks[i]
		}
	}
	if block == nil {
		block = &blockImpl{
			parent: t,
			name:   blockName,
		}
		t.blocks = append(t.blocks, block)
	}
	sub := &viewImpl{
		extends:  block,
		template: template,
		handler:  handler,
		renderer: t.renderer,
	}
	block.defaultpartial = sub
	return sub
}

func (t *viewImpl) Handler() *Handler {
	part := t.derivePartial(nil)
	page := part
	root := t
	for root.extends != nil && root.extends.parent != nil {
		root = root.extends.parent
		page = root.derivePartial(page)
	}
	handler := Handler{
		Fragment: part,
		Page:     page,
		Renderer: t.renderer,
	}

	return &handler
}

func (t *viewImpl) derivePartial(override *Partial) *Partial {
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
