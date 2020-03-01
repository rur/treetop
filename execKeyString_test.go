package treetop

import (
	"bytes"
	"html/template"
	"testing"
)

func TestKeyedStringExecutor_constructTemplate(t *testing.T) {
	tests := []struct {
		name    string
		exec    *KeyedStringExecutor
		view    *View
		data    interface{}
		want    string
		wantErr string
	}{
		{
			name: "basic",
			exec: func() *KeyedStringExecutor {
				exec, err := NewKeyedStringExecutor(map[string]string{
					"test.key": "<p>hello {{ . }}!</p>",
				})
				if err != nil {
					panic(err)
				}
				return exec
			}(),
			view:    NewView("test.key", Noop),
			data:    "world",
			want:    "<p>hello world!</p>",
			wantErr: "",
		},
		{
			name: "with subviews",
			exec: func() *KeyedStringExecutor {
				exec, err := NewKeyedStringExecutor(map[string]string{
					"base.html":    `<div> base, content: {{ block "content" . }} default here {{ end }} </div>`,
					"content.html": `<p id="content">hello {{ . }}!</p>`,
				})
				if err != nil {
					panic(err)
				}
				return exec
			}(),
			view: func() *View {
				b := NewView("base.html", Noop)
				b.NewDefaultSubView("content", "content.html", Noop)
				return b
			}(),
			data:    "world",
			want:    `<div> base, content: <p id="content">hello world!</p> </div>`,
			wantErr: "",
		},
		{
			name: "key not found",
			exec: func() *KeyedStringExecutor {
				exec, err := NewKeyedStringExecutor(map[string]string{
					"base.html":    `<div> base, content: {{ block "content" . }} default here {{ end }} </div>`,
					"content.html": `<p id="content">hello {{ . }}!</p>`,
				})
				if err != nil {
					panic(err)
				}
				return exec
			}(),
			view: func() *View {
				b := NewView("base.html", Noop)
				b.NewDefaultSubView("content", "content-other.html", Noop)
				return b
			}(),
			data:    "world",
			wantErr: "KeyedStringExecutor: no template found for key 'content-other.html'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.exec.constructTemplate(tt.view)
			if err != nil {
				if err.Error() != tt.wantErr {
					t.Errorf("KeyedStringExecutor.constructTemplate() error = %v, wantErr %v", err, tt.wantErr)
				} else if tt.wantErr == "" {
					t.Errorf("KeyedStringExecutor.constructTemplate() unexpected error = %v", err)
				}
				return
			}

			buf := new(bytes.Buffer)
			got.ExecuteTemplate(buf, tt.view.Defines, tt.data)
			gotString := buf.String()
			if gotString != tt.want {
				t.Errorf("KeyedStringExecutor.constructTemplate() got %v, want %v", gotString, tt.want)
			}
		})
	}
}

func TestNewKeyedStringExecutor(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name      string
		templates map[string]string
		key       string
		want      string
		wantErr   string
	}{
		{
			name: "",
			templates: map[string]string{
				"test.html": `<p>{{ . }}</p>`,
			},
			want: `<p>some data here</p>`,
		},
		{
			name: "",
			templates: map[string]string{
				"err.html": `<p>{{ .$TEST }}</p>`,
			},
			wantErr: `template: err.html:1: unexpected bad character U+0024 '$' in command`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewKeyedStringExecutor(tt.templates)
			if err != nil {
				if tt.wantErr == "" {
					t.Errorf("Unexpected error: %s", err)
				} else if tt.wantErr != err.Error() {
					t.Errorf("NewKeyedStringExecutor() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if tt.key == "" {
				return
			}
			buf := new(bytes.Buffer)
			tpl, err := template.New("test").AddParseTree("test", got.parsed[tt.key])
			if err != nil {
				t.Error(err)
			}

			err = tpl.Execute(buf, "some data here")
			if err != nil {
				t.Error(err)
			}

			gotStr := buf.String()
			if gotStr != tt.want {
				t.Errorf("NewKeyedStringExecutor() = %v, want %v", gotStr, tt.want)
			}
		})
	}
}
