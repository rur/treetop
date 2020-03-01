package treetop

import (
	"bytes"
	"testing"
)

func TestStringExecutor_constructTemplate(t *testing.T) {
	tests := []struct {
		name    string
		view    *View
		data    interface{}
		want    string
		wantErr string
	}{
		{
			name:    "basic",
			view:    NewView("<p>hello {{ . }}!</p>", Noop),
			data:    "world",
			want:    "<p>hello world!</p>",
			wantErr: "",
		},
		{
			name: "with subviews",
			view: func() *View {
				b := NewView(`<div> base, content: {{ block "content" . }} default here {{ end }} </div>`, Noop)
				b.NewDefaultSubView("content", `<p id="content">hello {{ . }}!</p>`, Noop)
				return b
			}(),
			data:    "world",
			want:    `<div> base, content: <p id="content">hello world!</p> </div>`,
			wantErr: "",
		},
		{
			name: "template parse error",
			view: func() *View {
				b := NewView(`<div> base, content: {{ b.^^lock "content" . }} default here {{ end }} </div>`, Noop)
				b.NewDefaultSubView("content", `<p id="content">hello {{ . }}!</p>`, Noop)
				return b
			}(),
			data:    "world",
			wantErr: `template: :1: function "b" not defined`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exec := StringExecutor{}
			got, err := exec.constructTemplate(tt.view)
			if err != nil {
				if err.Error() != tt.wantErr {
					t.Errorf("StringExecutor.constructTemplate() error = %v, wantErr %v", err, tt.wantErr)
				} else if tt.wantErr == "" {
					t.Errorf("StringExecutor.constructTemplate() unexpected error = %v", err)
				}
				return
			}

			buf := new(bytes.Buffer)
			got.ExecuteTemplate(buf, tt.view.Defines, tt.data)
			gotString := buf.String()
			if gotString != tt.want {
				t.Errorf("StringExecutor.constructTemplate() got %v, want %v", gotString, tt.want)
			}
		})
	}
}
