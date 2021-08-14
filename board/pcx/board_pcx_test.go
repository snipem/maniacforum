package pcx

import (
	"github.com/snipem/maniacforum/board"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var f *board.Forum
const boardNumber = "6"

func TestMain(m *testing.M) {
	var err error
	f, err = board.GetForum(board.PCXBoardURL, false)
	if err != nil {
		log.Fatal(err)
	}
	code:= m.Run()
	os.Exit(code)
}

func TestPCXForum(t *testing.T) {
	assert.Len(t, f.Boards,1, "PCX has one board")
	assert.Equal(t, boardNumber, f.Boards[0].ID, "Smalltalk board has this id")
}

func TestBoard(t *testing.T) {

	forum, err := f.GetBoard(boardNumber)
	assert.NoError(t, err)
	threads := forum.Threads

	assert.Contains(t, threads[0].Date, ":")
	// Expect first entry to be sticky
	assert.False(t, threads[0].IsSticky)

	assert.Equal(t, boardNumber, forum.ID)
	assert.Equal(t, "Smalltalk", forum.Title)

	// Expect tenth entry to be not sticky
	assert.False(t, threads[9].IsSticky)

}

func TestThread(t *testing.T) {
	thread, err := f.GetThread("1682", boardNumber)
	assert.NoError(t, err)
	if len(thread.Messages) == 0 {
		t.Errorf("No messages returned")
	}
	t.Log(thread.Messages)
	assert.Equal(t, "21.09.20 14:57", thread.Messages[0].Date)
}

func TestMessage(t *testing.T) {
	message, err := f.GetMessage("pxmboard.php?mode=message&brdid=6&msgid=82602")
	assert.Nil(t, err)
	t.Log("Message: ", message.Content)
	t.Log("Link: ", message.Link)
	expected := "Ich bilde mir ja ein, schon 1996 im PCX-Forum gewesen zu sein. Jetzt also ein Vierteljahrhundert her."
	if !strings.Contains(message.Content, expected) {
		t.Errorf("Message does not match, was '%s', expected '%s'", message.Content, expected)
	}
	expectedAuthor := "Pascal Parvex"
	if expectedAuthor != message.Author.Name {
		t.Errorf("Author does not match, was '%s', expected '%s'", message.Author.Name, expectedAuthor)
	}

	expectedTopic := "1996 - 2021: 25 Jahre PCX-Forum, ihr alten Säcke"
	if expectedTopic != message.Topic {
		t.Errorf("Topic does not match, was '%s', expected '%s'", message.Topic, expectedTopic)
	}

	assert.Equal(t, "82602", message.ID)
}
