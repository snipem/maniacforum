package board

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestForum(t *testing.T) {

	maniacforum := GetForum()
	assert.Equal(t, "1", maniacforum.Boards[0].ID)
	assert.Equal(t, "2", maniacforum.Boards[1].ID)
	assert.Equal(t, "4", maniacforum.Boards[2].ID)
	assert.Equal(t, "6", maniacforum.Boards[3].ID)
	assert.Equal(t, "26", maniacforum.Boards[4].ID)
	// Boards after this might change due to events like E3 or WM
}

func TestBoard(t *testing.T) {
	forum := GetBoard("1")
	threads := forum.Threads

	// TODO Flaky, because it sticks to the sticky note
	// TODO Extract number of responses from Date
	assert.Contains(t, threads[0].Date, "02.12.17 10:32")
	// Expect first entry to be sticky
	assert.True(t, threads[0].IsSticky)

	assert.Equal(t, "1", forum.ID)
	assert.Equal(t, "Smalltalk", forum.Title)

	// Expect tenth entry to be not sticky
	assert.False(t, threads[9].IsSticky)

}

func TestThread(t *testing.T) {
	thread := GetThread("173448", "1")
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

	assert.Equal(t, "4377586", message.ID)
}
