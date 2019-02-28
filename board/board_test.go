package board

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestThread(t *testing.T) {
	thread := GetThread("173448")
	if len(thread.Messages) == 0 {
		t.Errorf("No messages returned")
	}
	t.Log(thread.Messages)
	assert.Equal(t, "21.02.19 23:16", thread.Messages[0].Date)
}

func TestMessage(t *testing.T) {
	message := GetMessage("pxmboard.php?mode=message&brdid=1&msgid=4377586")
	t.Log("Message: ", message.Content)
	t.Log("Link: ", message.Link)
	expected := "Trophy-Sharing bedeutet nicht zwingend Cross-Buy"
	if !strings.Contains(message.Content, expected) {
		t.Errorf("Message does not match, was '%s', expected '%s'", message.Content, expected)
	}
	expectedAuthor := "snimat"
	if expectedAuthor != message.Author.Name {
		t.Errorf("Author does not match, was '%s', expected '%s'", message.Author.Name, expectedAuthor)
	}

	expectedTopic := "Re:Sigi kein Cross-Buy?"
	if expectedTopic != message.Topic {
		t.Errorf("Topic does not match, was '%s', expected '%s'", message.Topic, expectedTopic)
	}

	expectedAuthorID := 54889
	if expectedAuthorID != message.Author.ID {
		t.Errorf("Author Id does not match, was '%d', expected '%d'", message.Author.ID, expectedAuthorID)
	}
}
