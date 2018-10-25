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
		Partial    *Partial
		Page       *Partial
		Postscript []Partial
		Renderer   TemplateExec
	}
	type args struct {
		resp *httptest.ResponseRecorder
		req  *http.Request
	}
	req := httptest.NewRequest("GET", "/some/path", nil)
	req.Header.Set("Accept", FragmentContentType)

	cycle := Partial{
		Extends:  "root",
		Template: "test.templ.html",
		HandlerFunc: func(dw DataWriter, req *http.Request) {
			d, _ := dw.BlockData("testblock", req)
			dw.Data(fmt.Sprintf("Loaded sub data: %s", d))
		},
		Blocks: []Partial{
			{
				Extends:     "testblock",
				Template:    "sub.templ.html",
				HandlerFunc: Constant("my sub data"),
			},
		},
	}
	cycle.Blocks[0].Blocks = append(cycle.Blocks[0].Blocks, cycle)

	pagePart := Partial{
		Template:    "base.templ.html",
		HandlerFunc: Constant("base data"),
		Blocks:      []Partial{},
	}
	pagePart.Blocks = append(pagePart.Blocks, Partial{
		Extends:     "test",
		Template:    "test.templ.html",
		HandlerFunc: Constant("partial data"),
	})

	tests := []struct {
		name              string
		fields            fields
		args              args
		expect            string
		status            int
		expectLog         string
		expectContentType string
	}{
		{
			name: "Basic",
			fields: fields{
				Partial: &Partial{
					Template:    "test.templ.html",
					HandlerFunc: Constant("somedata"),
				},
				Postscript: []Partial{},
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
					Template: "test.templ.html",
					HandlerFunc: func(dw DataWriter, req *http.Request) {
						d, _ := dw.BlockData("testblock", req)
						dw.Data(fmt.Sprintf("Loaded sub data: %s", d))
					},
					Blocks: []Partial{
						{
							Extends:     "testblock",
							Template:    "sub.templ.html",
							HandlerFunc: Constant("my sub data"),
						},
					},
				},
				Postscript: []Partial{},
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
					Template:    "test.templ.html",
					HandlerFunc: blockDebug([]string{"testblock", "testblock2"}),
					Blocks: []Partial{
						{
							Extends:     "testblock",
							Template:    "sub.templ.html",
							HandlerFunc: blockDebug([]string{"deepblock", "deepblockB"}),
							Blocks: []Partial{
								{
									Extends:     "deepblock",
									Template:    "sub-sub.templ.html",
									HandlerFunc: Constant("~~sub-subA-data~~"),
								},
								{
									Extends:     "deepblockB",
									Template:    "sub-subB.templ.html",
									HandlerFunc: Constant("~~sub-subB-data~~"),
								},
							},
						},
						{
							Extends:     "testblock2",
							Template:    "sub2.templ.html",
							HandlerFunc: blockDebug([]string{"deepblock2", "deepblock2B"}),
							Blocks: []Partial{
								{
									Extends:     "deepblock2",
									Template:    "sub2-sub.templ.html",
									HandlerFunc: Constant("~~sub2-subA-data~~"),
								},
								{
									Extends:     "deepblock2B",
									Template:    "sub2-subB.templ.html",
									HandlerFunc: Constant("~~sub2-subB-data~~"),
								},
							},
						},
					},
				},
				Postscript: []Partial{},
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
				Postscript: []Partial{},
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
		{
			name: "Full page load from partial endpoint",
			fields: fields{
				Partial:    &pagePart.Blocks[0],
				Page:       &pagePart,
				Postscript: []Partial{},
				Renderer:   TemplExec,
			},
			args: args{
				resp: httptest.NewRecorder(),
				req:  httptest.NewRequest("GET", "/page", nil),
			},
			expect:            "Partials: [base.templ.html test.templ.html], Data: base data",
			status:            200,
			expectContentType: "text/html",
		},
	}
	for _, tt := range tests {
		var output string
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				Partial:    tt.fields.Partial,
				Page:       tt.fields.Page,
				Postscript: tt.fields.Postscript,
				Renderer:   tt.fields.Renderer,
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
		if len(tt.expectContentType) > 0 && tt.args.resp.Header().Get("Content-Type") != tt.expectContentType {
			t.Errorf("ContentType header = %v, want %v", tt.args.resp.Header().Get("Content-Type"), tt.expectContentType)
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
		Template:    "test.templ.html",
		HandlerFunc: Delegate("next"),
		Blocks: []Partial{
			{
				Extends:     "next",
				Template:    "sub.templ.html",
				HandlerFunc: Constant("my sub data"),
			},
		},
	}
	cycle.Blocks[0].Blocks = append(cycle.Blocks[0].Blocks, cycle)

	type fields struct {
		Extends     string
		Template    string
		HandlerFunc HandlerFunc
		Root        *Partial
		Blocks      []Partial
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
				Template:    "test.templ.html",
				HandlerFunc: blockDebug([]string{"testblock", "testblock2"}),
				Blocks: []Partial{
					{
						Extends:     "testblock",
						Template:    "sub.templ.html",
						HandlerFunc: blockDebug([]string{"deepblock", "deepblockB"}),
						Blocks: []Partial{
							{
								Extends:     "deepblock",
								Template:    "sub-sub.templ.html",
								HandlerFunc: Constant("~~sub-subA-data~~"),
							},
							{
								Extends:     "deepblockB",
								Template:    "sub-subB.templ.html",
								HandlerFunc: Constant("~~sub-subB-data~~"),
							},
						},
					},
					{
						Extends:     "testblock2",
						Template:    "sub2.templ.html",
						HandlerFunc: blockDebug([]string{"deepblock2", "deepblock2B"}),
						Blocks: []Partial{
							{
								Extends:     "deepblock2",
								Template:    "sub2-sub.templ.html",
								HandlerFunc: Constant("~~sub2-subA-data~~"),
							},
							{
								Extends:     "deepblock2B",
								Template:    "sub2-subB.templ.html",
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
			name: "Partial with a cycle",
			fields: fields{
				Extends:     "base",
				Template:    "test.templ.html",
				HandlerFunc: blockDebug([]string{"testblock", "testblock2"}),
				Blocks:      []Partial{cycle},
			},
			wantErr: true,
		},
		{
			name: "Partial with cycles",
			fields: fields{
				Extends:     "base",
				Template:    "test.templ.html",
				HandlerFunc: blockDebug([]string{"testblock", "testblock2"}),
				Blocks: []Partial{
					{
						Extends:     "testblock",
						HandlerFunc: Noop,
					},
				},
			},
			want: []string{"test.templ.html"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Partial{
				Extends:     tt.fields.Extends,
				Template:    tt.fields.Template,
				HandlerFunc: tt.fields.HandlerFunc,
				Blocks:      tt.fields.Blocks,
			}
			got, err := p.TemplateList()
			if (err != nil) != tt.wantErr {
				t.Errorf("Partial.TemplateList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Partial.TemplateList() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_insertPartial(t *testing.T) {
	type fields struct {
		Extends     string
		Template    string
		HandlerFunc HandlerFunc
		Blocks      []Partial
	}
	type args struct {
		part *Partial
		seen []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Partial
	}{
		{
			name: "basic",
			fields: fields{
				Extends:  "base",
				Template: "base.templ.html",
				Blocks: []Partial{
					Partial{
						Extends: "test",
					},
				},
			},
			args: args{
				part: &Partial{
					Extends:  "test",
					Template: "test-extended.templ.html",
				},
			},
			want: &Partial{
				Extends:  "base",
				Template: "base.templ.html",
				Blocks: []Partial{
					Partial{
						Extends:  "test",
						Template: "test-extended.templ.html",
					},
				},
			},
		},
		{
			name: "no match",
			fields: fields{
				Extends:  "base",
				Template: "base.templ.html",
				Blocks: []Partial{
					Partial{
						Extends: "test",
					},
				},
			},
			args: args{
				part: &Partial{
					Extends:  "matches-nothing",
					Template: "test-extended.templ.html",
				},
			},
			want: nil,
		},
		{
			name: "keep sublings",
			fields: fields{
				Extends:  "base",
				Template: "base.templ.html",
				Blocks: []Partial{
					Partial{
						Extends: "test",
					},
					Partial{
						Extends:  "test2",
						Template: "test2.templ.html",
					},
				},
			},
			args: args{
				part: &Partial{
					Extends:  "test",
					Template: "test-extended.templ.html",
				},
			},
			want: &Partial{
				Extends:  "base",
				Template: "base.templ.html",
				Blocks: []Partial{
					Partial{
						Extends:  "test",
						Template: "test-extended.templ.html",
					},
					Partial{
						Extends:  "test2",
						Template: "test2.templ.html",
					},
				},
			},
		},
		{
			name: "keep depth",
			fields: fields{
				Extends:  "base",
				Template: "base.templ.html",
				Blocks: []Partial{
					Partial{
						Extends: "test",
					},
					Partial{
						Extends:  "test2",
						Template: "test2.templ.html",
					},
				},
			},
			args: args{
				part: &Partial{
					Extends:  "test",
					Template: "test-extended.templ.html",
					Blocks: []Partial{
						Partial{
							Extends:  "test-sub",
							Template: "test-sub.templ.html",
						},
					},
				},
			},
			want: &Partial{
				Extends:  "base",
				Template: "base.templ.html",
				Blocks: []Partial{
					Partial{
						Extends:  "test",
						Template: "test-extended.templ.html",
						Blocks: []Partial{
							Partial{
								Extends:  "test-sub",
								Template: "test-sub.templ.html",
							},
						},
					},
					Partial{
						Extends:  "test2",
						Template: "test2.templ.html",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Partial{
				Extends:     tt.fields.Extends,
				Template:    tt.fields.Template,
				HandlerFunc: tt.fields.HandlerFunc,
				Blocks:      tt.fields.Blocks,
			}
			if got := insertPartial(p, tt.args.part, tt.args.seen...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("insertPartial(a,b) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_Include(t *testing.T) {
	type fields struct {
		Partial    *Partial
		Page       *Partial
		Postscript []Partial
		Renderer   TemplateExec
	}
	type args struct {
		defs []View
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantPage    []string
		wantPartial []string
		wantError   bool
	}{
		{
			name: "basic",
			fields: fields{
				Partial: &Partial{
					Extends:  "test",
					Template: "test.templ.html",
					Blocks: []Partial{
						Partial{
							Extends: "sub",
						},
					},
				},
				Page: &Partial{
					Extends:  "base",
					Template: "base.templ.html",
					Blocks: []Partial{
						Partial{
							Extends:  "test",
							Template: "test.templ.html",
							Blocks: []Partial{
								Partial{
									Extends: "sub",
								},
							},
						},
					},
				},
			},
			args: args{
				defs: []View{
					&viewImpl{
						extends:  &blockImpl{name: "sub"},
						template: "sub-impl.templ.html",
					},
				},
			},
			wantPage: []string{
				"base.templ.html",
				"test.templ.html",
				"sub-impl.templ.html",
			},
			wantPartial: []string{
				"test.templ.html",
				"sub-impl.templ.html",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				Partial:    tt.fields.Partial,
				Page:       tt.fields.Page,
				Postscript: tt.fields.Postscript,
				Renderer:   tt.fields.Renderer,
			}
			gotHandler := h.Include(tt.args.defs...)

			gotPage, err := gotHandler.Page.TemplateList()
			if err != nil && !tt.wantError {
				t.Errorf("Handler.Include() Page = got unexpected error %s", err.Error())
			} else if !reflect.DeepEqual(gotPage, tt.wantPage) {
				t.Errorf("Handler.Include() Page = %#v, want %#v", gotPage, tt.wantPage)
			}
			gotPartial, err := gotHandler.Partial.TemplateList()
			if err != nil && !tt.wantError {
				t.Errorf("Handler.Include() Partial = got unexpected error %s", err.Error())
			} else if !reflect.DeepEqual(gotPartial, tt.wantPartial) {
				t.Errorf("Handler.Include() Partial = %#v, want %#v", gotPartial, tt.wantPartial)
			}
		})
	}
}
