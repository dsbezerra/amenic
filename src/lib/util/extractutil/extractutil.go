package extractutil

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GetTrimmedText retrieves text from selection already trimmed.
func GetTrimmedText(sel *goquery.Selection) string {
	return strings.TrimSpace(sel.Text())
}

// IntIsNew ...
func IntIsNew(a, b int) bool {
	return a == 0 && b != 0
}
