package util

import (
	"regexp"
	"strings"
)

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
