package treetop

import (
	"fmt"
	"io"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// SprintViewInfo will create a string preview of view
func SprintViewInfo(v *View) string {
	if v == nil {
		return "nil"
	}
	handlerInfo := "nil"
	if v.HandlerFunc != nil {
		handlerInfo = runtime.
			FuncForPC(reflect.
				ValueOf(v.HandlerFunc).
				Pointer()).
			Name()
	}
	if v.Defines == "" {
		return fmt.Sprintf(
			"View(%s, %v)",
			previewString(v.Template, 20, 30),
			previewString(handlerInfo, 20, 30),
		)
	}
	return fmt.Sprintf(
		"SubView(%#v, %s, %v)",
		v.Defines,
		previewString(v.Template, 20, 30),
		previewString(handlerInfo, 20, 30),
	)
}

// previewString previews an arbitrary string on a single line.
// All whitespace will be stripped and it will be quoted and escaped.
// A middle ellipsis will be inserted if the string is too long.
func previewString(str string, before, after int) string {
	re := regexp.MustCompile(`\s`)
	str = strconv.Quote(re.ReplaceAllString(str, ""))
	if len(str) > before+after+2 {
		return str[:before] + "……" + str[len(str)-after:]
	}
	return str
}

// SprintViewTree create a string with a tree representation of a a view hierarchy
func SprintViewTree(v *View) string {
	if v == nil {
		return "- nil"
	}
	str := strings.Builder{}
	str.WriteString("- ")
	str.WriteString(SprintViewInfo(v))
	fprintViewTree(&str, []byte("  "), v.SubViews)
	return str.String()
}

// fprintViewTree delves recursively into view and sub views and writes
// a tree prepresentation of the supplied view
func fprintViewTree(w io.Writer, prefix []byte, views map[string]*View) {
	subCount := len(views)
	keys := make([]string, subCount)
	{
		i := 0
		for k := range views {
			keys[i] = k
			i++
		}
		sort.Strings(keys)
	}
	for i, k := range keys {
		last := i == len(keys)-1
		sub := views[k]
		w.Write(append([]byte{'\n'}, prefix...))
		if last {
			w.Write([]byte("'- " + k + ": " + SprintViewInfo(sub)))
		} else {
			w.Write([]byte("|- " + k + ": " + SprintViewInfo(sub)))
		}
		if sub != nil {
			var subPrefix []byte
			if last {
				subPrefix = append(prefix, []byte("   ")...)
			} else {
				subPrefix = append(prefix, []byte("|  ")...)
			}
			fprintViewTree(w, subPrefix, sub.SubViews)
		}
		if last && (sub == nil || len(sub.SubViews) == 0) {
			// add padding to mark end of a branch
			// add the prefix to padding, without trailing spaces
			for j := len(prefix) - 1; j > -1; j-- {
				if prefix[j] != ' ' {
					w.Write([]byte("\n"))
					w.Write(prefix[:j+1])
					break
				}
			}
		}
	}
}
