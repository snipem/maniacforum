package util

import (
	"regexp"
	"strconv"
	"strings"
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

	// Use own wrapper because termui's wrapping functionality has a bug and does cut content at the end
	wrappedContent := wrapAndformatQuote(enrichedContent, wrapAt)

	return wrappedContent, links

}

// wrapAndformatQuote returns a wrapped and quote formatted string.
// lines are also wrapped and coloured if they were originally coming from a
// quotation
func wrapAndformatQuote(unformatted string, wrapAt int) string {
	r := regexp.MustCompile("^>.*$")
	formatted := ""
	for _, line := range strings.Split(strings.TrimSuffix(unformatted, "\n"), "\n") {
		wrappedLines := wrapLine(line, wrapAt)
		if r.MatchString(line) { // if unwrapped line is a quote
			for _, wrappedLine := range wrappedLines {
				formatted = formatted + "[" + wrappedLine + "](fg:red)" + "\n"
			}
		} else { // just wrapp lines and append them
			for _, wrappedLine := range wrappedLines {
				formatted = formatted + wrappedLine + "\n"
			}
		}
	}
	return strings.TrimSpace(formatted)
}

// wrapLine returns wrapped lines from unwrapped string
// at nth character defined by wrapAt
func wrapLine(unwrapped string, wrapAt int) []string {
	var wrapped []string
	for _, word := range strings.Split(unwrapped, " ") {
		if len(wrapped) == 0 { // if new line is empty so far
			wrapped = append(wrapped, word)
		} else if len(wrapped[len(wrapped)-1]+" "+word) <= wrapAt { // if last wrapped line does not exceed the wrapAt limit
			wrapped[len(wrapped)-1] = wrapped[len(wrapped)-1] + " " + word
		} else { // start a new line if the wrapAt limit was exceeded
			wrapped = append(wrapped, word)
		}
	}

	return wrapped
}
