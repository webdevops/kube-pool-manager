package manager

import "strings"

func stringCompare(a, b string) bool {
	return strings.EqualFold(a, b)
}
