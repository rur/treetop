package treetop

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockRequest(path string, accept string) *http.Request {
	req := httptest.NewRequest("GET", path, nil)
	req.Header.Set("Accept", accept)
	return req
}

func TestWriter(t *testing.T) {
	type args struct {
		w         *httptest.ResponseRecorder
		req       *http.Request
		isPartial bool
	}
	tests := []struct {
		name  string
		args  args
		want  *writer
		want1 bool
	}{
		{
			name: "non tt request",
			args: args{
				w:         httptest.NewRecorder(),
				req:       mockRequest("/Some/path", "text/html"),
				isPartial: true,
			},
			want:  nil,
			want1: false,
		},
		{
			name: "fragment",
			args: args{
				w:         httptest.NewRecorder(),
				req:       mockRequest("/Some/path", PartialContentType+", "+FragmentContentType),
				isPartial: false,
			},
			want: &writer{
				status:      0,
				responseURL: "/Some/path",
				contentType: FragmentContentType,
			},
			want1: true,
		},
		{
			name: "partial",
			args: args{
				w:         httptest.NewRecorder(),
				req:       mockRequest("/Some/path", PartialContentType+", "+FragmentContentType),
				isPartial: true,
			},
			want: &writer{
				status:      0,
				responseURL: "/Some/path",
				contentType: PartialContentType,
			},
			want1: true,
		},
		{
			name: "partial",
			args: args{
				w:         httptest.NewRecorder(),
				req:       mockRequest("/Some/path", PartialContentType+", "+FragmentContentType),
				isPartial: false,
			},
			want: &writer{
				status:      0,
				responseURL: "/Some/path",
				contentType: FragmentContentType,
			},
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				ttW Writer
				ok  bool
			)
			if tt.args.isPartial {
				ttW, ok = NewPartialWriter(tt.args.w, tt.args.req)
			} else {
				ttW, ok = NewFragmentWriter(tt.args.w, tt.args.req)
			}

			if ok {
				fmt.Fprint(ttW, "<p>this is a test</p>")
				status := tt.args.w.Code
				contentType := tt.args.w.HeaderMap.Get("Content-Type")
				responseURL := tt.args.w.HeaderMap.Get("X-Response-Url")

				if tt.want.status > 0 && status != tt.want.status {
					t.Errorf("Writer() status writer = %v, want %v", status, tt.want.status)
				}
				if contentType != tt.want.contentType {
					t.Errorf("Writer() contentType writer = %v, want %v", contentType, tt.want.contentType)
				}
				if responseURL != tt.want.responseURL {
					t.Errorf("Writer() responseURL writer = %v, want %v", responseURL, tt.want.responseURL)
				}
			}

			if ok != tt.want1 {
				t.Errorf("Writer() ok = %v, want %v", ok, tt.want1)
			}
		})
	}
}
