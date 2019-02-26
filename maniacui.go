package main

import (
	"log"

	"github.com/snipem/maniacforum/board"
	"strings"

	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
)

var innerThreads board.Thread

func main() {

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	t := widgets.NewParagraph()
	t.Title = "Message"

	l := widgets.NewList()
	l.Title = "Themen"
	
	beitraege := widgets.NewList()
	beitraege.Title = "Beiträge"

	threads := board.GetThreads("pxmboard.php?mode=threadlist&brdid=1&sortorder=last")

	for _, thread := range threads {
		l.Rows = append(l.Rows, thread.Title)
	}

	l.TextStyle = ui.NewStyle(ui.ColorYellow)
	beitraege.TextStyle = ui.NewStyle(ui.ColorYellow)
	l.WrapText = false
	// l.SetRect(0, 0, 30, 50)
	// t.SetRect(30, 20, 100, 40)

	grid := ui.NewGrid()

	grid.Set(
		ui.NewCol(1.0/2,
			ui.NewRow(1.0/2, l),
			ui.NewRow(1.0/2, beitraege),
		),
		ui.NewCol(1.0/2, t),
	)

	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	// ui.Render(l)
	// ui.Render(t)

	ui.Render(grid)

	previousKey := ""
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		case "<Down>":
			l.ScrollDown()
		case "<Up>":
			l.ScrollUp()
		case "j":
			beitraege.ScrollDown()
			message := board.GetMessage(innerThreads.Messages[beitraege.SelectedRow].Link)
			t.Text = message.Content
		case "k":
			beitraege.ScrollUp()
		case "<Enter>":
			// log.Fatalf(threads[l.SelectedRow].Link)
			message := board.GetMessage(threads[l.SelectedRow].Link)
			// t.Text = threads[l.SelectedRow].Link
			innerThreads = board.GetThread(threads[l.SelectedRow].Id)
			beitraege.Rows = nil
			for _, message := range innerThreads.Messages {
				beitraege.Rows = append(beitraege.Rows, strings.Repeat("    ", message.Hiearachy-1)+ "○ " + message.Topic)
			}
			t.Text = message.Content
			// t.Text = "test"
			ui.Render(t)
			ui.Render(beitraege)
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

	}
}
