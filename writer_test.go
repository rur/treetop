package treetop

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func Test_dataWriter_BlockData(t *testing.T) {
	type fields struct {
		writer          http.ResponseWriter
		responseToken   string
		responseWritten bool
		dataCalled      bool
		data            interface{}
		status          int
		template        *Template
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
		want   interface{}
		want1  bool
	}{
		{
			name: "Nil case",
			fields: fields{
				writer:        &httptest.ResponseRecorder{},
				responseToken: "test-response",
				template:      &Template{},
			},
			args: args{
				name: "no-such-block",
				req:  req,
			},
			want:  nil,
			want1: false,
		},
		{
			name: "Simple data",
			fields: fields{
				writer:        &httptest.ResponseRecorder{},
				responseToken: "test-response",
				template: &Template{
					Blocks: []*Template{
						&Template{
							Extends:     "some-block",
							HandlerFunc: Constant("This is a test"),
						},
					},
				},
			},
			args: args{
				name: "some-block",
				req:  req,
			},
			want:  "This is a test",
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dw := &dataWriter{
				writer:          tt.fields.writer,
				responseToken:   tt.fields.responseToken,
				responseWritten: tt.fields.responseWritten,
				dataCalled:      tt.fields.dataCalled,
				data:            tt.fields.data,
				status:          tt.fields.status,
				template:        tt.fields.template,
			}
			got, got1 := dw.BlockData(tt.args.name, tt.args.req)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dataWriter.BlockData() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("dataWriter.BlockData() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
