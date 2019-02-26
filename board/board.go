package board

import (
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type Thread struct {
	Title          string
	Link           string
	Author         string
	Date           string
	Answers        int
	LastAnswerDate string
	LastAnswerLink string
}

type Message struct {
	Content string
	Link    string
	Topic   string
	Date    string
	Author  User
}

type User struct {
	Name string
	Id   int
}

func GetMessage(resource string) Message {
	var m Message
	doc := getDoc(resource)

	doc.Find(".bg2").Each(func(i int, s *goquery.Selection) {
		m.Content = s.Find("font").Text()
	})

	doc.Find("table").Each(func(i int, s *goquery.Selection) {
		m.Content = s.Find("font").Text()
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

func GetThreads() []Thread {

	var threads []Thread

	doc := getDoc("pxmboard.php?mode=threadlist&brdid=1&sortorder=last")

	// Find the review items
	doc.Find("#threadlist > a").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		var t Thread
		t.Title = s.Find("font").Text()
		t.Link, _ = s.Attr("href")
		// fmt.Printf("Thread %d: %s - %s\n", i, t.title, t.link)

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
