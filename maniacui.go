//run: tmux send-keys -t right "C-c"; sleep 0.1; tmux send-keys -t right "go run maniacui.go" "C-m"; tmux select-pane -t right
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

// TODO Get rid of global variables
var innerThreads board.Thread
var forum board.Board

var threadPanel *widgets.List
var messagePanel *widgets.Paragraph
var boardPanel *widgets.List
var tabpane *widgets.TabPane

var threads []board.Thread
var message board.Message

func loadBoard() {
	tabNr := strconv.Itoa(tabpane.ActiveTabIndex + 1)
	forum = board.GetBoard(tabNr)
	threads = forum.Threads

	// Clear board panel
	boardPanel.Rows = nil
	messagePanel.Text = ""
	threadPanel.Rows = nil

	boardPanel.SelectedRow = 0
	threadPanel.SelectedRow = 0

	for _, thread := range threads {
		boardPanel.Rows = append(boardPanel.Rows, thread.Title+" ["+thread.Date+"](fg:white)")
	}
}

func loadMessage() {
	if len(innerThreads.Messages) > 0 {
		message = board.GetMessage(innerThreads.Messages[threadPanel.SelectedRow].Link)
		message.EnrichedContent, message.Links = util.EnrichLinks(message.Content)
		messagePanel.Text = util.FormatQuote(message.EnrichedContent)
	}
}

func answer() {
	open.Run(board.BoardURL + "pxmboard.php?mode=messageform&brdid=" + forum.ID + "&msgid=" + message.ID)
}

// loadThread loads selected thread from board and displays the first message
func loadThread() {
	message = board.GetMessage(threads[boardPanel.SelectedRow].Link)
	innerThreads = board.GetThread(threads[boardPanel.SelectedRow].ID, forum.ID)

	// Clear thread panel
	threadPanel.Rows = nil
	threadPanel.SelectedRow = 0
	for _, m := range innerThreads.Messages {
		threadPanel.Rows = append(
			threadPanel.Rows,
			strings.Repeat("    ", m.Hierarchy-1)+
				"○ "+m.Topic+" ["+m.Date+" "+m.Author.Name+"](fg:white)")
	}
	message.EnrichedContent, message.Links = util.EnrichLinks(message.Content)
	messagePanel.Text = util.FormatQuote(message.EnrichedContent)
}

func openLink(nr int) {
	link := message.Links[nr-1]
	cleanedLink := strings.Replace(link, "[", "", 1)
	cleanedLink = strings.Replace(cleanedLink, "]", "", 1)
	open.Run(cleanedLink)
}

func initialize() {
	// Initialize
	loadBoard()
	loadThread()
}

func main() {

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	messagePanel = widgets.NewParagraph()
	boardPanel = widgets.NewList()
	threadPanel = widgets.NewList()

	tabpane = widgets.NewTabPane("Smalltalk", "O/T")
	// tabpane.SetRect(0, 1, 50, 4)
	tabpane.Border = false
	tabpane.ActiveTabIndex = 0

	initialize()

	// boardPanel.Title = forum.Title

	boardPanel.TextStyle = ui.NewStyle(ui.ColorRed)
	threadPanel.TextStyle = ui.NewStyle(ui.ColorRed)
	boardPanel.WrapText = false

	grid := ui.NewGrid()

	grid.Set(
		ui.NewRow(0.05, tabpane),
		ui.NewRow(0.95,
			ui.NewCol(1.0/2,
				ui.NewRow(0.5, boardPanel),
				ui.NewRow(0.5, threadPanel),
			),
			ui.NewCol(1.0/2, messagePanel),
		),
	)

	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	ui.Render(grid)

	renderTab := func() {
		switch tabpane.ActiveTabIndex {
		case 0:
			ui.Render(grid)
		case 1:
			ui.Render(grid)
		}
	}

	previousKey := ""
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "1", "2", "3", "4", "5", "6", "7", "8", "9", "0":
			linkNr, _ := strconv.Atoi(e.ID)
			openLink(linkNr)
		case "a":
			answer()
		case "q", "<C-c>":
			return
		case "b":
			tabpane.FocusLeft()
			ui.Clear()
			renderTab()
			initialize()
		case "n":
			tabpane.FocusRight()
			ui.Clear()
			renderTab()
			initialize()
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
			boardPanel.ScrollHalfPageDown()
		case "<C-u>":
			boardPanel.ScrollHalfPageUp()
		case "<C-f>":
			boardPanel.ScrollPageDown()
		case "<C-b>":
			boardPanel.ScrollPageUp()
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

		renderTab()
		ui.Render(boardPanel, messagePanel, threadPanel, tabpane)

	}
}
