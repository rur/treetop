package treetop

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSeeOtherPage(t *testing.T) {
	type args struct {
		w        *httptest.ResponseRecorder
		req      *http.Request
		location string
	}

	req := httptest.NewRequest("GET", "/some/path", nil)
	req.Header.Set("Accept", FragmentContentType)

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "basic",
			args: args{
				w:        httptest.NewRecorder(),
				req:      req,
				location: "test/",
			},
			want: "/some/test/",
		},
		{
			name: "non-treetop request",
			args: args{
				w:        httptest.NewRecorder(),
				req:      httptest.NewRequest("GET", "/some/path", nil),
				location: "test/",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SeeOtherPage(tt.args.w, tt.args.req, tt.args.location)
			got := tt.args.w.Result().Header.Get("X-Treetop-See-Other")
			if got != tt.want {
				t.Errorf("SeeOtherPage() = %v, want %v", got, tt.want)
			}
		})
	}
}
