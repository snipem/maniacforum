package util

import (
	"testing"
)

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
