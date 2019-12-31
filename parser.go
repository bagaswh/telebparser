package telebparser

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// MessageRoom represents one chat room.
type MessageRoom struct {
	// The name of the room.
	RoomName string

	// List of messages in the room.
	Messages []Message
}

// Message represents one individual message sent by user.
type Message struct {
	// ID is unique for each message.
	// It is assigned in `id` attribute.
	ID string

	// When the message is sent.
	DateSent time.Time

	// The name of the user who sends the message.
	SenderName string

	// If this message is replying other message, this is the ID this message is replying to.
	// Empty string if not replying.
	ReplyToID string

	// Content is the message content.
	// This data could be anything: text, media, audio, etc.
	Content interface{}
}

// Regular expression for date string on `.date`'s title attribute inside `.body`.
var dateRe = regexp.MustCompile("(\\d{2})\\.(\\d{2})\\.(\\d{4}) (\\d{2}):(\\d{2}):(\\d{2})")

// getTimeComponents extracts individual "date components" (day, month, year, etc.) from the date string.
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

// getTimeValue creates time.Time value from extracted date components.
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

// parseMessage parses individual `.message` element.
func parseMessage(s *goquery.Selection, prevFromName *string) Message {
	if !s.HasClass("default") {
		return Message{}
	}

	var fromName string

	ID, _ := s.Attr("id")

	body := s.Find(".body")

	// Message inside `div.message` that has `joined` class is sent by previous message's sender (recursively).
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

	// content parsing is tricky
	var content interface{}
	content = body.Find(".text").Text()

	return Message{ID, dateSent, fromName, "", content}
}

// parseFile parses an html file.
func parseFile(doc *goquery.Document, messages *[]Message) error {
	var prevFromName string
	doc.Find(".message").Each(func(i int, s *goquery.Selection) {
		message := parseMessage(s, &prevFromName)
		*messages = append(*messages, message)
	})

	return nil
}

func forEachFile(root string, fn func(r io.Reader) error) error {
	dirs, err := ioutil.ReadDir(root)
	if err != nil {
		return err
	}

	re := regexp.MustCompile("^messages\\d*\\.html")
	for i := range dirs {
		if dirs[i].IsDir() {
			continue
		}

		filename := dirs[i].Name()
		fullpath := path.Join(root, filename)
		if re.Match([]byte(filename)) {
			f, err := os.Open(fullpath)
			if err != nil {
				return err
			}

			fn(f)

			f.Close()
		}
	}

	return nil
}

func Parse(root string, messageRoom *MessageRoom) error {
	hasGotRoomName := false
	forEachFile(root, func(r io.Reader) error {
		doc, err := goquery.NewDocumentFromReader(r)
		if err != nil {
			return err
		}

		if !hasGotRoomName {
			// Parse room name.
			messageRoom.RoomName = doc.Find(".page_header").Find(".text").Text()
			hasGotRoomName = true
		}

		parseFile(doc, &messageRoom.Messages)

		return nil
	})

	return nil
}
