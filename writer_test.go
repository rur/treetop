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
		wantError       string
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
			name: "partial replace history state",
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
		{
			name: "Ignore invalid URL",
			args: args{
				w:       httptest.NewRecorder(),
				req:     mockRequest("/Some/path", TemplateContentType),
				status:  201,
				pageURL: "/$%^&*",
			},

			wantWriter: true,

			wantContentType: TemplateContentType,
			wantStatus:      201,
			wantBody:        `<p>this is a test</p>`,
			wantPageURL:     "/$%^&*",
		},
		{
			name: "Escape non-ascii urls",
			args: args{
				w:       httptest.NewRecorder(),
				req:     mockRequest("/Some/path", TemplateContentType),
				status:  201,
				pageURL: "/☆",
			},

			wantWriter: true,

			wantContentType: TemplateContentType,
			wantStatus:      201,
			wantBody:        `<p>this is a test</p>`,
			wantPageURL:     "/%E2%98%86",
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
			_, err := fmt.Fprint(ttW, "<p>this is a test</p>")
			if err != nil {
				t.Errorf("Writer got error %s", err)
			}
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

func TestNewPartialWriter_Basic(t *testing.T) {
	request := mockRequest("/test", TemplateContentType)
	rsp := httptest.NewRecorder()
	if tW, ok := NewPartialWriter(rsp, request); ok {
		tW.Write([]byte("something here!"))
	} else {
		t.Error("Failed to recognize template request")
		return
	}

	if rsp.Code != 200 {
		t.Errorf("Expecting status 200 got %d", rsp.Code)
	}

	cType := rsp.HeaderMap.Get("Content-Type")
	if cType != TemplateContentType {
		t.Errorf("Expecting content type %s, got %s", TemplateContentType, cType)
	}

	pageURL := rsp.HeaderMap.Get("X-Page-URL")
	if pageURL != "/test" {
		t.Errorf("Expecting page URL %s, got %s", "/test", pageURL)
	}

	body := rsp.Body.String()
	if body != "something here!" {
		t.Errorf("Expecting body 'something here!', got '%s'", body)
	}
}

// Test that a treetop writer implements http.ResponseWriter interface and behavior
func TestNewPartialWriter_AsResponseWriter(t *testing.T) {
	request := mockRequest("/test", TemplateContentType)
	rsp := httptest.NewRecorder()
	if tW, ok := NewPartialWriter(rsp, request); ok {
		rw, ok := tW.(http.ResponseWriter)
		if !ok {
			t.Error("Failed to coerce partial writer to http.ResponseWriter type")
			return
		}
		rw.WriteHeader(http.StatusTeapot)
		rw.Write([]byte("I'm a teapot!"))
	} else {
		t.Error("Failed to recognize template request")
		return
	}

	if rsp.Code != 418 {
		t.Errorf("Expecting status 418 got %d", rsp.Code)
	}

	cType := rsp.HeaderMap.Get("Content-Type")
	if cType != TemplateContentType {
		t.Errorf("Expecting content type %s, got %s", TemplateContentType, cType)
	}

	pageURL := rsp.HeaderMap.Get("X-Page-URL")
	if pageURL != "/test" {
		t.Errorf("Expecting page URL %s, got %s", "/test", pageURL)
	}

	body := rsp.Body.String()
	if body != "I'm a teapot!" {
		t.Errorf("Expecting body 'I'm a teapot!', got '%s'", body)
	}
}

func TestNewFragmentWriter_Basic(t *testing.T) {
	request := mockRequest("/test", TemplateContentType)
	rsp := httptest.NewRecorder()
	if tW, ok := NewFragmentWriter(rsp, request); ok {
		tW.Write([]byte("something here!"))
	} else {
		t.Error("Failed to recognize template request")
		return
	}

	if rsp.Code != 200 {
		t.Errorf("Expecting status 200 got %d", rsp.Code)
	}

	cType := rsp.HeaderMap.Get("Content-Type")
	if cType != TemplateContentType {
		t.Errorf("Expecting content type %s, got %s", TemplateContentType, cType)
	}

	pageURL := rsp.HeaderMap.Get("X-Page-URL")
	if pageURL != "" {
		t.Errorf("Expecting page URL not be set, got %s", pageURL)
	}

	body := rsp.Body.String()
	if body != "something here!" {
		t.Errorf("Expecting body 'something here!', got '%s'", body)
	}
}

func TestNewFragmentWriter_AddPageURL(t *testing.T) {
	request := mockRequest("/test", TemplateContentType)
	rsp := httptest.NewRecorder()
	if tW, ok := NewFragmentWriter(rsp, request); ok {
		tW.DesignatePageURL("/test-other")
		tW.Write([]byte("something here!"))
	} else {
		t.Error("Failed to recognize template request")
		return
	}

	if rsp.Code != 200 {
		t.Errorf("Expecting status 200 got %d", rsp.Code)
	}

	cType := rsp.HeaderMap.Get("Content-Type")
	if cType != TemplateContentType {
		t.Errorf("Expecting content type %s, got %s", TemplateContentType, cType)
	}

	pageURL := rsp.HeaderMap.Get("X-Page-URL")
	if pageURL != "/test-other" {
		t.Errorf("Expecting page URL %s, got %s", "/test-other", pageURL)
	}

	body := rsp.Body.String()
	if body != "something here!" {
		t.Errorf("Expecting body 'something here!', got '%s'", body)
	}
}

func Test_hexEscapeNonASCII(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "all ASCII",
			s:    "hello world!!",
			want: "hello world!!",
		},
		{
			name: "ASCII digits",
			s:    "abc123",
			want: "abc123",
		},
		{
			name: "Single white star, three byte range",
			s:    "☆",
			want: "%e2%98%86",
		},
		{
			name: "multile characters with diaeresis",
			s:    "äöü",
			want: "%c3%a4%c3%b6%c3%bc",
		},
		{
			name: "Latin small letter c with acute",
			s:    "ć",
			want: "%c4%87",
		},
		{
			name: "ASCII special characters",
			s:    "@*_+-./",
			want: "@*_+-./",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hexEscapeNonASCII(tt.s); got != tt.want {
				t.Errorf("hexEscapeNonASCII() = %v, want %v", got, tt.want)
			}
		})
	}
}
