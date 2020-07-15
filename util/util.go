package util

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/mitchellh/go-wordwrap"
)

// EnrichContent enriches links in content with numbers, returns enriched content and list of links
// It also wraps the text at the number of characters at wrapAt
func EnrichContent(content string, wrapAt int) (string, []string) {
	var enrichedContent = content
	var links []string
	r := regexp.MustCompile(`\[.*\]`)
	links = r.FindAllString(content, -1)

	for i := 0; i < len(links); i++ {
		// TODO make prettier
		cleanLink := strings.ReplaceAll(links[i], "[", "")
		cleanLink = strings.ReplaceAll(cleanLink, "]", "")
		enrichedContent = strings.Replace(enrichedContent, links[i], "["+strconv.Itoa(i+1)+"]"+cleanLink, 1)
	}

	// Use own wrapper because termui's wrapping functionality does cut content at the end
	wrappedContent := wordwrap.WrapString(enrichedContent, uint(wrapAt))

	return wrappedContent, links

}

// FormatQuote formats a quote with TermUi specific color formatting
func FormatQuote(unformatted string) string {
	r := regexp.MustCompile("^>.*$")
	formatted := ""
	for _, line := range strings.Split(strings.TrimSuffix(unformatted, "\n"), "\n") {
		if r.MatchString(line) {
			formatted = formatted + "[" + line + "](fg:red)" + "\n"
		} else {
			formatted = formatted + line + "\n"
		}
	}
	return strings.TrimSpace(formatted)
}
