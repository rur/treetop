package treetop

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
)

// PrintTemplateHandler is a cheap and cheerful way to debug a view handlers by 'stringing'
// a preview of the state of a supplied handler instance
func PrintTemplateHandler(h *TemplateHandler) string {
	var handlerInfo string
	if h.Partial.Handler != nil {
		handlerInfo = runtime.FuncForPC(reflect.ValueOf(h.Partial.Handler).Pointer()).Name()
	} else {
		handlerInfo = "nil"
	}
	if h.Page == nil {
		// only fragment requests can be handled
		return fmt.Sprintf("PartialHandler(%s, %s)", previewTemplate(h.Partial.Template, 10, 10), handlerInfo)
	} else if h.Partial == nil {
		// only full page requests can be handled
		return fmt.Sprintf("PageHandler(%s, %s)", previewTemplate(h.Partial.Template, 10, 10), handlerInfo)
	} else {
		// both full page and partial page requests can be handled
		return fmt.Sprintf("PartialHandler(%s, %s)", previewTemplate(h.Partial.Template, 10, 10), handlerInfo)
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
