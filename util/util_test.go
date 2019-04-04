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

	assert.Equal(t, FormatQuote("> Test"), "[> Test](fg:red)")

	have := `Bla
> Test 123
Ergebnis`

	want := `Bla
[> Test 123](fg:red)
Ergebnis`

	assert.Equal(t, FormatQuote(have), want)
}

func TestQuoteFormattingNothingToFormat(t *testing.T) {
	assert.Equal(t, FormatQuote("nothing to format"), "nothing to format")

	origin := "[http://eins.de]"
	have := FormatQuote(origin)
	assert.Equal(t, have, origin)

}
