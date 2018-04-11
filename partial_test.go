package treetop

import "testing"

func Test_partialInternal_String(t *testing.T) {
	type fields struct {
		template    string
		handlerFunc HandlerFunc
		blocks      map[string]Block
		includes    map[Block]Handler
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
				includes:    map[Block]Handler{},
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
				includes:    map[Block]Handler{},
				extends: &blockInternal{
					name: "test",
					defaultHandler: &partialInternal{
						template:    "blockdefault.templ.html",
						handlerFunc: Noop,
						blocks:      map[string]Block{},
						includes:    map[Block]Handler{},
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
				includes: map[Block]Handler{
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
