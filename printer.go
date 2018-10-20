package treetop

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
)

func PrintHandler(h *Handler) string {
	var handlerInfo string
	if h.Partial.HandlerFunc != nil {
		handlerInfo = runtime.FuncForPC(reflect.ValueOf(h.Partial.HandlerFunc).Pointer()).Name()
	} else {
		handlerInfo = "nil"
	}
	if h.Page == nil {
		// only fragment requests can be handled
		return fmt.Sprintf("FragmentHandler(%s, %s)", previewTemplate(h.Partial.Template, 10, 10), handlerInfo)
	} else if h.Partial == nil {
		// only full page requests can be handled
		return fmt.Sprintf("PageHandler(%s, %s)", previewTemplate(h.Partial.Template, 10, 10), handlerInfo)
	} else {
		// both full page and partial page requests can be handled
		return fmt.Sprintf("PartialHandler(%s, %s)", previewTemplate(h.Partial.Template, 10, 10), handlerInfo)
	}
}

// Used to preview an arbitrary template string on a single line. A middle ellipsis will be inserted
// if the string is too long. All whitespace is stripped and quotes are escaped.
func previewTemplate(str string, before, after int) string {
	re := regexp.MustCompile(`\s`)
	str = strconv.Quote(re.ReplaceAllString(str, ""))
	if len(str) > before+after+2 {
		return str[:before] + "……" + str[len(str)-after:]
	} else {
		return str
	}
}
