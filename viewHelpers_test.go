package treetop

import (
	"net/http"
	"testing"
)

func TestIsTreetopRequest(t *testing.T) {
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
			if got := IsTreetopRequest(tt.req); got != tt.want {
				t.Errorf("IsTreetopRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
