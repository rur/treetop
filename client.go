package treetop

import (
	"bytes"
	"net/http"
	"strconv"
	"time"

	"github.com/rur/treetop/internal"
)

var (
	// ServeClientLibrary TODO: add docs
	ServeClientLibrary http.Handler
)

func init() {
	ts, err := strconv.ParseInt(internal.Modified, 10, 64)
	if err != nil {
		panic(err)
	}
	modTime := time.Unix(ts, 0)
	if err != nil {
		panic(err)
	}
	content := bytes.NewReader([]byte(internal.ScriptContent))

	ServeClientLibrary = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "text/javascript; charset=utf-8")
		http.ServeContent(rw, r, "treetop.js", modTime, content)
	})
}
