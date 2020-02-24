package treetop

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestBeginResponse(t *testing.T) {
	rsp := BeginResponse(context.Background(), httptest.NewRecorder())
	if rsp.ResponseID() == 0 {
		t.Error("Expecting a non-zero treetop response ID to be assigned")
	}
	rsp.Cancel()
	select {
	case <-rsp.Context().Done():
	case <-time.After(1 * time.Millisecond):
		t.Error("Expecting cancel to resolve the treetop context")
	}
}

func TestResponse_WithView(t *testing.T) {
	rsp := BeginResponse(context.Background(), httptest.NewRecorder())

	view := NewView("view.html", Noop)
	view.NewDefaultSubView("test", "test.html", Constant("test!!"))

	rsp = rsp.WithView(view)
	d := rsp.HandleSubView("test", nil)
	if d != "test!!" {
		t.Errorf("Expecting subview data to be 'test!!' got %v", d)
	}
}

func Test_ResponseWrapper_HandleSubView(t *testing.T) {
	type fields struct {
		http.ResponseWriter
		responseID      uint32
		responseWritten bool
		dataCalled      bool
		data            interface{}
		status          int
		subViews        map[string]*View
	}
	req := httptest.NewRequest("GET", "/some/path", nil)
	type args struct {
		name string
		req  *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		data   interface{}
		status int
	}{
		{
			name: "Nil case",
			fields: fields{
				ResponseWriter: &httptest.ResponseRecorder{},
				responseID:     1234,
			},
			args: args{
				name: "no-such-block",
				req:  req,
			},
			data: nil,
		},
		{
			name: "Simple data",
			fields: fields{
				ResponseWriter: &httptest.ResponseRecorder{},
				responseID:     1234,
				subViews: map[string]*View{
					"some-block": &View{
						HandlerFunc: Constant("This is a test"),
					},
				},
			},
			args: args{
				name: "some-block",
				req:  req,
			},
			data: "This is a test",
		},
		{
			name: "Adopt sub-handler HTTP status",
			fields: fields{
				ResponseWriter: &httptest.ResponseRecorder{},
				responseID:     1234,
				status:         400,
				subViews: map[string]*View{
					"some-block": &View{
						HandlerFunc: func(rsp Response, _ *http.Request) interface{} {
							rsp.Status(501)
							return "Not Implemented"
						},
					},
				},
			},
			args: args{
				name: "some-block",
				req:  req,
			},
			data:   "Not Implemented",
			status: 501,
		},
		{
			name: "ResponseID passed down",
			fields: fields{
				ResponseWriter: &httptest.ResponseRecorder{},
				responseID:     1234,
				subViews: map[string]*View{
					"some-block": &View{
						HandlerFunc: func(rsp Response, _ *http.Request) interface{} {
							return fmt.Sprintf("Response token %v", rsp.ResponseID())
						},
					},
				},
			},
			args: args{
				name: "some-block",
				req:  req,
			},
			data: "Response token 1234",
		},
		{
			name: "Block not found",
			fields: fields{
				ResponseWriter: &httptest.ResponseRecorder{},
				responseID:     1234,
				subViews: map[string]*View{
					"some-block": &View{
						HandlerFunc: Constant("This should not happen"),
					},
				},
			},
			args: args{
				name: "some-other-block",
				req:  req,
			},
			data: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp := &ResponseWrapper{
				ResponseWriter: tt.fields.ResponseWriter,
				responseID:     tt.fields.responseID,
				finished:       tt.fields.responseWritten,
				status:         tt.fields.status,
				subViews:       tt.fields.subViews,
			}
			got := rsp.HandleSubView(tt.args.name, tt.args.req)
			if !reflect.DeepEqual(got, tt.data) {
				t.Errorf("ResponseWrapper.HandleSubView() got = %v, want %v", got, tt.data)
			}
			if rsp.status != tt.status {
				t.Errorf("ResponseWrapper.status = %v, want %v", rsp.status, tt.status)
			}
		})
	}
}

// Apply updates to the ResponseWrapper instance which should
// affect the headers added by the template writer.
func TestResponseWrapper_NewTemplateWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	rsp := BeginResponse(context.Background(), rec)
	rsp.Status(http.StatusTeapot)
	rsp.ReplacePageURL("/some/url")
	req := mockRequest("/some/url", TemplateContentType)
	w, ok := rsp.NewTemplateWriter(req)
	if !ok {
		t.Error("Expecting a template writer to have been created, but it was not okay :(")
		return
	}
	fmt.Fprint(w, `<p id="test">hello world!</p>`)

	buf := new(bytes.Buffer)
	buf.ReadFrom(rec.Body)
	got := buf.String()
	expect := `<p id="test">hello world!</p>`
	if expect != got {
		t.Errorf("Expecting body to be\n%s\ngot\n%s", expect, got)
	}

	if pageURL := rec.Header().Get("X-Page-URL"); pageURL != "/some/url" {
		t.Errorf("Expecting X-Page-URL header of %s got %s", pageURL, "/some/url")
	}

	if val := rec.Header().Get("X-Response-History"); val != "replace" {
		t.Errorf("Expecting X-Response-History header of %s got %s", val, "replace")
	}

	if rec.Code != http.StatusTeapot {
		t.Errorf("Expecting response status code of %d, got %d", http.StatusTeapot, rec.Code)
	}
}

// ResponseWrapper should not create a template writer for a non-template request
func TestResponseWrapper_NewTemplateWriter_NonTemplateRequest(t *testing.T) {
	rec := httptest.NewRecorder()
	rsp := BeginResponse(context.Background(), rec)
	rsp.Status(http.StatusTeapot)
	rsp.ReplacePageURL("/some/url")
	req := mockRequest("/some/url", "*/*")
	_, ok := rsp.NewTemplateWriter(req)
	if ok {
		t.Error("Expecting a template writer not to have been created, but it was :(")
		return
	}
}

// ResponseWrapper should not create a template writer if the underlying response has
// already been written
func TestResponseWrapper_NewTemplateWriter_WritingHijacked(t *testing.T) {
	rec := httptest.NewRecorder()
	rsp := BeginResponse(context.Background(), rec)

	// emulate handler writing directly to the response object
	fmt.Fprint(rsp, `<p id="test">hello world!</p>`)

	if !rsp.Finished() {
		t.Error("Expecting the response wrapper to have flagged header write. Finished() return true")
		return
	}

	req := mockRequest("/some/url", TemplateContentType)
	_, ok := rsp.NewTemplateWriter(req)
	if ok {
		t.Error("Expecting a template writer not to have been created, but it was :(")
		return
	}
}
