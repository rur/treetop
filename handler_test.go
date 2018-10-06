package treetop

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestHandler_ServeHTTP(t *testing.T) {
	type fields struct {
		Partial      *Partial
		Postscript   []*Partial
		FragmentOnly bool
		Renderer     TemplateExec
	}
	type args struct {
		resp *httptest.ResponseRecorder
		req  *http.Request
	}
	req := httptest.NewRequest("GET", "/some/path", nil)

	cycle := Partial{
		Extends: "root",
		Content: "test.templ.html",
		HandlerFunc: func(dw DataWriter, req *http.Request) {
			d, _ := dw.BlockData("testblock", req)
			dw.Data(fmt.Sprintf("Loaded sub data: %s", d))
		},
		Blocks: []*Partial{
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
				Partial: &Partial{
					Content:     "test.templ.html",
					HandlerFunc: Constant("somedata"),
				},
				Postscript: []*Partial{},
				Renderer:   TemplExec,
			},
			args: args{
				resp: httptest.NewRecorder(),
				req:  req,
			},
			expect: "Partials: [test.templ.html], Data: somedata",
			status: 200,
		},
		{
			name: "Partial with a block",
			fields: fields{
				Partial: &Partial{
					Content: "test.templ.html",
					HandlerFunc: func(dw DataWriter, req *http.Request) {
						d, _ := dw.BlockData("testblock", req)
						dw.Data(fmt.Sprintf("Loaded sub data: %s", d))
					},
					Blocks: []*Partial{
						{
							Extends:     "testblock",
							Content:     "sub.templ.html",
							HandlerFunc: Constant("my sub data"),
						},
					},
				},
				Postscript: []*Partial{},
				Renderer:   TemplExec,
			},
			args: args{
				resp: httptest.NewRecorder(),
				req:  req,
			},
			expect: "Partials: [test.templ.html sub.templ.html], Data: Loaded sub data: my sub data",
			status: 200,
		},
		{
			name: "Partial with a nested blocks",
			fields: fields{
				Partial: &Partial{
					Content:     "test.templ.html",
					HandlerFunc: blockDebug([]string{"testblock", "testblock2"}),
					Blocks: []*Partial{
						{
							Extends:     "testblock",
							Content:     "sub.templ.html",
							HandlerFunc: blockDebug([]string{"deepblock", "deepblockB"}),
							Blocks: []*Partial{
								{
									Extends:     "deepblock",
									Content:     "sub-sub.templ.html",
									HandlerFunc: Constant("~~sub-subA-data~~"),
								},
								{
									Extends:     "deepblockB",
									Content:     "sub-subB.templ.html",
									HandlerFunc: Constant("~~sub-subB-data~~"),
								},
							},
						},
						{
							Extends:     "testblock2",
							Content:     "sub2.templ.html",
							HandlerFunc: blockDebug([]string{"deepblock2", "deepblock2B"}),
							Blocks: []*Partial{
								{
									Extends:     "deepblock2",
									Content:     "sub2-sub.templ.html",
									HandlerFunc: Constant("~~sub2-subA-data~~"),
								},
								{
									Extends:     "deepblock2B",
									Content:     "sub2-subB.templ.html",
									HandlerFunc: Constant("~~sub2-subB-data~~"),
								},
							},
						},
					},
				},
				Postscript: []*Partial{},
				Renderer:   TemplExec,
			},
			args: args{
				resp: httptest.NewRecorder(),
				req:  req,
			},
			expect: "Partials: [test.templ.html sub.templ.html sub2.templ.html sub-sub.templ.html sub-subB.templ.html sub2-sub.templ.html sub2-subB.templ.html], " +
				"Data: [" +
				"{testblock [{deepblock ~~sub-subA-data~~} {deepblockB ~~sub-subB-data~~}]} " +
				"{testblock2 [{deepblock2 ~~sub2-subA-data~~} {deepblock2B ~~sub2-subB-data~~}]}]",
			status: 200,
		},
		{
			name: "Partial with a cycle",
			fields: fields{
				Partial:    &cycle,
				Postscript: []*Partial{},
				Renderer:   TemplExec,
			},
			args: args{
				resp: httptest.NewRecorder(),
				req:  req,
			},
			expect: "Internal Server Error\n",
			status: 500,
			expectLog: "aggregateTemplates: Encountered naming cycle within nested blocks:\n" +
				"* root -> testblock -> root",
		},
	}
	for _, tt := range tests {
		var output string
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				Partial:      tt.fields.Partial,
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
	fmt.Fprintf(w, "Partials: %s, Data: %v", templates, data)
	return nil
}

func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	f()
	log.SetOutput(os.Stderr)
	return buf.String()
}

func blockDebug(blocknames []string) HandlerFunc {
	return func(dw DataWriter, req *http.Request) {
		var d []struct {
			Block string
			Data  interface{}
		}
		for _, n := range blocknames {
			data, ok := dw.BlockData(n, req)
			if ok {
				d = append(
					d,
					struct {
						Block string
						Data  interface{}
					}{Block: n, Data: data},
				)
			}
		}
		dw.Data(d)
	}
}

func TestPartial_TemplateList(t *testing.T) {
	cycle := Partial{
		Extends:     "prev",
		Content:     "test.templ.html",
		HandlerFunc: Delegate("next"),
		Blocks: []*Partial{
			{
				Extends:     "next",
				Content:     "sub.templ.html",
				HandlerFunc: Constant("my sub data"),
			},
		},
	}
	cycle.Blocks[0].Blocks = append(cycle.Blocks[0].Blocks, &cycle)

	type fields struct {
		Extends     string
		Content     string
		HandlerFunc HandlerFunc
		Parent      *Partial
		Blocks      []*Partial
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "Nested example",
			fields: fields{
				Extends:     "base",
				Content:     "test.templ.html",
				HandlerFunc: blockDebug([]string{"testblock", "testblock2"}),
				Blocks: []*Partial{
					{
						Extends:     "testblock",
						Content:     "sub.templ.html",
						HandlerFunc: blockDebug([]string{"deepblock", "deepblockB"}),
						Blocks: []*Partial{
							{
								Extends:     "deepblock",
								Content:     "sub-sub.templ.html",
								HandlerFunc: Constant("~~sub-subA-data~~"),
							},
							{
								Extends:     "deepblockB",
								Content:     "sub-subB.templ.html",
								HandlerFunc: Constant("~~sub-subB-data~~"),
							},
						},
					},
					{
						Extends:     "testblock2",
						Content:     "sub2.templ.html",
						HandlerFunc: blockDebug([]string{"deepblock2", "deepblock2B"}),
						Blocks: []*Partial{
							{
								Extends:     "deepblock2",
								Content:     "sub2-sub.templ.html",
								HandlerFunc: Constant("~~sub2-subA-data~~"),
							},
							{
								Extends:     "deepblock2B",
								Content:     "sub2-subB.templ.html",
								HandlerFunc: Constant("~~sub2-subB-data~~"),
							},
						},
					},
				},
			},
			want: []string{
				"test.templ.html",
				"sub.templ.html", "sub2.templ.html",
				"sub-sub.templ.html", "sub-subB.templ.html",
				"sub2-sub.templ.html", "sub2-subB.templ.html",
			},
		},
		{
			name: "Nested example",
			fields: fields{
				Extends:     "base",
				Content:     "test.templ.html",
				HandlerFunc: blockDebug([]string{"testblock", "testblock2"}),
				Blocks:      []*Partial{&cycle},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Partial{
				Extends:     tt.fields.Extends,
				Content:     tt.fields.Content,
				HandlerFunc: tt.fields.HandlerFunc,
				Parent:      tt.fields.Parent,
				Blocks:      tt.fields.Blocks,
			}
			got, err := p.TemplateList()
			if (err != nil) != tt.wantErr {
				t.Errorf("Partial.TemplateList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Partial.TemplateList() = %v, want %v", got, tt.want)
			}
		})
	}
}
