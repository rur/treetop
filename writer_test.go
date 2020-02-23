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
		w            *httptest.ResponseRecorder
		req          *http.Request
		status       int
		isPartial    bool
		pageURL      string
		replaceState bool
	}
	tests := []struct {
		name            string
		args            args
		wantWriter      bool
		wantContentType string
		wantStatus      int
		wantBody        string
		wantPageURL     string
	}{
		{
			name: "non tt request",
			args: args{
				w:         httptest.NewRecorder(),
				req:       mockRequest("/Some/path", "text/html"),
				isPartial: true,
			},
			wantWriter: false,
		},
		{
			name: "fragment",
			args: args{
				w:         httptest.NewRecorder(),
				req:       mockRequest("/Some/path", TemplateContentType),
				status:    201,
				isPartial: false,
			},
			wantWriter: true,

			wantContentType: TemplateContentType,
			wantStatus:      201,
			wantBody:        `<p>this is a test</p>`,
		},
		{
			name: "partial",
			args: args{
				w:         httptest.NewRecorder(),
				req:       mockRequest("/Some/path", TemplateContentType),
				status:    201,
				isPartial: true,
			},
			wantWriter: true,

			wantContentType: TemplateContentType,
			wantStatus:      201,
			wantBody:        `<p>this is a test</p>`,
			wantPageURL:     "/Some/path",
		},
		{
			name: "partial",
			args: args{
				w:            httptest.NewRecorder(),
				req:          mockRequest("/Some/path", TemplateContentType),
				status:       201,
				isPartial:    true,
				pageURL:      "/Some/other/path",
				replaceState: true,
			},

			wantWriter:      true,
			wantContentType: TemplateContentType,
			wantStatus:      201,
			wantBody:        `<p>this is a test</p>`,
			wantPageURL:     "/Some/other/path",
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

			if ok != tt.wantWriter {
				t.Errorf("Writer() ok = %v, want %v", ok, tt.wantWriter)
			}

			if !ok {
				return
			}
			if tt.args.status > 0 {
				ttW.Status(tt.args.status)
			}
			if tt.args.pageURL != "" {
				if tt.args.replaceState {
					ttW.ReplacePageURL(tt.args.pageURL)
				} else {
					ttW.DesignatePageURL(tt.args.pageURL)
				}
			}
			fmt.Fprint(ttW, "<p>this is a test</p>")
			status := tt.args.w.Code
			contentType := tt.args.w.HeaderMap.Get("Content-Type")
			pageURL := tt.args.w.HeaderMap.Get("X-Page-Url")
			responseHistory := tt.args.w.HeaderMap.Get("X-Response-History")
			var wantHistory string
			if tt.args.replaceState {
				wantHistory = "replace"
			}

			if tt.wantStatus > 0 && status != tt.wantStatus {
				t.Errorf("Writer() status writer = %v, want %v", status, tt.wantStatus)
			}
			if contentType != tt.wantContentType {
				t.Errorf("Writer() contentType writer = %v, want %v", contentType, tt.wantContentType)
			}
			if pageURL != tt.wantPageURL {
				t.Errorf("Writer() page url header = %v, want %v", pageURL, tt.wantPageURL)
			}
			if responseHistory != wantHistory {
				t.Errorf("Writer() response history header = %v, want %v", responseHistory, wantHistory)
			}
		})
	}
}
