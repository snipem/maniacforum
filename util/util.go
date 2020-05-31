package util

import (
	"regexp"
	"strconv"
	"strings"
)

// EnrichLinks enriches links in content with numbers, returns enriched content and list of links
func EnrichLinks(content string) (string, []string) {
	var enrichedContent = content
	var links []string
	r := regexp.MustCompile(`\[.*\]`)
	links = r.FindAllString(content, -1)

	for i := 0; i < len(links); i++ {
		enrichedContent = strings.Replace(enrichedContent, links[i], "["+strconv.Itoa(i+1)+"]"+links[i], 1)
	}

	return enrichedContent, links

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
