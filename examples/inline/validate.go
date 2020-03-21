package inline

import (
	"regexp"
	"strings"
)

var (
	emailRE = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	nameRE  = regexp.MustCompile("^[\\w\\- .]{2,}$")
)

func assertValidName(value string) string {
	if len(value) > 255 {
		return "name is too large"
	}
	if nameRE.MatchString(value) {
		return ""
	}
	return "invalid name"
}

func assertValidEmail(value string) string {
	if len(value) > 255 {
		return "email address is too large"
	}
	if emailRE.MatchString(value) {
		return ""
	}
	return "invalid email"
}

func assertValidContry(value string) string {
	for _, v := range CountryOptions {
		if value == v {
			return ""
		}
	}
	return "Unknown country, expecting: " + strings.Join(CountryOptions, ", ")
}

func assertValidDescription(value string) string {
	if len(value) > 255 {
		return "message is too large"
	}
	return ""
}
