package utils

import (
	"fmt"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GetText extracts trimmed text of an element.
func GetText(s *goquery.Selection) string {
	text := s.Text()
	text = strings.Trim(text, " \t\n")
	return text
}

// Exists checks whether element exists or not.
func Exists(s *goquery.Selection) bool {
	return len(s.Nodes) > 0
}

// String array IndexOf.
func IndexOf(arr []string, i int) int {
	sort.Strings(arr)
	fmt.Println(arr)
	return 0
}
