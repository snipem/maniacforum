package board

import (
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Forum represents the whole forum
type Forum struct {
	Boards []Board
}

// Thread contains information about a Maniac Forum Thread
type Thread struct {
	ID             string
	Title          string
	Link           string
	Author         string
	Date           string
	Answers        int
	LastAnswerDate string
	LastAnswerLink string
	Messages       []Message
}

// Message contains information about a Maniac Forum Message. Single response to a Thread.
type Message struct {
	ID              string
	Content         string
	Link            string
	Topic           string
	Date            string
	EnrichedContent string
	Links           []string
	Hierarchy       int
	Author          User
}

// Board in forum, like Smalltalk, O/T, etc.
type Board struct {
	ID      string
	Threads []Thread
	Title   string
}

// User contains User data
type User struct {
	Name string
	ID   int
}

// GetThread fatches a Thread based on a Thread id
func GetThread(threadID string, boardID string) Thread {
	resource := "pxmboard.php?mode=thread&brdid=" + boardID + "&thrdid=" + threadID
	var t Thread
	doc := getDoc(resource)

	doc.Find("li").Each(func(i int, s *goquery.Selection) {
		var m Message
		m.Topic = s.Find("a > font").Text()

		m.Hierarchy = s.ParentsFiltered("ul").Length()
		m.Link, _ = s.Find("a").Attr("href")
		m.Author.Name = strings.TrimSpace(s.Find("span").Find("span").Text())

		// Remove sub element from doc that is included in date
		s.Find("li > span > font > b").Remove()
		foundDate := s.Find("li > span > font").Text()
		m.Date = strings.Replace(foundDate, " - ", "", 1)

		t.Messages = append(t.Messages, m)
	})

	return t
}

// GetMessage fetches a message based on it's resource string
func GetMessage(resource string) Message {

	if resource == "" {
		log.Fatalf("Resource id is empty")
	}

	var m Message
	m.Link = resource
	values, _ := url.ParseQuery(resource)
	m.ID = values.Get("msgid")

	doc := getDoc(resource)

	doc.Find(".bg2 > td > font").Each(func(i int, s *goquery.Selection) {
		m.Content = s.Text()
	})

	doc.Find("table > tbody > tr > td > table > tbody > tr > td > b").Each(func(i int, s *goquery.Selection) {
		m.Topic = s.Text()
	})

	doc.Find("table > tbody > tr:nth-child(2) > td#norm > a").Each(func(i int, s *goquery.Selection) {
		// Extract user id from link in username
		m.Author.Name = s.Text()
		href, _ := s.Attr("href")
		var re = regexp.MustCompile(".*usrid=")
		out := re.ReplaceAllString(href, "")
		m.Author.ID, _ = strconv.Atoi(out)
	})

	return m
}

var BoardURL = "https://www.maniac-forum.de/forum/"

func getDoc(resource string) *goquery.Document {
	// Request the HTML page.
	res, err := http.Get(BoardURL + resource)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return doc

}

// GetForum retuns the whole forum
func GetForum() Forum {
	// TODO Scrape this
	return Forum{
		[]Board{
			Board{
				ID:    "1",
				Title: "Smalltalk",
			},
			Board{
				ID:    "2",
				Title: "For Sale",
			},
			Board{
				ID:    "4",
				Title: "Tech'n'Cheats",
			},
			Board{
				ID:    "6",
				Title: "OT",
			},
			Board{
				ID:    "26",
				Title: "Filme & Serien",
			},
			Board{
				ID:    "8",
				Title: "Online Gaming",
			},
		},
	}
}

func GetBoard(boardID string) Board {

	var board Board

	resource := "pxmboard.php?mode=threadlist&brdid=" + boardID + "&sortorder=last"
	doc := getDoc(resource)

	// TODO Get from actual board
	board.Title = "Smalltalk"
	board.ID = boardID

	// Find the review items
	doc.Find("#threadlist > a").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		var t Thread
		t.Title = s.Find("font").Text()
		t.Link, _ = s.Attr("href")

		id, _ := s.Attr("onclick")

		id = strings.Replace(id, "ld(", "", 1)
		t.ID = strings.Replace(id, ",0)", "", 1)
		// t.BoardId = "TODO"

		board.Threads = append(board.Threads, t)
	})

	doc.Find("#threadlist > font").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		var t = board.Threads[i]
		t.Author = s.Find("span").Text()
		// fmt.Printf("Thread %d: %s - %s - %s\n", i, t.Title, t.Link, t.Author)

		// Remove sub element from doc that is included in date
		s.Find("b").Remove()
		foundDate := s.Text()
		foundDate = strings.Replace(foundDate, "\n", " ", -1)
		t.Date = strings.Replace(foundDate, " am ", "", 1)

		board.Threads[i] = t
	})

	return board
}
