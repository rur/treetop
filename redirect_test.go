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
	req.Header.Set("Accept", TemplateContentType)

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
		ttWant       string
		locationWant string
		status       int
	}{
		{
			name: "basic",
			args: args{
				w:        httptest.NewRecorder(),
				req:      req,
				location: "test/",
				status:   http.StatusSeeOther,
			},
			ttWant:       "/some/test/",
			locationWant: "",
			status:       http.StatusNoContent,
		},
		{
			name: "non-treetop request",
			args: args{
				w:        httptest.NewRecorder(),
				req:      httptest.NewRequest("GET", "/some/path", nil),
				location: "test/",
				status:   http.StatusFound,
			},
			ttWant:       "",
			locationWant: "/some/test/",
			status:       http.StatusFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Redirect(tt.args.w, tt.args.req, tt.args.location, tt.args.status)
			ttGot := tt.args.w.Result().Header.Get("X-Treetop-See-Other")
			if ttGot != tt.ttWant {
				t.Errorf("SeeOtherPage() = %v, want %v", ttGot, tt.ttWant)
			}
			locGot := tt.args.w.Result().Header.Get("Location")
			if locGot != tt.locationWant {
				t.Errorf("http.Redirect() = %v, want %v", locGot, tt.locationWant)
			}
			status := tt.args.w.Result().StatusCode
			if status != tt.status {
				t.Errorf("http.Redirect() = %v, want %v", status, tt.status)
			}
		})
	}
}
