package treetop

import "testing"

func Test_fragmentInternal_String(t *testing.T) {
	type fields struct {
		template    string
		handlerFunc HandlerFunc
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
			},
			want: "<Fragment Template: 'test.templ.html'>",
		},
		{
			name: "block with default handler",
			fields: fields{
				template:    "test.templ.html",
				handlerFunc: Noop,
			},
			want: "<Fragment Template: 'test.templ.html'>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &fragmentInternal{
				template:    tt.fields.template,
				handlerFunc: tt.fields.handlerFunc,
			}
			if got := h.String(); got != tt.want {
				t.Errorf("fragmentInternal.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
