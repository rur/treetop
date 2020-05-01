package inline

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	emailRE = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	nameRE  = regexp.MustCompile(`^[\w'\-,.][^0-9_!¡?÷?¿/\\+=@#$%ˆ&*(){}|~<>;:[\]]{2,}$`)
)

func processInputName(form url.Values, field string) (string, string) {
	value := form.Get(field)
	if len(value) > 255 {
		return value, "name is too large"
	}
	if nameRE.MatchString(value) {
		return value, ""
	}
	return value, "Invalid name"
}

func processInputEmail(form url.Values, field string) (string, string) {
	value := form.Get(field)
	confirm := form.Get(field + "_confirm")
	if len(value) > 255 {
		return value, "email address is too large"
	}
	if strings.ToLower(value) != strings.ToLower(confirm) {
		return value, "Email address does not match confirmation address."
	}
	if emailRE.MatchString(value) {
		return value, ""
	}
	return value, "Invalid email"
}

func processInputContry(form url.Values, field string) (string, string) {
	value := form.Get(field)
	for _, v := range CountryOptions {
		if value == v {
			return value, ""
		}
	}
	return value, "Unknown country, expecting: " + strings.Join(CountryOptions, ", ")
}

func processInputDescription(form url.Values, field string) (string, string) {
	value := form.Get(field)
	if len(value) > 1000 {
		return value, "Message is too large, max 1000 character bytes"
	}
	return value, ""
}
