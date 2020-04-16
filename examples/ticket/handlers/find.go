package handlers

import (
	"net/http"

	"github.com/rur/treetop"
)

// Method: GET
// Doc: search the database for a list of users matching a query string
func GetFindUserHandler(fieldName, callbackURL string) treetop.ViewHandlerFunc {
	return func(rsp treetop.Response, req *http.Request) interface{} {
		query := req.URL.Query()

		data := struct {
			FieldName   string
			CallbackURL string
			Results     []string
			QueryString string
		}{
			FieldName:   fieldName,
			CallbackURL: callbackURL,
			QueryString: query.Get("search-query"),
		}

		// For demo purposes, filter out any characters not in the latin alphabet.
		// All other characters must be in an allowlist, otherwise the result set will be empty
		filteredQuery := make([]byte, 0, len(data.QueryString))
	FILTER:
		for _, codePoint := range data.QueryString {
			if (codePoint >= 64 && codePoint <= 90) || (codePoint >= 97 && codePoint <= 122) {
				filteredQuery = append(filteredQuery, byte(codePoint))
				continue
			}
			switch codePoint {
			case ' ', '-', '_', '.', '\t':
				// allowed non latin alphabet character, skip for filter
				continue
			default:
				filteredQuery = nil
				break FILTER
			}
		}
		if len(filteredQuery) == 0 {
			return data
		}

		// For example purposes, vary number of results based
		// on the number of characters in the input query.
		for i := len(filteredQuery) - 1; i < 26; i++ {
			data.Results = append(data.Results, "Example User "+string(i+65))
			if len(data.Results) == 5 {
				break
			}
		}

		return data
	}
}
