package treetop

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsTemplateRequest(t *testing.T) {
	tests := []struct {
		name string
		req  *http.Request
		want bool
	}{
		{
			name: "basic any content type",
			req:  mockRequest("/Some/path", "*/*"),
			want: false,
		},
		{
			name: "single specific contnet type",
			req:  mockRequest("/Some/path", TemplateContentType),
			want: true,
		},
		{
			name: "basic any content type",
			req:  mockRequest("/Some/path", "text/html; "+TemplateContentType),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTemplateRequest(tt.req); got != tt.want {
				t.Errorf("IsTemplateRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedirect(t *testing.T) {
	type args struct {
		w        *httptest.ResponseRecorder
		req      *http.Request
		location string
		status   int
	}
	req := httptest.NewRequest("GET", "/some/path", nil)
	req.Header.Set("Accept", TemplateContentType)
	tests := []struct {
		name         string
		args         args
		wantHeader   bool
		locationWant string
		status       int
	}{
		{
			name: "basic",
			args: args{
				w:        httptest.NewRecorder(),
				req:      mockRequest("/some/path", TemplateContentType),
				location: "test/",
				status:   http.StatusSeeOther,
			},
			wantHeader:   true,
			locationWant: "/some/test/",
			status:       http.StatusOK,
		},
		{
			name: "non-treetop request",
			args: args{
				w:        httptest.NewRecorder(),
				req:      mockRequest("/some/path", "*/*"),
				location: "test/",
				status:   http.StatusFound,
			},
			wantHeader:   false,
			locationWant: "/some/test/",
			status:       http.StatusFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Redirect(tt.args.w, tt.args.req, tt.args.location, tt.args.status)
			ttGot := tt.args.w.Result().Header.Get("X-Treetop-Redirect")
			if tt.wantHeader && ttGot == "" {
				t.Errorf("Redirect() expecting X-Treetop-Redirect header to be set")
			}
			locGot := tt.args.w.Result().Header.Get("Location")
			if locGot != tt.locationWant {
				t.Errorf("Redirect() = %v, want %v", locGot, tt.locationWant)
			}
			status := tt.args.w.Result().StatusCode
			if status != tt.status {
				t.Errorf("Redirect() = %v, want %v", status, tt.status)
			}
		})
	}
}
