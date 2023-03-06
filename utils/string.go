package utils

import "strings"

func ContainsI(a string, b string) (string, string) {
	contain := strings.Contains(
		strings.ToLower(a),
		strings.ToLower(b),
	)

	if contain {
		return strings.ToLower(a), strings.ToLower(b)
	}
	return a, ""
}
