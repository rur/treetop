package inline

// Emulate serverside persistence using an encoded cookie
//

import (
	"fmt"
	"net/http"
	"sync"
	"text/template"
	"time"

	"github.com/rur/treetop"
)

const formDataCookieName = "inline-form-data"

// cookieServer emulates server persistence using a HTTP cookie.
// There is a limit to the amount of data we can store (<4k);
// it's good enough for a demo
type cookieServer struct {
	sync.RWMutex
	// per-request resource cache. Multiple handlers may be called for a given
	// request, cookieServer ensures they share the same FormData instances
	cache map[int]*FormData
}

func newCookieServer() *cookieServer {
	return &cookieServer{
		cache: make(map[int]*FormData),
	}
}

// get from cache data for this request with an 'found' flag
func (srv *cookieServer) get(respID int) (*FormData, bool) {
	srv.RLock()
	defer srv.RUnlock()
	fd, ok := srv.cache[respID]
	return fd, ok
}

// put form in cache for a given tt request ID
func (srv *cookieServer) put(respID int, fd *FormData) {
	srv.Lock()
	srv.cache[respID] = fd
	srv.Unlock()
}

// teardown frees any request scoped resources, in this demo we purge
// form data from the response cache
func (srv *cookieServer) teardown(respID int) {
	srv.Lock()
	delete(srv.cache, respID)
	srv.Unlock()
}

// formDataHandlerFunc is a treetop request handler that expects a FormData
// instance as it's first argument.
type formDataHandlerFunc func(fd *FormData, rsp treetop.Response, req *http.Request) interface{}

// bind is middleware function that wraps a treetop view handler function.
// The middleward will decode a FormData instance from the request cookies
// and pass it as the first argument of the wrapped handler.
func (srv *cookieServer) bind(hdl formDataHandlerFunc) treetop.ViewHandlerFunc {
	return func(rsp treetop.Response, req *http.Request) interface{} {
		var (
			data *FormData
			rID  = int(rsp.ResponseID())
		)

		if fd, ok := srv.get(rID); ok {
			// Another handler has already loaded the resources for this request
			return hdl(fd, rsp, req)
		}

		if cookie, err := req.Cookie(formDataCookieName); err == nil {
			data = &FormData{}
			err := data.UnmarshalBase64([]byte(cookie.Value))
			if err != nil {
				http.Error(
					rsp,
					template.HTMLEscapeString(fmt.Sprintf("Error unmarshalling cookie base64 data, %s", err)),
					http.StatusBadRequest,
				)
				return nil
			}
		} else if req.Method == "GET" || req.Method == "HEAD" {
			// There is no form data cookie so set a default (read-only requests)
			//
			// Note: This is the similar to an 'unauthenticated' path. In that case
			//       a redirect might be more appropriate.
			fDataCopy := defaultFormData
			data = &fDataCopy
			err := setCookieFormData(rsp, data)
			if err != nil {
				http.Error(
					rsp,
					template.HTMLEscapeString(fmt.Sprintf("Error commiting cookie base64 data, %s", err)),
					http.StatusInternalServerError,
				)
				return nil
			}
		} else {
			pageErrorMessage(rsp, req,
				"Missing cookie, if cookies are disabled the demo isn't going to work.",
				http.StatusBadRequest)
			return nil
		}
		if data != nil {
			// add it to the cache
			srv.put(rID, data)
			go func() {
				// When the Treetop response context is done free any sever resources
				// associated with this request.
				<-rsp.Context().Done()
				srv.teardown(rID)
			}()
		}
		return hdl(data, rsp, req)
	}
}

// setCookieFormData will set the form data cookie in the response headers.
// Note: like any header this must be called before the headers are written
func setCookieFormData(w http.ResponseWriter, data *FormData) error {
	b64, err := data.MarshalBase64()
	if err != nil {
		return err
	}
	cookie := http.Cookie{
		Name:     formDataCookieName,
		Path:     "/inline",
		Value:    string(b64),
		Expires:  time.Now().Add(1 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)
	return nil
}
