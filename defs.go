package treetop

type templateDef struct {
	templateDef string
	extends     *blockDef
	handler     HandlerFunc
	blocks      []*blockDef
}

func (t *templateDef) Block(string) BlockDef {
	// TODO: implement this
	return &blockDef{
		parent: t.extends,
	}
}

func (t *templateDef) PartialHandler() *Handler {
	// TODO: implement this
	return &Handler{}
}

func (t *templateDef) FragmentHandler() *Handler {
	// TODO: implement this
	return &Handler{
		FragmentOnly: true,
	}
}

type blockDef struct {
	name            string
	parent          *blockDef
	defaultTemplate *templateDef
}

func (b *blockDef) Extend(string, HandlerFunc) TemplateDef {
	return &templateDef{
		extends: b,
	}
}
func (b *blockDef) Default(string, HandlerFunc) TemplateDef {
	return &templateDef{
		extends: b,
	}
}
