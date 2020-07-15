package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnrichLinks(t *testing.T) {
	message := `Text
123
12345
[http://eins.de] testset
text [https://zwei.de] weiterer text
nox
[http://drei.de]`

	enrichedMessage, links := EnrichContent(message, 100)
	assert.Len(t, links, 3)

	assert.Equal(t, "[http://eins.de]", links[0])
	assert.Equal(t, "[https://zwei.de]", links[1])
	assert.Equal(t, "[http://drei.de]", links[2])

	assert.Contains(t, enrichedMessage, "[1]http://eins.de testset")
	assert.Contains(t, enrichedMessage, "[2]https://zwei.de weiterer text")
	assert.Contains(t, enrichedMessage, "[3]http://drei.de")

}

func TestFullEnriching(t *testing.T) {

	have := `Erste Zeile
> Link 123
not commented`

	want := `Erste
Zeile
[> Link](fg:red)
[123](fg:red)
not
commented`

	enriched, _ := EnrichContent(have, 7)

	assert.Equal(t, want, enriched)

}

func TestQuoteFormatting(t *testing.T) {

	assert.Equal(t, formatQuote("> Test"), "[> Test](fg:red)")

	have := `Bla
> Test 123
Ergebnis`

	want := `Bla
[> Test](fg:red)
[123](fg:red)
Ergebnis`

	assert.Equal(t, formatQuote(have), want)
}

func TestQuoteFormattingNothingToFormat(t *testing.T) {
	assert.Equal(t, formatQuote("nothing to format"), "nothing to format")

	origin := "[http://eins.de]"
	have := formatQuote(origin)
	assert.Equal(t, have, origin)

}
