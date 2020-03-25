package ticket

import (
	"regexp"
	"strings"
)

var (
	wsRegex = regexp.MustCompile(`\s+`)
)

// remove redundant whitespace from a string that is to be used as a visual summary
func sanitizeSummary(s string) string {
	return strings.TrimSpace(wsRegex.ReplaceAllString(s, " "))
}
