package treetop

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func Test_ResponseWrapper_HandleSubView(t *testing.T) {
	type fields struct {
		http.ResponseWriter
		responseID      uint32
		responseWritten bool
		dataCalled      bool
		data            interface{}
		status          int
		partial         *View
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
				partial:        &View{},
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
				partial: &View{
					SubViews: map[string]*View{
						"some-block": &View{
							HandlerFunc: Constant("This is a test"),
						},
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
				partial: &View{
					SubViews: map[string]*View{
						"some-block": &View{
							HandlerFunc: func(rsp Response, _ *http.Request) interface{} {
								rsp.Status(501)
								return "Not Implemented"
							},
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
				partial: &View{
					SubViews: map[string]*View{
						"some-block": &View{
							HandlerFunc: func(rsp Response, _ *http.Request) interface{} {
								return fmt.Sprintf("Response token %v", rsp.ResponseID())
							},
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
				partial: &View{
					SubViews: map[string]*View{
						"some-block": &View{
							HandlerFunc: Constant("This should not happen"),
						},
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
				view:           tt.fields.partial,
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
