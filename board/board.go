package board

import (
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

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
	Content   string
	Link      string
	Topic     string
	Date      string
	Hierarchy int
	Author    User
}

type Board struct {
	Threads []Thread
}

// User contains User data
type User struct {
	Name string
	ID   int
}

// GetThread fatches a Thread based on a Thread id
func GetThread(id string) Thread {
	resource := "pxmboard.php?mode=thread&brdid=1&thrdid=" + id
	var t Thread
	doc := getDoc(resource)

	doc.Find("li").Each(func(i int, s *goquery.Selection) {
		var m Message
		m.Topic = s.Find("a > font").Text()
		m.Hierarchy = s.ParentsFiltered("ul").Length()
		m.Link, _ = s.Find("a").Attr("href")
		m.Author.Name = strings.TrimSpace(s.Find("span").Find("span").Text())
		// m.Date = s.Find("span > font").Text()

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
	doc := getDoc(resource)

	doc.Find(".bg2 > td > font").Each(func(i int, s *goquery.Selection) {
		m.Content = s.Text()
	})

	doc.Find("tbody > tr.bg1 > td#norm > a").Each(func(i int, s *goquery.Selection) {
		m.Author.Name = s.Text()
	})
	return m
}

var boardURL = "https://www.maniac-forum.de/forum/"

func getDoc(resource string) *goquery.Document {
	// Request the HTML page.
	res, err := http.Get(boardURL + resource)
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

func GetBoard(resource string) Board {

	var board Board

	doc := getDoc(resource)

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

		board.Threads[i] = t
	})

	return board
}
