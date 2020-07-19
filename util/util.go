package util

import (
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// EnrichContent enriches links in content with numbers, returns enriched content and list of links
// It also wraps the text at the number of characters at wrapAt
func EnrichContent(content string, wrapAt int) (string, []string) {
	var enrichedContent = content
	var links []string
	r := regexp.MustCompile(`\[.*\]`)
	links = r.FindAllString(content, -1)

	for i := 0; i < len(links); i++ {
		// TODO make prettier
		cleanLink := strings.ReplaceAll(links[i], "[", "")
		cleanLink = strings.ReplaceAll(cleanLink, "]", "")
		enrichedContent = strings.Replace(enrichedContent, links[i], "["+strconv.Itoa(i+1)+"]"+cleanLink, 1)
	}

	// Use own wrapper because termui's wrapping functionality has a bug and does cut content at the end
	wrappedContent := wrapAndformatQuote(enrichedContent, wrapAt)

	return wrappedContent, links

}

// wrapAndformatQuote returns a wrapped and quote formatted string.
// lines are also wrapped and coloured if they were originally coming from a
// quotation
func wrapAndformatQuote(unformatted string, wrapAt int) string {
	r := regexp.MustCompile("^>.*$")
	formatted := ""
	for _, line := range strings.Split(strings.TrimSuffix(unformatted, "\n"), "\n") {
		wrappedLines := wrapLine(line, wrapAt)
		if r.MatchString(line) { // if unwrapped line is a quote
			for _, wrappedLine := range wrappedLines {
				formatted = formatted + "[" + wrappedLine + "](fg:red)" + "\n"
			}
		} else { // just wrapp lines and append them
			for _, wrappedLine := range wrappedLines {
				formatted = formatted + wrappedLine + "\n"
			}
		}
	}
	return strings.TrimSpace(formatted)
}

// wrapLine returns wrapped lines from unwrapped string
// at nth character defined by wrapAt
func wrapLine(unwrapped string, wrapAt int) []string {
	var wrapped []string
	for _, word := range strings.Split(unwrapped, " ") {
		if len(wrapped) == 0 { // if new line is empty so far
			wrapped = append(wrapped, word)
		} else if len(wrapped[len(wrapped)-1]+" "+word) <= wrapAt { // if last wrapped line does not exceed the wrapAt limit
			wrapped[len(wrapped)-1] = wrapped[len(wrapped)-1] + " " + word
		} else { // start a new line if the wrapAt limit was exceeded
			wrapped = append(wrapped, word)
		}
	}

	return wrapped
}

// HandleMouseClickEventOnTabBar returns true if the TabPane has been clicked, and sets the
// ActiveTabIndex of the TabPane to the tab clicked
func HandleMouseClickEventOnTabBar(e ui.Event, bar *widgets.TabPane) bool {
	payload := e.Payload.(ui.Mouse)
	x0, y0 := bar.Inner.Min.X, bar.Inner.Min.Y
	x1, y1 := bar.Inner.Max.X, bar.Inner.Max.Y
	if x0 <= payload.X && payload.X <= x1 && y0 <= payload.Y && payload.Y <= y1 {

		// Calculate clicked tab by splitting up the whole string bar "Smalltalk | For Sale | ... "
		// at the Y position of the mouse event. The number of | in the resulting string will reflect
		// the clicked tab
		wholeTabBarString := strings.Join(bar.TabNames, " | ")
		tabNrClicked := strings.Count(wholeTabBarString[0:payload.X], "|")

		bar.ActiveTabIndex = tabNrClicked
		return true
	}
	return false
}

// HandleMouseClickEventOnList returns true if the list has been clicked and set's the active
// element of the list to the element clicked.
func HandleMouseClickEventOnList(e ui.Event, list *widgets.List) bool {

	payload := e.Payload.(ui.Mouse)

	border := 0
	if list.BorderTop {
		border = 1
	}
	x0, y0 := list.Inner.Min.X, list.Inner.Min.Y
	x1, y1 := list.Inner.Max.X, list.Inner.Max.Y
	if x0 <= payload.X && payload.X <= x1 && y0 <= payload.Y && payload.Y <= y1 {
		list.SelectedRow = payload.Y - list.Rectangle.Min.Y - border + list.TopRow
		return true
	}
	return false

}

func ExtractIDsFromLink(link string) (boardID string, threadID string, messageID string) {
	u, err := url.Parse(link)
	if err != nil {
		log.Fatal(err)
	}
	// https://www.maniac-forum.de/forum/pxmboard.php?mode=board&brdid=6&thrdid=178514&msgid=4746825"
	boardID = u.Query().Get("brdid")
	threadID = u.Query().Get("thrdid")
	messageID = u.Query().Get("msgid")

	return
}
