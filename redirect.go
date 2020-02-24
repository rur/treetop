package treetop

import (
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"unicode/utf8"
)

// SeeOtherPage forces a treetop client to redirect the web browser to the supplied location.
// This allows the server to effectively 'break out' of in-page navigation, and direct the user elsewhere.
//
// The response body and content-type will be ignored by the treetop client handler.
// Similar to "303 See Other" an independent GET request should result.
func SeeOtherPage(w http.ResponseWriter, req *http.Request, location string) bool {
	if !IsTreetopRequest(req) {
		// not a treetop request, do nothing
		return false
	}
	// url handling lifted from golang http.Redirect
	// For more information see https://golang.org/src/net/http/server.go?s=60163:60228#L1998
	if u, err := url.Parse(location); err == nil {
		if u.Scheme == "" && u.Host == "" {
			oldpath := req.URL.Path
			if oldpath == "" {
				oldpath = "/"
			}

			// no leading http://server
			if location == "" || location[0] != '/' {
				// make relative path absolute
				olddir, _ := path.Split(oldpath)
				location = olddir + location
			}

			var query string
			if i := strings.Index(location, "?"); i != -1 {
				location, query = location[:i], location[i:]
			}

			// clean up but preserve trailing slash
			trailing := strings.HasSuffix(location, "/")
			location = path.Clean(location)
			if trailing && !strings.HasSuffix(location, "/") {
				location += "/"
			}
			location += query
		}
	}

	// This custom header instructs treetop client to load a new page.
	// Similar to the Location header in a standard HTTP redirect
	w.Header().Set("X-Treetop-See-Other", hexEscapeNonASCII(location))
	w.WriteHeader(http.StatusNoContent)
	return true
}

// Redirect is a helper that will instruct the Treetop client library to direct the web browser
// to a new URL. If the request is not from a Treetop client, the 3xx redirect method is used.
//
// This is necessary because 3xx HTTP redirects are opaque to XHR, when a full browser redirect
// is needed a 'X-Treetop-See-Other' header is used.
//
// Example:
// 		treetop.Redirect(w, req, "/some/other/path", http.StatusSeeOther)
//
func Redirect(w http.ResponseWriter, req *http.Request, location string, status int) {
	if ok := SeeOtherPage(w, req, location); !ok {
		http.Redirect(w, req, location, status)
	}
}

// hexEscapeNonASCII is used to sanitize header values.
//
// Lifted from golang codebase,
// see issue https://go-review.googlesource.com/c/go/+/31732
func hexEscapeNonASCII(s string) string {
	newLen := 0
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			newLen += 3
		} else {
			newLen++
		}
	}
	if newLen == len(s) {
		return s
	}
	b := make([]byte, 0, newLen)
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			b = append(b, '%')
			b = strconv.AppendInt(b, int64(s[i]), 16)
		} else {
			b = append(b, s[i])
		}
	}
	return string(b)
}
