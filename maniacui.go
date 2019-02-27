package main

import (
	"log"

	"strings"

	"github.com/snipem/maniacforum/board"
	"github.com/snipem/maniacforum/util"

	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
)

var innerThreads board.Thread
var beitraege *widgets.List
var t *widgets.Paragraph
var l *widgets.List
var threads []board.Thread

func loadBeitrag()  {
	if len(innerThreads.Messages) > 0 {
		message := board.GetMessage(innerThreads.Messages[beitraege.SelectedRow].Link)
		t.Text = util.FormatQuote(message.Content)
	}
}

func loadThread()  {
			// log.Fatalf(threads[l.SelectedRow].Link)
			message := board.GetMessage(threads[l.SelectedRow].Link)
			// t.Text = threads[l.SelectedRow].Link
			innerThreads = board.GetThread(threads[l.SelectedRow].Id)
			beitraege.Rows = nil
			beitraege.SelectedRow = 0
			for _, message := range innerThreads.Messages {
				beitraege.Rows = append(beitraege.Rows, strings.Repeat("    ", message.Hiearachy-1)+"○ "+message.Topic+" "+message.Date+" "+message.Author.Name)
			}
			t.Text = message.Content
			// t.Text = "test"
	
}

func main() {

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	t = widgets.NewParagraph()

	l = widgets.NewList()
	// TODO get from page
	l.Title = "Smalltalk"

	beitraege = widgets.NewList()

	threads = board.GetThreads("pxmboard.php?mode=threadlist&brdid=1&sortorder=last")

	for _, thread := range threads {
		l.Rows = append(l.Rows, thread.Title)
	}

	l.TextStyle = ui.NewStyle(ui.ColorRed)
	beitraege.TextStyle = ui.NewStyle(ui.ColorRed)
	l.WrapText = false
	// l.SetRect(0, 0, 30, 50)
	// t.SetRect(30, 20, 100, 40)

	tabpane := widgets.NewTabPane(["S", "OT"])
	tabpane.SetRect(0, 1, 50, 4)
	tabpane.Border = false


	grid := ui.NewGrid()

	grid.Set(
		ui.NewCol(1.0/2,
			ui.NewRow(0.1, tabpane),
			ui.NewRow(0.4, l),
			ui.NewRow(0.5, beitraege),
		),
		ui.NewCol(1.0/2, t),
	)

	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	// ui.Render(l)
	// ui.Render(t)

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
		case "q", "<C-c>":
			return
		case "b":
			tabpane.FocusLeft()
			ui.Clear()
			renderTab()
		case "n":
			tabpane.FocusRight()
			ui.Clear()
			renderTab()
		case "J","<Down>":
			l.ScrollDown()
			loadThread()
		case "K","<Up>":
			l.ScrollUp()
			loadThread()
		case "j":
			beitraege.ScrollDown()
			loadBeitrag()
		case "k":
			beitraege.ScrollUp()
			loadBeitrag()
		case "<Enter>":
			loadThread()
		case "<C-d>":
			l.HalfPageDown()
		case "<C-u>":
			l.HalfPageUp()
		case "<C-f>":
			l.PageDown()
		case "<C-b>":
			l.PageUp()
		case "g":
			if previousKey == "g" {
				l.ScrollTop()
			}
		case "<Home>":
			l.ScrollTop()
		case "G", "<End>":
			l.ScrollBottom()
		}

		if previousKey == "g" {
			previousKey = ""
		} else {
			previousKey = e.ID
		}

		ui.Render(l)
		ui.Render(t)
		ui.Render(beitraege)

		ui.Render(tabpane)

		renderTab()

	}
}
