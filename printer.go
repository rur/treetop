package treetop

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
)

// PrintHandler is a cheap and cheerful way to debug a view handlers by 'stringing'
// a preview of the state of a supplied handler instance
func PrintHandler(h *Handler) string {
	var handlerInfo string
	if h.Fragment.HandlerFunc != nil {
		handlerInfo = runtime.FuncForPC(reflect.ValueOf(h.Fragment.HandlerFunc).Pointer()).Name()
	} else {
		handlerInfo = "nil"
	}
	if h.Page == nil {
		// only fragment requests can be handled
		return fmt.Sprintf("FragmentHandler(%s, %s)", previewTemplate(h.Fragment.Template, 10, 10), handlerInfo)
	} else if h.Fragment == nil {
		// only full page requests can be handled
		return fmt.Sprintf("PageHandler(%s, %s)", previewTemplate(h.Fragment.Template, 10, 10), handlerInfo)
	} else {
		// both full page and partial page requests can be handled
		return fmt.Sprintf("PartialHandler(%s, %s)", previewTemplate(h.Fragment.Template, 10, 10), handlerInfo)
	}
}

// previewTemplate previews an arbitrary template string on a single line.
// All whitespace will be stripped and it will be quoted and escaped.
// A middle ellipsis will be inserted if the string is too long.
func previewTemplate(str string, before, after int) string {
	re := regexp.MustCompile(`\s`)
	str = strconv.Quote(re.ReplaceAllString(str, ""))
	if len(str) > before+after+2 {
		return str[:before] + "……" + str[len(str)-after:]
	}
	return str
}
