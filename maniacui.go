//run: tmux send-keys -t right "C-c"; sleep 0.1; tmux send-keys -t right "go run maniacui.go" "C-m"
package main

import (
	"log"
	"strconv"

	"strings"

	"github.com/skratchdot/open-golang/open"
	"github.com/snipem/maniacforum/board"
	"github.com/snipem/maniacforum/util"

	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
)

var innerThreads board.Thread
var forum board.Board
var threadPanel *widgets.List
var messagePanel *widgets.Paragraph
var boardPanel *widgets.List
var threads []board.Thread
var message board.Message

func loadBoard() {
	forum := board.GetBoard("pxmboard.php?mode=threadlist&brdid=1&sortorder=last")
	threads = forum.Threads
}

func loadMessage() {
	if len(innerThreads.Messages) > 0 {
		message = board.GetMessage(innerThreads.Messages[threadPanel.SelectedRow].Link)
		message.EnrichedContent, message.Links = util.EnrichLinks(message.Content)
		messagePanel.Text = util.FormatQuote(message.EnrichedContent)
	}
}

// loadThread loads selected thread from board and displays the first message
func loadThread() {
	message := board.GetMessage(threads[boardPanel.SelectedRow].Link)
	innerThreads = board.GetThread(threads[boardPanel.SelectedRow].ID)

	// Clear thread panel
	threadPanel.Rows = nil
	threadPanel.SelectedRow = 0
	for _, message := range innerThreads.Messages {
		threadPanel.Rows = append(
			threadPanel.Rows,
			strings.Repeat("    ", message.Hierarchy-1)+
				"○ "+message.Topic+" ["+message.Date+" "+message.Author.Name+"](fg:white)")
	}
	messagePanel.Text = message.Content
}

func openLink(nr int) {
	link := message.Links[nr-1]
	cleanedLink := strings.Replace(link, "[", "", 1)
	cleanedLink = strings.Replace(cleanedLink, "]", "", 1)
	open.Run(cleanedLink)
}

func main() {

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	messagePanel = widgets.NewParagraph()
	boardPanel = widgets.NewList()
	threadPanel = widgets.NewList()

	loadBoard()

	boardPanel.Title = forum.Title

	for _, thread := range threads {
		boardPanel.Rows = append(boardPanel.Rows, thread.Title+" ["+thread.Date+"](fg:white)")
	}

	boardPanel.TextStyle = ui.NewStyle(ui.ColorRed)
	threadPanel.TextStyle = ui.NewStyle(ui.ColorRed)
	boardPanel.WrapText = false

	grid := ui.NewGrid()

	grid.Set(
		ui.NewCol(1.0/2,
			ui.NewRow(1.0/2, boardPanel),
			ui.NewRow(1.0/2, threadPanel),
		),
		ui.NewCol(1.0/2, messagePanel),
	)

	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	ui.Render(grid)

	previousKey := ""
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "1", "2", "3", "4", "5", "6", "7", "8", "9", "0":
			linkNr, _ := strconv.Atoi(e.ID)
			openLink(linkNr)
		case "q", "<C-c>":
			return
		case "J", "<Down>":
			boardPanel.ScrollDown()
			loadThread()
		case "K", "<Up>":
			boardPanel.ScrollUp()
			loadThread()
		case "j":
			threadPanel.ScrollDown()
			loadMessage()
		case "k":
			threadPanel.ScrollUp()
			loadMessage()
		case "<Enter>":
			loadThread()
		case "<C-d>":
			boardPanel.HalfPageDown()
		case "<C-u>":
			boardPanel.HalfPageUp()
		case "<C-f>":
			boardPanel.PageDown()
		case "<C-b>":
			boardPanel.PageUp()
		case "g":
			if previousKey == "g" {
				boardPanel.ScrollTop()
			}
		case "<Home>":
			boardPanel.ScrollTop()
		case "G", "<End>":
			boardPanel.ScrollBottom()
		}

		if previousKey == "g" {
			previousKey = ""
		} else {
			previousKey = e.ID
		}

		ui.Render(boardPanel, messagePanel, threadPanel)
	}
}
