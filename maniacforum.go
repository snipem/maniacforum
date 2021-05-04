package main

// run: make run

import (
	"flag"
	"fmt"
	"log"
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

// maniacforum ...
type maniacforum struct {
	active   maniacforumModel
	ui       uiContent
}

type maniacforumModel struct {
	threads board.Thread
	board   board.Board
	forum   *board.Forum
	message board.Message
	state   maniacforumState
}

type maniacforumState struct {
	activePane int
	maxPane    int
}

type uiContent struct {
	threadPanel  *widgets.List
	messagePanel *widgets.List
	boardPanel   *widgets.List
	tabpane      *widgets.TabPane
}

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
   e  - Die Nachricht im Standard-Browser öffnen
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

func (mf *maniacforum) loadBoard() {
	boardID := mf.active.forum.Boards[mf.ui.tabpane.ActiveTabIndex].ID
	mf.active.board = mf.active.forum.GetBoard(boardID)
	// threads = mf.active.board.Threads

	// Clear board panel
	mf.ui.boardPanel.Rows = nil
	mf.ui.messagePanel.Rows = nil
	mf.ui.threadPanel.Rows = nil

	mf.ui.boardPanel.SelectedRow = 0
	mf.ui.threadPanel.SelectedRow = 0

	for _, thread := range mf.active.board.Threads {

		threadPrefix := ""
		if thread.IsSticky {
			threadPrefix = "⋌ "
		}

		mf.ui.boardPanel.Rows = append(mf.ui.boardPanel.Rows, threadPrefix+thread.Title+" ["+thread.Date+"](fg:white)")
	}
}

func (mf *maniacforum) loadMessage() {
	if len(mf.active.threads.Messages) > 0 {
		start := time.Now()
		var err error
		mf.active.message, err = mf.active.forum.GetMessage(mf.active.threads.Messages[mf.ui.threadPanel.SelectedRow].Link)

		// Just print the error to the screen
		if err != nil {
			log.Print(err)
		}

		mf.active.message.EnrichedContent, mf.active.message.Links = util.EnrichContent(mf.active.message.Content, mf.ui.messagePanel.Inner.Dx())
		mf.ui.messagePanel.Rows = strings.Split(mf.active.message.EnrichedContent, "\n")
		mf.ui.messagePanel.ScrollTop()

		// TODO Copy these two commands into function
		mf.active.threads.Messages[mf.ui.threadPanel.SelectedRow].Read = true
		board.SetMessageAsRead(mf.active.message.ID)

		board.Logger.Printf("loading message %s by '%s' took %s", mf.active.message.ID, mf.active.message.Author.Name, time.Since(start))

		// Fully render ui before fetching messages for cache
		ui.Clear()

		fetchAheadMessages := 2

		// Get the next two messages for the cache, ignore them for now, but make them available for the cache
		if len(mf.active.threads.Messages) > mf.ui.threadPanel.SelectedRow+fetchAheadMessages {
			for i := 1; i <= fetchAheadMessages; i++ {
				// Go routine will run in background even if function finishes. The actual message is returned
				// and the content of the fetch ahead messages is stored into the cache
				go mf.active.forum.GetMessage(mf.active.threads.Messages[mf.ui.threadPanel.SelectedRow+i].Link)
			}
		}
	}

	// Render thread for read messages
	mf.renderThread()
}

// selectNextUnreadMessage selects the next unread message in the current thread
func (mf *maniacforum) selectNextUnreadMessage() {
	for i := mf.ui.threadPanel.SelectedRow + 1; i < len(mf.active.threads.Messages); i++ {
		if !mf.active.threads.Messages[i].Read {
			mf.ui.threadPanel.SelectedRow = i
			return
		}
	}
}

// answerMessage uses the default system browser to open the answerMessage link of the currently selected message
func (mf *maniacforum) answerMessage() {
	answerURL := mf.active.forum.URL + "pxmboard.php?mode=messageform&brdid=" + mf.active.board.ID + "&msgid=" + mf.active.message.ID
	err := open.Run(answerURL)
	if err != nil {
		log.Println(err)
	}
}

// openMessage uses the default system browser to open currently selected message
func (mf *maniacforum) openMessage() {
	answerURL := mf.active.forum.URL + "pxmboard.php?mode=message&brdid=" + mf.active.board.ID + "&msgid=" + mf.active.message.ID
	err := open.Run(answerURL)
	if err != nil {
		log.Println(err)
	}
}

// loadThread loads selected thread from board and displays the first message
func (mf *maniacforum) loadThread() {
	// FIXME this logic with Threads and threads seems illogic
	var err error
	mf.active.message, err = mf.active.forum.GetMessage(mf.active.board.Threads[mf.ui.boardPanel.SelectedRow].Link)

	// Just print the error to the screen
	if err != nil {
		log.Print(err)
	}

	mf.active.threads = mf.active.forum.GetThread(mf.active.board.Threads[mf.ui.boardPanel.SelectedRow].ID, mf.active.board.ID)
	mf.ui.threadPanel.SelectedRow = 0

	mf.renderThread()

	mf.ui.messagePanel.ScrollTop()

	fetchAheadThreads := 2

	// Get the next two messages for the cache, ignore them for now, but make them available for the cache
	if len(mf.active.board.Threads) >= mf.ui.boardPanel.SelectedRow+fetchAheadThreads {
		for i := 1; i <= fetchAheadThreads; i++ {
			// Go routine will run in background even if function finishes. The actual message is returned
			// and the content of the fetch ahead messages is stored into the cache
			go mf.active.forum.GetThread(mf.active.board.Threads[mf.ui.boardPanel.SelectedRow+i].ID, mf.active.board.ID)
		}
	}
}

func (mf *maniacforum) renderThread() {

	mf.ui.threadPanel.Rows = nil

	// Clear thread panel
	for _, m := range mf.active.threads.Messages {
		messageColor := "red"

		if m.Read {
			messageColor = "grey"
		}

		mf.ui.threadPanel.Rows = append(
			mf.ui.threadPanel.Rows,
			strings.Repeat("    ", m.Hierarchy-1)+
				"○ ["+m.Topic+"](fg:"+messageColor+") ["+m.Date+" "+m.Author.Name+"](fg:white)")
	}
	mf.active.message.EnrichedContent, mf.active.message.Links = util.EnrichContent(mf.active.message.Content, mf.ui.messagePanel.Inner.Dx())
	// TODO Workaround for termui not rendering the first line starting with a quote in red. Add a leading line
	mf.ui.messagePanel.Rows = strings.Split("\n"+mf.active.message.EnrichedContent, "\n")

}

// openLinks opens a link in the displayed message with the default system browser
func (mf *maniacforum) openLink(nr int) error {
	if nr > len(mf.active.message.Links) {
		return fmt.Errorf("no link with number %d in message", nr)
	}
	link := mf.active.message.Links[nr-1]
	cleanedLink := strings.Replace(link, "[", "", 1)
	cleanedLink = strings.Replace(cleanedLink, "]", "", 1)
	err := open.Run(cleanedLink)
	if err != nil {
		return err
	}

	return nil
}

func (mf *maniacforum) loadForum(forumUrl string, ignoreSSL bool) (err error) {
	mf.active.forum, err = board.GetForum(forumUrl, ignoreSSL)

	if err != nil {
		return err
	}

	var boardNames []string

	for _, thread := range mf.active.forum.Boards {
		boardNames = append(boardNames, thread.Title)
	}

	mf.ui.tabpane = widgets.NewTabPane(boardNames...)
	// mf.ui.tabpane.SetRect(0, 1, 50, 4)
	mf.ui.tabpane.Border = false
	mf.ui.tabpane.ActiveTabIndex = 0

	return nil
}

func (mf *maniacforum) initialize() {
	// Initialize
	mf.active.state.maxPane = 3
	mf.loadBoard()
	mf.loadThread()
}

// colorize the ui depending on the active pane
func (mf *maniacforum) colorize() {
	inactiveColor := ui.ColorWhite
	activeColor := ui.ColorRed

	mf.ui.boardPanel.TextStyle = ui.NewStyle(activeColor)
	mf.ui.threadPanel.TextStyle = ui.NewStyle(activeColor)
	mf.ui.tabpane.ActiveTabStyle = ui.NewStyle(activeColor)

	mf.ui.boardPanel.BorderStyle = ui.NewStyle(inactiveColor)
	mf.ui.threadPanel.BorderStyle = ui.NewStyle(inactiveColor)
	mf.ui.tabpane.BorderStyle = ui.NewStyle(inactiveColor)
	mf.ui.messagePanel.BorderStyle = ui.NewStyle(inactiveColor)

	switch mf.active.state.activePane {
	case 1:
		mf.ui.boardPanel.TextStyle = ui.NewStyle(activeColor)
		mf.ui.boardPanel.BorderStyle = ui.NewStyle(activeColor)
	case 2:
		mf.ui.threadPanel.TextStyle = ui.NewStyle(activeColor)
		mf.ui.threadPanel.BorderStyle = ui.NewStyle(activeColor)
	case 3:
		mf.ui.messagePanel.BorderStyle = ui.NewStyle(activeColor)

	}

}

func main() {
	err := run()
	if err != nil {
		log.Println(err)
	}
}

func run() error {

	ignoreSSL := flag.Bool("ignoreSSL", false, "Ignore SSL. Future legacy mode, if SSL handshake fails on old machines.")
	boardUrl := flag.String("url", board.DefaultBoardURL, "URL of Maniac Forum")

	flag.Parse()

	var mf maniacforum

	if err := ui.Init(); err != nil {
		return fmt.Errorf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// Activate Board Pane first
	mf.active.state.activePane = 1

	mf.ui.messagePanel = widgets.NewList()
	mf.ui.boardPanel = widgets.NewList()
	mf.ui.threadPanel = widgets.NewList()

	mf.ui.messagePanel.WrapText = false

	err := mf.loadForum(*boardUrl, *ignoreSSL)
	if err != nil {
		return err
	}

	mf.ui.boardPanel.WrapText = false
	mf.colorize()

	grid := ui.NewGrid()

	grid.Set(
		ui.NewRow(0.05, mf.ui.tabpane),
		ui.NewRow(0.95,
			ui.NewCol(1.0/2,
				ui.NewRow(0.5, mf.ui.boardPanel),
				ui.NewRow(0.5, mf.ui.threadPanel),
			),
			ui.NewCol(1.0/2, mf.ui.messagePanel),
		),
	)

	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	// UI has to be rendered to determine sizes for wrapping, this will
	// show an empty UI before the initialize function is called
	ui.Render(grid)
	mf.initialize()

	// Render initially
	ui.Render(mf.ui.boardPanel, mf.ui.messagePanel, mf.ui.threadPanel, mf.ui.tabpane)

	previousKey := ""
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "<Tab>":
			if mf.active.state.activePane < mf.active.state.maxPane {
				mf.active.state.activePane++
			} else {
				mf.active.state.activePane = 1
			}
			mf.colorize()
		case "Z": // <S-Tab> is rendered as <Escape>[Z, just check for [Z for now
			if previousKey == "[" {
				if mf.active.state.activePane > 1 {
					mf.active.state.activePane--
				} else {
					mf.active.state.activePane = mf.active.state.maxPane
				}
				mf.colorize()
			}
		case "1", "2", "3", "4", "5", "6", "7", "8", "9", "0":
			linkNr, _ := strconv.Atoi(e.ID)
			err := mf.openLink(linkNr)
			if err != nil {
				log.Println(err)
			}
		case "a":
			mf.answerMessage()
		case "e":
			mf.openMessage()
		case "q", "<C-c>":
			return nil
		case "r":
			board.ClearCache()
			ui.Clear()
			mf.initialize()
		case "?":
			enrichedHelp, helpLinks := util.EnrichContent(helpPage, mf.ui.messagePanel.Inner.Dx())
			mf.active.message.Links = helpLinks
			mf.ui.messagePanel.Rows = strings.Split(enrichedHelp, "\n")
		case "b":
		case "<Left>":
			mf.ui.tabpane.FocusLeft()
			ui.Clear()
			mf.initialize()
		case "n":
		case "<Right>":
			mf.ui.tabpane.FocusRight()
			ui.Clear()
			mf.initialize()
		case "<MouseWheelDown>":
			switch mf.active.state.activePane {
			case 1:
				mf.ui.boardPanel.ScrollDown()
			case 2:
				mf.ui.threadPanel.ScrollDown()
			case 3:
				mf.ui.messagePanel.ScrollPageDown()
			}
		case "<Down>":
			switch mf.active.state.activePane {
			case 1:
				mf.ui.boardPanel.ScrollDown()
				mf.loadThread()
			case 2:
				mf.ui.threadPanel.ScrollDown()
				mf.loadMessage()
			case 3:
				mf.ui.messagePanel.ScrollPageDown()
			}
		case "<MouseWheelUp>":
			switch mf.active.state.activePane {
			case 1:
				mf.ui.boardPanel.ScrollUp()
			case 2:
				mf.ui.threadPanel.ScrollUp()
			case 3:
				mf.ui.messagePanel.ScrollPageUp()
			}
		case "<Up>":
			switch mf.active.state.activePane {
			case 1:
				mf.ui.boardPanel.ScrollUp()
				mf.loadThread()
			case 2:
				mf.ui.threadPanel.ScrollUp()
				mf.loadMessage()
			case 3:
				mf.ui.messagePanel.ScrollPageUp()
			}
		case "J", "z":
			mf.ui.boardPanel.ScrollDown()
			mf.loadThread()
		case "K":
			mf.ui.boardPanel.ScrollUp()
			mf.loadThread()
		case "j":
			mf.ui.threadPanel.ScrollDown()
			mf.loadMessage()
		case "k":
			mf.ui.threadPanel.ScrollUp()
			mf.loadMessage()
		case "u":
			mf.selectNextUnreadMessage()
			mf.loadMessage()
		case "<Enter>":
			mf.loadThread()
		case "<C-d>":
			mf.ui.boardPanel.ScrollHalfPageDown()
		case "<C-u>":
			mf.ui.boardPanel.ScrollHalfPageUp()
		case "<C-f>":
			mf.ui.boardPanel.ScrollPageDown()
		case "<C-b>":
			mf.ui.boardPanel.ScrollPageUp()
		case "g":
			if previousKey == "g" {
				mf.ui.boardPanel.ScrollTop()
			}
		case "<Home>":
			mf.ui.boardPanel.ScrollTop()
		case "G", "<End>":
			mf.ui.boardPanel.ScrollBottom()
		case "<Resize>":
			termWidth, termHeight := ui.TerminalDimensions()
			grid.SetRect(0, 0, termWidth, termHeight)
			ui.Render(grid)
		case "<MouseLeft>":

			if util.HandleMouseClickEventOnTabBar(e, mf.ui.tabpane) {
				mf.loadBoard()
				mf.active.state.activePane = 0
				ui.Clear()
				mf.initialize()
			} else if util.HandleMouseClickEventOnList(e, mf.ui.boardPanel) {
				mf.loadThread()
				mf.active.state.activePane = 1
			} else if util.HandleMouseClickEventOnList(e, mf.ui.threadPanel) {
				mf.loadMessage()
				mf.active.state.activePane = 2
			} else if util.HandleMouseClickEventOnList(e, mf.ui.messagePanel) {
				mf.active.state.activePane = 3
			}

			mf.colorize()

		}

		if previousKey == "g" {
			previousKey = ""
		} else {
			previousKey = e.ID
		}

		ui.Render(mf.ui.boardPanel, mf.ui.messagePanel, mf.ui.threadPanel, mf.ui.tabpane)
	}
	return nil
}
