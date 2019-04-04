package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEnrichLinks(t *testing.T) {
	message := `Text
123
12345
[http://eins.de] testset
text [https://zwei.de] weiterer text
nox
[http://drei.de]`

	enrichedMessage, links := EnrichLinks(message)
	assert.Len(t, links, 3)

	assert.Equal(t, "[http://eins.de]", links[0])
	assert.Equal(t, "[https://zwei.de]", links[1])
	assert.Equal(t, "[http://drei.de]", links[2])

	assert.Contains(t, enrichedMessage, "[1][http://eins.de] testset")
	assert.Contains(t, enrichedMessage, "[2][https://zwei.de] weiterer text")
	assert.Contains(t, enrichedMessage, "[3][http://drei.de]")

}

func TestQuoteFormatting(t *testing.T) {
	formatted := FormatQuote("> Test")
	if formatted != "[> Test](fg:red)" {
		t.Errorf("Message was '%s' not expected", formatted)
	}

	if FormatQuote("nothing to format") != "nothing to format" {
		t.Errorf("Message was not expected")
	}

	have := `Bla
> Test 123
Ergebnis`

	want := `Bla
[> Test 123](fg:red)
Ergebnis`

	got := FormatQuote(have)

	if got != want {
		t.Errorf("Message was not expected")
	}
}

func TestQuoteFormattingNothingToFormat(t *testing.T) {
	// Check nothing to format
	origin := "[http://eins.de]"
	formatted := FormatQuote(origin)
	assert.Contains(t, formatted, origin)

}
