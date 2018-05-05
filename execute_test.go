package treetop

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/spf13/afero"
)

func TestTemplateFileSystem(t *testing.T) {
	appFS := afero.NewMemMapFs()
	// create test files and directories
	appFS.Mkdir("templates", 0755)
	afero.WriteFile(appFS, "templates/a.templ.html", []byte(
		`<div>{{ block "hello" .Hello }}<span>Hello, {{ . }}</span>{{ end }}</div>`,
	), 0644)
	afero.WriteFile(appFS, "templates/b.templ.html", []byte(
		`{{ block "hello" . }}<h1>Extended Hello, {{ . }}<h1>{{ end }}`,
	), 0644)

	type args struct {
		templates []string
		data      struct{ Hello string }
	}
	tests := []struct {
		name         string
		fs           http.FileSystem
		args         args
		want         TemplateExec
		errorMessage string
		rendered     string
	}{{
		name: "empty",
		fs:   afero.NewHttpFs(appFS),
		args: args{
			templates: []string{},
			data:      struct{ Hello string }{"World"},
		},
		errorMessage: "No non-empty template paths were yielded for this route",
	}, {
		name: "simple",
		fs:   afero.NewHttpFs(appFS),
		args: args{
			templates: []string{"templates/a.templ.html"},
			data:      struct{ Hello string }{"World"},
		},
		rendered: `<div><span>Hello, World</span></div>`,
	}, {
		name: "extend block",
		fs:   afero.NewHttpFs(appFS),
		args: args{
			templates: []string{"templates/a.templ.html", "templates/b.templ.html"},
			data:      struct{ Hello string }{"World"},
		},
		rendered: `<div><h1>Extended Hello, World<h1></div>`,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execute := TemplateFileSystem(tt.fs)
			render := new(bytes.Buffer)
			if err := execute(render, tt.args.templates, tt.args.data); err != nil {
				if tt.errorMessage != "" {
					if err.Error() != tt.errorMessage {
						t.Errorf("TemplateFileSystem() expected error '%v', wanted '%v'", err, tt.errorMessage)
					}
				} else {
					t.Errorf("TemplateFileSystem() unexpected error '%v'", err)
				}
			} else if render == nil {
				if tt.rendered != "" {
					t.Errorf("Nothing written TemplateFileSystem(), expecting  '%v'", tt.rendered)
				}
			} else if render.String() != tt.rendered {
				t.Errorf("TemplateFileSystem() got '%v', expecting  '%v'", render.String(), tt.rendered)

			}
		})
	}
}
