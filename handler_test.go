package treetop

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestHandler_ServeHTTP(t *testing.T) {
	type fields struct {
		Template     *Template
		Postscript   []*Template
		FragmentOnly bool
		Renderer     TemplateExec
	}
	type args struct {
		resp *httptest.ResponseRecorder
		req  *http.Request
	}
	req := httptest.NewRequest("GET", "/some/path", nil)

	cycle := Template{
		Content: "test.templ.html",
		HandlerFunc: func(dw DataWriter, req *http.Request) {
			d, _ := dw.BlockData("testblock", req)
			dw.Data(fmt.Sprintf("Loaded sub data: %s", d))
		},
		Blocks: []*Template{
			{
				Extends:     "testblock",
				Content:     "sub.templ.html",
				HandlerFunc: Constant("my sub data"),
			},
		},
	}
	cycle.Blocks[0].Blocks = append(cycle.Blocks[0].Blocks, &cycle)

	tests := []struct {
		name      string
		fields    fields
		args      args
		expect    string
		status    int
		expectLog string
	}{
		{
			name: "Basic",
			fields: fields{
				Template: &Template{
					Content:     "test.templ.html",
					HandlerFunc: Constant("somedata"),
				},
				Postscript: []*Template{},
				Renderer:   TemplExec,
			},
			args: args{
				resp: httptest.NewRecorder(),
				req:  req,
			},
			expect: "Templates: [test.templ.html], Data: \"somedata\"",
			status: 200,
		},
		{
			name: "Template with a block",
			fields: fields{
				Template: &Template{
					Content: "test.templ.html",
					HandlerFunc: func(dw DataWriter, req *http.Request) {
						d, _ := dw.BlockData("testblock", req)
						dw.Data(fmt.Sprintf("Loaded sub data: %s", d))
					},
					Blocks: []*Template{
						{
							Extends:     "testblock",
							Content:     "sub.templ.html",
							HandlerFunc: Constant("my sub data"),
						},
					},
				},
				Postscript: []*Template{},
				Renderer:   TemplExec,
			},
			args: args{
				resp: httptest.NewRecorder(),
				req:  req,
			},
			expect: "Templates: [test.templ.html sub.templ.html], Data: \"Loaded sub data: my sub data\"",
			status: 200,
		},
		{
			name: "Template with a cycle",
			fields: fields{
				Template:   &cycle,
				Postscript: []*Template{},
				Renderer:   TemplExec,
			},
			args: args{
				resp: httptest.NewRecorder(),
				req:  req,
			},
			expect:    "Internal Server Error\n",
			status:    500,
			expectLog: "aggregateTemplate: Max iterations reached, it is likely that there is a cycle in template definitions",
		},
	}
	for _, tt := range tests {
		var output string
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				Template:     tt.fields.Template,
				Postscript:   tt.fields.Postscript,
				FragmentOnly: tt.fields.FragmentOnly,
				Renderer:     tt.fields.Renderer,
			}
			output = captureOutput(func() {
				h.ServeHTTP(tt.args.resp, tt.args.req)
			})
		})

		body := tt.args.resp.Body.String()
		if body != tt.expect {
			t.Errorf("Response body = %v, want %v", body, tt.expect)
		}
		if tt.args.resp.Code != tt.status {
			t.Errorf("Response body = %v, want %v", tt.args.resp.Code, tt.status)
		}
		if len(tt.expectLog) > 0 && !strings.Contains(output, tt.expectLog) {
			t.Errorf("Log output = %v, want %v", output, tt.expectLog)
		}
	}
}

func TemplExec(w io.Writer, templates []string, data interface{}) error {
	fmt.Fprintf(w, "Templates: %s, Data: %#v", templates, data)
	return nil
}

func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	f()
	log.SetOutput(os.Stderr)
	return buf.String()
}
