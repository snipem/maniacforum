package board

import (
	"strings"
	"testing"
)

func TestMessage(t *testing.T) {
	message := GetMessage("pxmboard.php?mode=message&brdid=1&msgid=4460819")
	t.Log(message.Content)
	t.Log(message.Link)
	t.Log("tests how me")
	if !strings.Contains(message.Content, "Tetris") {
		t.Error("Not yet working")
	}
}
