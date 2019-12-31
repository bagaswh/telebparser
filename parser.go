package telebparser

import (
	"io"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type MessageRoom struct {
	RoomName string
	Messages []Message
}

type Message struct {
	ID        string
	DateSent  time.Time
	FromName  string
	ReplyToID string
	Content   string
}

var dateRe = regexp.MustCompile("(\\d{2})\\.(\\d{2})\\.(\\d{4}) (\\d{2}):(\\d{2}):(\\d{2})")

func getTimeComponents(dateString string) (day, month, year, hour, minute, second int) {
	dateComponents := dateRe.FindSubmatch([]byte(dateString))
	day, _ = strconv.Atoi(string(dateComponents[1]))
	month, _ = strconv.Atoi(string(dateComponents[2]))
	year, _ = strconv.Atoi(string(dateComponents[3]))
	hour, _ = strconv.Atoi(string(dateComponents[4]))
	minute, _ = strconv.Atoi(string(dateComponents[5]))
	second, _ = strconv.Atoi(string(dateComponents[6]))

	return
}

func getTimeValue(day, month, year, hour, minute, second int, locString string) (time.Time, error) {
	timeLocation, err := time.LoadLocation(locString)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, timeLocation), nil
}

func parseTime(dateString string, locString string) (time.Time, error) {
	day, month, year, hour, minute, second := getTimeComponents(dateString)
	timeValue, err := getTimeValue(day, month, year, hour, minute, second, locString)
	if err != nil {
		return time.Time{}, err
	}
	return timeValue, nil
}

func parseMessage(s *goquery.Selection, prevFromName *string) Message {
	if !s.HasClass("default") {
		return Message{}
	}

	var fromName string

	ID, _ := s.Attr("id")

	body := s.Find(".body")
	if s.HasClass("joined") {
		fromName = *prevFromName
	} else {
		fromName = body.Find(".from_name").Text()
		*prevFromName = fromName
	}

	dateString, _ := body.Find(".date").Attr("title")
	dateSent, err := parseTime(dateString, "Asia/Jakarta")
	if err != nil {
		log.Fatal(err)
		return Message{}
	}

	content := body.Find(".text").Text()

	message := Message{ID, dateSent, fromName, content}

	return message
}

func Parse(f io.Reader, messages *[]Message) error {
	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		return err
	}

	var prevFromName string
	doc.Find(".message").Each(func(i int, s *goquery.Selection) {
		message := parseMessage(s, &prevFromName)
		*messages = append(*messages, message)
	})

	return nil
}
