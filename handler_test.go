package treetop

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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

	tests := []struct {
		name   string
		fields fields
		args   args
		expect string
		status int
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
			expect: "Templates: [test.templ.html], Data: 'somedata'",
			status: 200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				Template:     tt.fields.Template,
				Postscript:   tt.fields.Postscript,
				FragmentOnly: tt.fields.FragmentOnly,
				Renderer:     tt.fields.Renderer,
			}
			h.ServeHTTP(tt.args.resp, tt.args.req)
		})

		body := tt.args.resp.Body.String()
		if body != tt.expect {
			t.Errorf("Response body = %v, want %v", body, tt.expect)
		}
		if tt.args.resp.Code != tt.status {
			t.Errorf("Response body = %v, want %v", tt.args.resp.Code, tt.status)
		}
	}
}

func TemplExec(w io.Writer, templates []string, data interface{}) error {
	fmt.Fprintf(w, "Templates: %s, Data: %v", templates, data)
	return nil
}
