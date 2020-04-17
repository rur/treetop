package inputs

func SearchForUser(queryString string) []string {
	// For demo purposes, filter out any characters not in the latin alphabet.
	// All other characters must be in an allowlist, otherwise the result set will be empty
	filteredQuery := make([]byte, 0, len(queryString))
FILTER:
	for _, codePoint := range queryString {
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
		return nil
	}

	var results []string
	// For example purposes, vary number of results based
	// on the number of characters in the input query.
	for i := len(filteredQuery) - 1; i < 26; i++ {
		results = append(results, "Example User "+string(i+65))
		if len(results) == 5 {
			break
		}
	}
	return results
}
