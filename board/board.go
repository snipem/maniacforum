package board

import (
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Thread struct {
	Id string
	Title          string
	Link           string
	Author         string
	Date           string
	Answers        int
	LastAnswerDate string
	LastAnswerLink string
	Messages []Message
}

type Message struct {
	Content string
	Link    string
	Topic   string
	Date    string
	Hiearachy int
	Author  User
}

type User struct {
	Name string
	Id   int
}

func GetThread(id string) Thread {
	resource := "pxmboard.php?mode=thread&brdid=1&thrdid="+id
	var t Thread
	doc := getDoc(resource)	

	doc.Find("li").Each(func(i int, s *goquery.Selection) {
		var m Message
		m.Topic = s.Find("a > font").Text()
		m.Hiearachy = s.ParentsFiltered("ul").Length()
		m.Link, _ = s.Find("a").Attr("href")
		m.Author.Name = strings.TrimSpace(s.Find("span").Find("span").Text())
		// m.Date = s.Find("span > font").Text()

		t.Messages = append(t.Messages, m)
	})

	return t
}

func GetMessage(resource string) Message {

	if resource == "" {
		log.Fatalf("Resource id is empty")
	}

	var m Message
	doc := getDoc(resource)

	doc.Find(".bg2 > td > font").Each(func(i int, s *goquery.Selection) {
		m.Content = s.Text()
	})

	doc.Find("table").Each(func(i int, s *goquery.Selection) {
		// m.Topic = s.Find("font").Text()
	})
	return m
}

var boardUrl = "https://www.maniac-forum.de/forum/"

func getDoc(resource string) *goquery.Document {
	// Request the HTML page.
	res, err := http.Get(boardUrl + resource)
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

func GetThreads(resource string) []Thread {

	var threads []Thread

	doc := getDoc(resource)

	// Find the review items
	doc.Find("#threadlist > a").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		var t Thread
		t.Title = s.Find("font").Text()
		t.Link, _ = s.Attr("href")

		id, _ := s.Attr("onclick")

		id = strings.Replace(id, "ld(", "", 1)
		t.Id = strings.Replace(id, ",0)", "", 1)
		// t.BoardId = "TODO"

		threads = append(threads, t)
	})

	doc.Find("#threadlist > font").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		var t = threads[i]
		t.Author = s.Find("span").Text()
		// fmt.Printf("Thread %d: %s - %s - %s\n", i, t.Title, t.Link, t.Author)

		threads[i] = t
	})

	return threads
}
