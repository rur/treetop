package treetop

import (
	"net/http"
	"reflect"
	"testing"
)

type TestBase struct {
	Title   string
	Content interface{}
	Footer  interface{}
}

func handler(w DataWriter, r *http.Request) {
	content, ok := w.Delegate("content", r)
	if !ok {
		content = 999
	}
	footer, ok := w.Delegate("footer", r)
	if !ok {
		footer = 999
	}
	w.Data(TestBase{
		Title:   "Base!!",
		Content: content,
		Footer:  footer,
	})
}

func contentHandler(w DataWriter, r *http.Request) {
	w.Data(123)
}

func footerHandler(w DataWriter, r *http.Request) {
	w.Data("This is the footer")
}

func Test_resolveTemplatesForPartial(t *testing.T) {
	base := NewPage("base.templ.html", handler)
	content := base.DefineBlock("content")
	footer := base.DefineBlock("footer").WithDefault("footer.templ.html", footerHandler)
	sub := content.Extend("sub.templ.html", contentHandler)
	subContent := sub.DefineBlock("subContent")
	subSub := subContent.Extend("sub_sub.templ.html", Noop)
	otherFooter := footer.Extend("other_footer.templ.html", footerHandler)
	subWithOther := sub.Includes(otherFooter)
	subWithOther2 := subWithOther.Includes(
		NewPage("orphan.templ.html", handler),
	)

	type args struct {
		block   Block
		primary Partial
	}
	tests := []struct {
		name       string
		args       args
		handlerMap map[Block]Partial
		templates  []string
	}{
		{
			name: "basic",
			args: args{
				block:   base.Extends(),
				primary: base,
			},
			handlerMap: map[Block]Partial{
				base.Extends(): base,
				footer:         footer.Default(),
			},
			templates: []string{
				"base.templ.html",
				"footer.templ.html",
			},
		},
		{
			name: "sub",
			args: args{
				block:   base.Extends(),
				primary: sub,
			},
			handlerMap: map[Block]Partial{
				base.Extends(): base,
				footer:         footer.Default(),
				content:        sub,
			},
			templates: []string{
				"base.templ.html",
				"sub.templ.html",
				"footer.templ.html",
			},
		},
		{
			name: "includes",
			args: args{
				block:   base.Extends(),
				primary: subWithOther,
			},
			handlerMap: map[Block]Partial{
				base.Extends(): base,
				footer:         otherFooter,
				content:        subWithOther,
			},
			templates: []string{
				"base.templ.html",
				"sub.templ.html",
				"other_footer.templ.html",
			},
		},
		{
			name: "ignore orphan",
			args: args{
				block:   base.Extends(),
				primary: subWithOther2,
			},
			handlerMap: map[Block]Partial{
				base.Extends(): base,
				footer:         otherFooter,
				content:        subWithOther2,
			},
			templates: []string{
				"base.templ.html",
				"sub.templ.html",
				"other_footer.templ.html",
			},
		},
		{
			name: "subSub",
			args: args{
				block:   base.Extends(),
				primary: subSub,
			},
			handlerMap: map[Block]Partial{
				base.Extends(): base,
				footer:         footer.Default(),
				content:        sub,
				subContent:     subSub,
			},
			templates: []string{
				"base.templ.html",
				"sub.templ.html",
				"sub_sub.templ.html",
				"footer.templ.html",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blockMap := resolveBlockMap(tt.args.block, tt.args.primary)
			if !reflect.DeepEqual(blockMap, tt.handlerMap) {
				t.Errorf("resolveTemplatesForPartial() got = %v, want %v", blockMap, tt.handlerMap)
			}

			handler, ok := blockMap[tt.args.block]
			if !ok {
				t.Errorf("Handler was not found for root block %s", tt.args.block)
			}

			templates := resolvePartialTemplates(handler, blockMap)
			if !reflect.DeepEqual(templates, tt.templates) {
				t.Errorf("resolveTemplatesForPartial() got1 = %v, want %v", templates, tt.templates)
			}
		})
	}
}

func Test_partialInternal_String(t *testing.T) {
	type fields struct {
		template    string
		handlerFunc HandlerFunc
		blocks      map[string]Block
		includes    map[Block]Partial
		extends     Block
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "basic",
			fields: fields{
				template:    "test.templ.html",
				handlerFunc: Noop,
				blocks:      map[string]Block{},
				includes:    map[Block]Partial{},
				extends:     &blockInternal{name: "test"},
			},
			want: "<Partial Template: 'test.templ.html' Extends: <Block Name: 'test'>>",
		},
		{
			name: "block with default handler",
			fields: fields{
				template:    "test.templ.html",
				handlerFunc: Noop,
				blocks:      map[string]Block{},
				includes:    map[Block]Partial{},
				extends: &blockInternal{
					name: "test",
					defaultPartial: &partialInternal{
						template:    "blockdefault.templ.html",
						handlerFunc: Noop,
						blocks:      map[string]Block{},
						includes:    map[Block]Partial{},
						extends:     nil,
					},
				},
			},
			want: "<Partial Template: 'test.templ.html' Extends: <Block Name: 'test'>>",
		},
		{
			name: "with includes",
			fields: fields{
				template:    "test.templ.html",
				handlerFunc: Noop,
				blocks:      map[string]Block{},
				includes: map[Block]Partial{
					&blockInternal{name: "a"}: &partialInternal{template: "include_a.templ.html"},
					&blockInternal{name: "b"}: &partialInternal{template: "include_b.templ.html"},
					&blockInternal{name: "c"}: &partialInternal{template: "include_c.templ.html"},
				},
				extends: &blockInternal{
					name: "test",
				},
			},
			want: "<Partial Template: 'test.templ.html' Extends: <Block Name: 'test'> Includes: x3>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &partialInternal{
				template:    tt.fields.template,
				handlerFunc: tt.fields.handlerFunc,
				blocks:      tt.fields.blocks,
				includes:    tt.fields.includes,
				extends:     tt.fields.extends,
			}
			if got := h.String(); got != tt.want {
				t.Errorf("partialInternal.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
