package util

// run: go test -timeout 30s github.com/snipem/maniacforum/util

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

	assert.Equal(t, wrapAndformatQuote("> Test", 100), "[> Test](fg:red)")

	have := `Bla
> Test 123
Ergebnis`

	want := `Bla
[> Test](fg:red)
[123](fg:red)
Ergebnis`

	assert.Equal(t, want, wrapAndformatQuote(have, 7))
}

func TestQuoteFormattingNothingToFormat(t *testing.T) {
	assert.Equal(t, "nothing to format", wrapAndformatQuote("nothing to format", 100))

	origin := "[http://eins.de]"
	have := wrapAndformatQuote(origin, 100)
	assert.Equal(t, have, origin)

}

func TestExtractIDsFromLink(t *testing.T) {

	boardID, threadID, messageID, err := ExtractIDsFromLink("https://www.maniac-forum.de/forum/pxmboard.php?mode=board&brdid=6&thrdid=178514&msgid=4746825")
	assert.NoError(t, err)
	assert.Equal(t, "6", boardID)
	assert.Equal(t, "178514", threadID)
	assert.Equal(t, "4746825", messageID)

}
