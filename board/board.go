package board

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/snipem/maniacforum/util"

	"net/url"

	"github.com/PuerkitoBio/goquery"
)

// DefaultBoardURL is the default base url of the forum
var DefaultBoardURL = "https://www.maniac-forum.de/forum/"

// Forum represents the whole forum
type Forum struct {
	Boards    []Board
	URL       string
	ignoreSSL bool
}

// Thread contains information about a Maniac Forum Thread
type Thread struct {
	ID             string
	Title          string
	Link           string
	Author         string
	Date           string
	Answers        int
	IsSticky       bool
	LastAnswerDate string
	LastAnswerLink string
	Messages       []Message
	Board          *Board
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
	Read            bool
	Thread          *Thread
	Board           *Board
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

// Logger is a logger for the board
var Logger *log.Logger
var readLogfile string
var c *cache.Cache

var useCache = true

func init() {
	c = cache.New(5*time.Minute, 10*time.Minute)
	readLogfile = getReadLogFilePath()

	logfile := "/dev/null"
	if _, ok := os.LookupEnv("MANIACFORUM_DEBUG"); ok {
		logfile = "maniacforum.log"
	}

	f, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	// TODO how to safely handle file? If activated log is not written
	// defer f.Close()

	Logger = log.New(f, "board.go ", log.LstdFlags)
}

func ClearCache() {
	c.Flush()
}

// getReadLogFilePath from env var or default .config file path
func getReadLogFilePath() string {
	var path string
	if env, ok := os.LookupEnv("MANIACFORUM_READLOG_FILE"); ok {
		path = env
	} else {
		usr, _ := user.Current()
		path = usr.HomeDir + "/.maniacread.log"
	}

	// Create file if not existing
	_, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}

	return path
}

// GetThread fetches a Thread based on a Thread id
func (f *Forum) GetThread(threadID string, boardID string) Thread {
	resource := "pxmboard.php?mode=thread&brdid=" + boardID + "&thrdid=" + threadID
	var t Thread
	doc, _ := f.getDoc(resource)

	doc.Find("li").Each(func(i int, s *goquery.Selection) {
		var m Message
		m.Topic = s.Find("a > font").Text()

		m.Hierarchy = s.ParentsFiltered("ul").Length()
		m.Link, _ = s.Find("a").Attr("href")
		m.Author.Name = strings.TrimSpace(s.Find("span").Find("span").Text())

		name, _ := s.Find("a").Attr("name")

		m.ID = cleanMessageID(name)
		m.Read = IsMessageRead(m.ID)

		// Remove sub element from doc that is included in date
		s.Find("li > span > font > b").Remove()
		foundDate := s.Find("li > span > font").Text()
		m.Date = strings.Replace(foundDate, " - ", "", 1)

		t.Messages = append(t.Messages, m)
	})

	return t
}

func cleanMessageID(dirty string) string {
	// Drop leading P from name
	return strings.Replace(dirty, "p", "", 1)
}

// GetMessage fetches a message based on it's resource string
func (f *Forum) GetMessage(resource string) (Message, error) {

	if resource == "" {
		return Message{}, fmt.Errorf("resource id is empty")
	}

	var m Message
	m.Link = resource
	values, _ := url.ParseQuery(resource)
	m.ID = cleanMessageID(values.Get("msgid"))

	doc, err := f.getDoc(resource)
	if err != nil {
		return Message{}, err
	}

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

	return m, nil
}

// SetMessageAsRead sets a message as read
func SetMessageAsRead(id string) {

	if IsMessageRead(id) {
		return
	}

	f, err := os.OpenFile(readLogfile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(id + "\n"); err != nil {
		panic(err)
	}
}

// IsMessageRead checks if a message has been read
func IsMessageRead(id string) bool {

	if strings.Compare(id, "") == 0 {
		return false
	}

	b, err := ioutil.ReadFile(readLogfile)
	if err != nil {
		panic(err)
	}
	s := string(b)

	isRead := strings.Contains(s, id)
	return isRead
}

// getDoc fetches a resource of the board directly or via cache if `useCache` is true
func (f *Forum) getDoc(resource string) (document *goquery.Document, err error) {
	// Request the HTML page.
	forumURL := util.JoinURL(f.URL, resource)

	var body string

	// Check if resource is already in cache
	cachedBody, foundInCache := c.Get(forumURL)
	if useCache && foundInCache { // Use cached resource if cache is used
		body = cachedBody.(string)
	} else { // Fetch resource if not in cache
		body, err = f.httpGet(forumURL)
		if err != nil {
			return nil, err
		}
		c.Set(forumURL, body, cache.DefaultExpiration)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	return doc, err

}

// httpGet fetches the content of a url and returns the body of the response
func (f *Forum) httpGet(url string) (string, error) {

	client := &http.Client{}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: f.ignoreSSL}

	Logger.Printf("Fetching %s", strings.Replace(url, f.URL, "", 1))

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "maniacforum-cli")
	res, err := client.Do(req)

	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("status for request %s code is not 200: %s", res.Request.URL, res.Status)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	return buf.String(), nil
}

// httpPost posts the data to a url and returns the body of the response
func (f *Forum) httpPost(url string, data url.Values) (string, error) {

	client := &http.Client{}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: f.ignoreSSL}

	Logger.Printf("Fetching %s", strings.Replace(url, f.URL, "", 1))

	// TODO Use User Agent
	// req, err := http.NewRequest("POST", url, nil)
	// req.Header.Add("User-Agent", "maniacforum-cli")
	// req.PostForm = data
	res, err := client.PostForm(url, data)

	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("status code is not 200: %d %s", res.StatusCode, res.Status)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	return buf.String(), nil
}

// searchMessages returns the list of matching messages from the new search of the forum
// boardID = -1 will search every forum as for the documentation of the service
func (f *Forum) searchMessages(query string, authorName string, boardID string, searchInBody bool, searchInTopic bool) ([]Message, error) {

	cbxBody := "0"
	cbxSubject := "0"

	if searchInBody {
		cbxBody = "1"
	}

	if searchInTopic {
		cbxSubject = "1"
	}

	body, _ := f.httpPost(util.JoinURL(f.URL,"search/search.php"), url.Values{
		"phrase":     {query},
		"autor":      {authorName},
		"board":      {boardID},
		"cbxBody":    {cbxBody},
		"cbxSubject": {cbxSubject},
		"suche":      {"durchsuchen"},
	})
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	var messages []Message

	// First run for getting topic name and link
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		var m Message

		m.Topic = s.Text()
		var exists bool
		m.Link, exists = s.Attr("href")
		if !exists {
			log.Print("Can't extract links from search")
		}

		m.Board = &Board{}
		m.Thread = &Thread{}

		m.Board.ID, m.Thread.ID, m.ID, err = util.ExtractIDsFromLink(m.Link)
		if err != nil {
			log.Fatal("Cannot extraect IDs from link")
		}

		messages = append(messages, m)
	})

	// Second run for getting the non HTML encapsulated author names and dates
	splittedResults := strings.Split(body, "Matches:")
	if len(splittedResults) != 2 {
		return messages, nil
	}
	matches := strings.Split(splittedResults[1], "<br>")

	// Results start at second <br> ignore first
	matches = matches[1:]

	re := regexp.MustCompile("von: (.*) , (.*)")
	for i := 0; i < len(matches); i++ {

		match := re.FindStringSubmatch(matches[i])
		if len(match) == 3 {
			messages[i].Author.Name = match[1]
			messages[i].Date = match[2]
		}

	}

	return messages, nil
}

// GetForum returns the forum
func GetForum(forumUrl string, ignoreSSL bool) (*Forum, error) {

	f := &Forum{
		URL:       forumUrl,
		ignoreSSL: ignoreSSL,
	}

	mainPage, err := f.getDoc("pxmboard.php")
	if err != nil {
		return nil, err
	}
	var boards []Board

	mainPage.Find("#norm > a").Each(func(index int, item *goquery.Selection) {
		href, _ := item.Attr("href")

		hrefURL, err := url.Parse(href)
		if err != nil {
			return
		}

		boardID := hrefURL.Query().Get("brdid")

		if boardID != "" {

			boards = append(boards, Board{
				Title: item.Text(),
				ID:    boardID,
			})

		}
	})

	f.Boards = boards

	return f, nil

}

// GetBoard fetches a Board like Smalltalk and the list of threads
func (f *Forum) GetBoard(boardID string) Board {

	var board Board

	resource := "pxmboard.php?mode=threadlist&brdid=" + boardID + "&sortorder=last"
	doc, _ := f.getDoc(resource)

	board.Title = doc.Find(".currentBoard > span").Text()
	board.ID = boardID

	doc.Find("#threadlist > a").Each(func(i int, s *goquery.Selection) {
		var t Thread
		t.Title = s.Find("font").Text()
		t.Link, _ = s.Attr("href")

		id, _ := s.Attr("onclick")

		id = strings.Replace(id, "ld(", "", 1)
		t.ID = strings.Replace(id, ",0)", "", 1)

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

	doc.Find("#threadlist > img").Each(func(i int, s *goquery.Selection) {

		var t = board.Threads[i]

		imageSrc, _ := s.Attr("src")

		if strings.Contains(imageSrc, "fixed.gif") {
			t.IsSticky = true
		} else {
			t.IsSticky = false
		}

		board.Threads[i] = t
	})

	return board
}
