package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"strings"

	"github.com/skratchdot/open-golang/open"
	"github.com/snipem/maniacforum/board"
	"github.com/snipem/maniacforum/util"

	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
)

// Forum > Board > Threads > Message

// TODO Get rid of global variables
var activeThreads board.Thread
var activeBoard board.Board
var activeForum board.Forum

var threadPanel *widgets.List
var messagePanel *widgets.List
var boardPanel *widgets.List
var tabpane *widgets.TabPane

var threads []board.Thread
var message board.Message

var activePane int
var maxPane = 3

var version = "1.2"

var helpPage = `maniacforum ` + version + `

Hilfe
======

Kontext-Steuerung
------------------

<Tab> - Fokus-Wechsel auf Boards, Threads, Unterthreads und Nachrichten
 ↑ ↓  - Zur Auswahl im aktuellen ausgewählten Menü
 ← →  - Auswahl des Boards
   a  - Auf Nachricht im Standard-Browser antworten
   ?  - Hilfsseite
   q  - Beenden
 0-9  - Links im Standard-Browser öffnen

Globale Steuerung
-----------------

   j  - Nächster Unterthread
   k  - Vorheriger Unterthread
   u  - Nächster ungelesener Unterthread
   z  - Nächster Thread
   J  - Nächster Thread
   K  - Vorheriger Thread

https://github.com/snipem/maniacforum
`

func loadBoard() {
	boardID := activeForum.Boards[tabpane.ActiveTabIndex].ID
	activeBoard = board.GetBoard(boardID)
	threads = activeBoard.Threads

	// Clear board panel
	boardPanel.Rows = nil
	messagePanel.Rows = nil
	threadPanel.Rows = nil

	boardPanel.SelectedRow = 0
	threadPanel.SelectedRow = 0

	for _, thread := range threads {
		boardPanel.Rows = append(boardPanel.Rows, thread.Title+" ["+thread.Date+"](fg:white)")
	}
}

func getMessagesToLoad() []string {
	if threadPanel.SelectedRow < len(activeThreads.Messages)-1 {
		return []string{
			// TODO maybe fetch more ahead
			activeThreads.Messages[threadPanel.SelectedRow].Link,
			activeThreads.Messages[threadPanel.SelectedRow+1].Link}
	}
	return []string{activeThreads.Messages[threadPanel.SelectedRow].Link}
}

func loadMessage() {
	if len(activeThreads.Messages) > 0 {
		start := time.Now()
		message = board.GetMessage(getMessagesToLoad())
		message.EnrichedContent, message.Links = util.EnrichLinks(message.Content)

		messagePanel.Rows = strings.Split("\n"+util.FormatQuote(message.EnrichedContent), "\n")

		messagePanel.ScrollTop()
		board.Logger.Printf("loading message took %s", time.Since(start))

		// TODO Copy these two commands into function
		activeThreads.Messages[threadPanel.SelectedRow].Read = true
		board.SetMessageAsRead(message.ID)
	}

	// Render thread for read messages
	renderThread()
}

// selectNextUnreadMessage selects the next unread message in the current thread
func selectNextUnreadMessage() {
	for i := threadPanel.SelectedRow + 1; i < len(activeThreads.Messages); i++ {
		if !activeThreads.Messages[i].Read {
			threadPanel.SelectedRow = i
			return
		}
	}
}

// answer uses the default system browser to open the answer link of the currently selected message
func answer() {
	open.Run(board.BoardURL + "pxmboard.php?mode=messageform&brdid=" + activeBoard.ID + "&msgid=" + message.ID)
}

// loadThread loads selected thread from board and displays the first message
func loadThread() {
	activeThreads = board.GetThread(threads[boardPanel.SelectedRow].ID, activeBoard.ID)
	threadPanel.SelectedRow = 0

	loadMessage()
	renderThread()

	messagePanel.ScrollTop()
}

func renderThread() {

	threadPanel.Rows = nil

	// Clear thread panel
	for _, m := range activeThreads.Messages {
		messageColor := "red"

		if m.Read {
			messageColor = "grey"
		}

		threadPanel.Rows = append(
			threadPanel.Rows,
			strings.Repeat("    ", m.Hierarchy-1)+
				"○ ["+m.Topic+"](fg:"+messageColor+") ["+m.Date+" "+m.Author.Name+"](fg:white)")
	}

}

// openLinks opens a link in the displayed message with the default system browser
func openLink(nr int) {
	link := message.Links[nr-1]
	cleanedLink := strings.Replace(link, "[", "", 1)
	cleanedLink = strings.Replace(cleanedLink, "]", "", 1)
	open.Run(cleanedLink)
}

func loadForum() {
	activeForum = board.GetForum()
	var boardNames []string

	for _, thread := range activeForum.Boards {
		boardNames = append(boardNames, thread.Title)
	}

	tabpane = widgets.NewTabPane(boardNames...)
	// tabpane.SetRect(0, 1, 50, 4)
	tabpane.Border = false
	tabpane.ActiveTabIndex = 0
}

func initialize() {
	// Initialize
	loadBoard()
	loadThread()
}

// colorize the ui depending on the active pane
func colorize() {
	inactiveColor := ui.ColorWhite
	activeColor := ui.ColorRed

	boardPanel.TextStyle = ui.NewStyle(activeColor)
	threadPanel.TextStyle = ui.NewStyle(activeColor)
	tabpane.ActiveTabStyle = ui.NewStyle(activeColor)

	boardPanel.BorderStyle = ui.NewStyle(inactiveColor)
	threadPanel.BorderStyle = ui.NewStyle(inactiveColor)
	tabpane.BorderStyle = ui.NewStyle(inactiveColor)
	messagePanel.BorderStyle = ui.NewStyle(inactiveColor)

	switch activePane {
	case 1:
		boardPanel.TextStyle = ui.NewStyle(activeColor)
		boardPanel.BorderStyle = ui.NewStyle(activeColor)
	case 2:
		threadPanel.TextStyle = ui.NewStyle(activeColor)
		threadPanel.BorderStyle = ui.NewStyle(activeColor)
	case 3:
		messagePanel.BorderStyle = ui.NewStyle(activeColor)

	}

}

func main() {

	if len(os.Args) > 1 {
		fmt.Printf(helpPage)
		os.Exit(0)
	}

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// Activate Board Pane first
	activePane = 1

	messagePanel = widgets.NewList()
	boardPanel = widgets.NewList()
	threadPanel = widgets.NewList()

	messagePanel.WrapText = true

	loadForum()
	initialize()

	boardPanel.WrapText = false
	colorize()

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
		start := time.Now()
		board.Logger.Printf("======== KEY PRESS '%s' =========", e.ID)
		switch e.ID {
		case "<Tab>":
			if activePane < maxPane {
				activePane++
			} else {
				activePane = 1
			}
			colorize()
		case "1", "2", "3", "4", "5", "6", "7", "8", "9", "0":
			linkNr, _ := strconv.Atoi(e.ID)
			openLink(linkNr)
		case "a":
			answer()
		case "q", "<C-c>":
			return
		case "?":
			messagePanel.Rows = strings.Split(util.FormatQuote(helpPage), "\n")
		case "b":
		case "<Left>":
			tabpane.FocusLeft()
			ui.Clear()
			renderTab()
			initialize()
		case "n":
		case "<Right>":
			tabpane.FocusRight()
			ui.Clear()
			renderTab()
			initialize()
		case "<Down>":
			switch activePane {
			case 1:
				boardPanel.ScrollDown()
				loadThread()
			case 2:
				threadPanel.ScrollDown()
				loadMessage()
			case 3:
				messagePanel.ScrollPageDown()
			}
		case "<Up>":
			switch activePane {
			case 1:
				boardPanel.ScrollUp()
				loadThread()
			case 2:
				threadPanel.ScrollUp()
				loadMessage()
			case 3:
				messagePanel.ScrollPageUp()
			}
		case "J", "z":
			boardPanel.ScrollDown()
			loadThread()
		case "K":
			boardPanel.ScrollUp()
			loadThread()
		case "j":
			threadPanel.ScrollDown()
			loadMessage()
		case "k":
			threadPanel.ScrollUp()
			loadMessage()
		case "u":
			selectNextUnreadMessage()
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
		board.Logger.Printf("UI interaction for %s freezed for %s", e.ID, time.Since(start))

	}
}
