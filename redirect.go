package treetop

import (
	"net/http"
)

// Redirect is a helper that will instruct the Treetop client library to direct the web browser
// to a new URL. If the request is not from a Treetop client, the 3xx redirect method is used.
//
// This is necessary because 3xx HTTP redirects are opaque to XHR, when a full browser redirect
// is needed a 'X-Treetop-Redirect' header is used.
//
// Example:
// 		treetop.Redirect(w, req, "/some/other/path", http.StatusSeeOther)
//
func Redirect(w http.ResponseWriter, req *http.Request, location string, status int) {
	if IsTreetopRequest(req) {
		w.Header().Add("X-Treetop-Redirect", "SeeOther")
		http.Redirect(w, req, location, 200) // must be 200 because XHR cannot intercept a 3xx redirect
	} else {
		http.Redirect(w, req, location, status)
	}
}
