package treetop

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func Test_dataWriter_BlockData(t *testing.T) {
	type fields struct {
		http.ResponseWriter
		responseId      uint32
		responseWritten bool
		dataCalled      bool
		data            interface{}
		status          int
		partial         Partial
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
		flag   bool
		status int
	}{
		{
			name: "Nil case",
			fields: fields{
				ResponseWriter: &httptest.ResponseRecorder{},
				responseId:     1234,
				partial:        Partial{},
			},
			args: args{
				name: "no-such-block",
				req:  req,
			},
			data: nil,
			flag: false,
		},
		{
			name: "Simple data",
			fields: fields{
				ResponseWriter: &httptest.ResponseRecorder{},
				responseId:     1234,
				partial: Partial{
					Blocks: []Partial{
						Partial{
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
			data: "This is a test",
			flag: true,
		},
		{
			name: "Adopt sub-handler HTTP status",
			fields: fields{
				ResponseWriter: &httptest.ResponseRecorder{},
				responseId:     1234,
				status:         400,
				partial: Partial{
					Blocks: []Partial{
						Partial{
							Extends: "some-block",
							HandlerFunc: func(dw DataWriter, _ *http.Request) {
								dw.Status(501)
								dw.Data("Not Implemented")
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
			flag:   true,
			status: 501,
		},
		{
			name: "ResponseId passed down",
			fields: fields{
				ResponseWriter: &httptest.ResponseRecorder{},
				responseId:     1234,
				partial: Partial{
					Blocks: []Partial{
						Partial{
							Extends: "some-block",
							HandlerFunc: func(dw DataWriter, _ *http.Request) {
								dw.Data(fmt.Sprintf("Response token %v", dw.ResponseId()))
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
			flag: true,
		},
		{
			name: "Block not found",
			fields: fields{
				ResponseWriter: &httptest.ResponseRecorder{},
				responseId:     1234,
				partial: Partial{
					Blocks: []Partial{
						Partial{
							Extends:     "some-block",
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
			flag: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dw := &dataWriter{
				writer:          &statusRecorder{ResponseWriter: tt.fields.ResponseWriter},
				responseId:      tt.fields.responseId,
				responseWritten: tt.fields.responseWritten,
				dataCalled:      tt.fields.dataCalled,
				data:            tt.fields.data,
				status:          tt.fields.status,
				partial:         &tt.fields.partial,
			}
			got, got1 := dw.BlockData(tt.args.name, tt.args.req)
			if !reflect.DeepEqual(got, tt.data) {
				t.Errorf("dataWriter.BlockData() got = %v, want %v", got, tt.data)
			}
			if got1 != tt.flag {
				t.Errorf("dataWriter.BlockData() flag = %v, want %v", got1, tt.flag)
			}
			if dw.status != tt.status {
				t.Errorf("dataWriter.status = %v, want %v", dw.status, tt.status)
			}
		})
	}
}
