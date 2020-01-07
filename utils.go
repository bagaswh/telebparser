package telebparser

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GetText extracts trimmed text of an element.
func getText(s *goquery.Selection) string {
	text := s.Text()
	text = strings.Trim(text, " \t\n")
	return text
}

// Exists checks whether element exists or not.
func exists(s *goquery.Selection) bool {
	return len(s.Nodes) > 0
}
