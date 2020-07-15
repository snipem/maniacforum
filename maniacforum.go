package main

// run: make run

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

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
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

var version = "dev"

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

[https://github.com/snipem/maniacforum]
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

		threadPrefix := ""
		if thread.IsSticky {
			threadPrefix = "⋌ "
		}

		boardPanel.Rows = append(boardPanel.Rows, threadPrefix+thread.Title+" ["+thread.Date+"](fg:white)")
	}
}

func loadMessage() {
	if len(activeThreads.Messages) > 0 {
		start := time.Now()
		message = board.GetMessage(activeThreads.Messages[threadPanel.SelectedRow].Link)

		message.EnrichedContent, message.Links = util.EnrichContent(message.Content, messagePanel.Inner.Dx())
		messagePanel.Rows = strings.Split(message.EnrichedContent, "\n")
		messagePanel.ScrollTop()

		// TODO Copy these two commands into function
		activeThreads.Messages[threadPanel.SelectedRow].Read = true
		board.SetMessageAsRead(message.ID)

		board.Logger.Printf("loading message %s by '%s' took %s", message.ID, message.Author.Name, time.Since(start))

		// Fully render ui before fetching messages for cache
		ui.Clear()

		fetchAheadMessages := 2

		// Get the next two messages for the cache, ignore them for now, but make them available for the cache
		if len(activeThreads.Messages) >= threadPanel.SelectedRow+fetchAheadMessages {
			for i := 1; i <= fetchAheadMessages; i++ {
				// Go routine will run in background even if function finishes. The actual message is returned
				// and the content of the fetch ahead messages is stored into the cache
				go board.GetMessage(activeThreads.Messages[threadPanel.SelectedRow+i].Link)
			}
		}
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
	err := open.Run(board.BoardURL + "pxmboard.php?mode=messageform&brdid=" + activeBoard.ID + "&msgid=" + message.ID)
	log.Fatal(err)
}

// loadThread loads selected thread from board and displays the first message
func loadThread() {
	message = board.GetMessage(threads[boardPanel.SelectedRow].Link)
	activeThreads = board.GetThread(threads[boardPanel.SelectedRow].ID, activeBoard.ID)
	threadPanel.SelectedRow = 0

	renderThread()

	messagePanel.ScrollTop()

	fetchAheadThreads := 2

	// Get the next two messages for the cache, ignore them for now, but make them available for the cache
	if len(activeBoard.Threads) >= boardPanel.SelectedRow+fetchAheadThreads {
		for i := 1; i <= fetchAheadThreads; i++ {
			// Go routine will run in background even if function finishes. The actual message is returned
			// and the content of the fetch ahead messages is stored into the cache
			go board.GetThread(activeBoard.Threads[boardPanel.SelectedRow+i].ID, activeBoard.ID)
		}
	}
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
	message.EnrichedContent, message.Links = util.EnrichContent(message.Content, messagePanel.Inner.Dx())
	// TODO Workaround for termui not rendering the first line starting with a quote in red. Add a leading line
	messagePanel.Rows = strings.Split("\n"+message.EnrichedContent, "\n")

}

// openLinks opens a link in the displayed message with the default system browser
func openLink(nr int) error {
	if nr > len(message.Links) {
		return fmt.Errorf("No link with number %d in message", nr)
	}
	link := message.Links[nr-1]
	cleanedLink := strings.Replace(link, "[", "", 1)
	cleanedLink = strings.Replace(cleanedLink, "]", "", 1)
	err := open.Run(cleanedLink)
	if err != nil {
		return err
	}

	return nil
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
	run()
}

func run() {

	if len(os.Args) > 1 {
		fmt.Print(helpPage)
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

	messagePanel.WrapText = false

	loadForum()

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

	// UI has to be rendered to determine sizes for wrapping, this will
	// show an empty UI before the initialize function is called
	ui.Render(grid)
	initialize()

	// Render initially
	ui.Render(boardPanel, messagePanel, threadPanel, tabpane)

	previousKey := ""
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
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
			enrichedHelp, helpLinks := util.EnrichContent(helpPage, messagePanel.Inner.Dx())
			message.Links = helpLinks
			messagePanel.Rows = strings.Split(enrichedHelp, "\n")
		case "b":
		case "<Left>":
			tabpane.FocusLeft()
			ui.Clear()
			initialize()
		case "n":
		case "<Right>":
			tabpane.FocusRight()
			ui.Clear()
			initialize()
		case "<MouseWheelDown>":
			messagePanel.ScrollPageDown()
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
		case "<MouseWheelUp>":
			messagePanel.ScrollPageUp()
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
		case "<Resize>":
			termWidth, termHeight := ui.TerminalDimensions()
			grid.SetRect(0, 0, termWidth, termHeight)
			ui.Clear()
		case "<MouseLeft>":

			if handleMouseClickEventOnTabBar(e, tabpane) {
				loadBoard()
				activePane = 0
				ui.Clear()
				initialize()
			} else if handleMouseClickEventOnList(e, boardPanel) {
				loadThread()
				activePane = 1
			} else if handleMouseClickEventOnList(e, threadPanel) {
				loadMessage()
				activePane = 2
			}
			colorize()

		}

		if previousKey == "g" {
			previousKey = ""
		} else {
			previousKey = e.ID
		}

		ui.Render(boardPanel, messagePanel, threadPanel, tabpane)
	}
}

func handleMouseClickEventOnTabBar(e ui.Event, bar *widgets.TabPane) bool {
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

func handleMouseClickEventOnList(e ui.Event, list *widgets.List) bool {

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
