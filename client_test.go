package treetop

import (
	"io/ioutil"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/rur/treetop/internal"
)

func TestServeClientLibrary(t *testing.T) {
	rsp := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/js/treetop.js", nil)
	req.Header.Set("Accept", "application/javascript")
	ServeClientLibrary.ServeHTTP(rsp, req)
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		t.Error(err)
		return
	}

	mod := rsp.Header().Get("last-modified")
	if !regexp.MustCompile(`^[\w]{3}, \d\d \w{3} 20\d\d [01]\d:[0-5]\d:[0-5]\d \w+$`).MatchString(mod) {
		t.Errorf("Unexpected timestamp format: '%s'", mod)
	}
	cType := rsp.Header().Get("content-type")
	if !strings.Contains(cType, "text/javascript") {
		t.Errorf("Expecting content type to match a JS file, got %s", cType)
	}

	cLen := rsp.Header().Get("Content-Length")
	if length, err := strconv.Atoi(cLen); err != nil || length != len(body) {
		t.Errorf("Expecting content length to match body %d, got %s", len(body), cLen)
	}

	if string(body) != internal.ScriptContent {
		t.Errorf("Expecting body to equal the script content, got:\n%s", string(body))
	}
}
