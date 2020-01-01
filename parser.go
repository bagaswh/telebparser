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
	"github.com/bagaswh/telebparser/utils"
)

// Constants of message types.
const (
	messageTypeText  = 1
	messageTypeVideo = 2
	messageTypeAudio = 3
	messageTypeVoice = 4
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
	DateSent string

	// The name of the user who sends the message.
	SenderName string

	// If this message is replying other message, this is the ID this message is replying to.
	// Empty string if not replying.
	ReplyToID string

	// Type of the message.
	MessageType int

	// Content is the message content.
	// This data could be anything: text, media, audio, etc.
	Content interface{}

	// The path of the media if the message is a media type.
	MediaPath string

	// The path of media's thumbnail.
	MediaThumbnailPath string
}

// getTimeComponents extracts individual "date components" (day, month, year, etc.) from the date string.
func getTimeComponents(dateString string) (day, month, year, hour, minute, second int) {
	// Regular expression for date string on `.date`'s title attribute inside `.body`.
	var dateRe = regexp.MustCompile("(\\d{2})\\.(\\d{2})\\.(\\d{4}) (\\d{2}):(\\d{2}):(\\d{2})")

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
		fromName = utils.GetText(body.Find(".from_name"))
		*prevFromName = fromName
	}

	dateString, _ := body.Find(".date").Attr("title")
	dateSent, err := parseTime(dateString, "Asia/Jakarta")
	if err != nil {
		log.Fatal(err)
		return Message{}
	}

	var replyToID string
	if el := body.Find(".reply_to"); utils.Exists(el) {
		href, _ := el.Find("a").Attr("href")
		replyToID = href[7:]
	}

	var messageType int

	// content parsing
	var content interface{}
	var mediaPath string
	var el *goquery.Selection
	if el = body.Find(".text"); utils.Exists(el) {
		messageType = messageTypeText
		content = utils.GetText(el)
	} else if el = body.Find(".media_wrap"); utils.Exists(el) {
		var mediaEl *goquery.Selection
		if mediaEl = el.Find(".video_file_wrap"); utils.Exists(mediaEl) {
			messageType = messageTypeVideo
		}
		mediaPath, _ = mediaEl.Attr("href")
	}

	return Message{ID, dateSent.Format("01/02/2006 15:04:05"), fromName, replyToID, messageType, content, mediaPath, ""}
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
	forEachFile(root, func(r io.Reader) error {
		doc, err := goquery.NewDocumentFromReader(r)
		if err != nil {
			return err
		}

		if messageRoom.RoomName == "" {
			// Parse room name.
			messageRoom.RoomName = utils.GetText(doc.Find(".page_header").Find(".text"))
		}

		parseFile(doc, &messageRoom.Messages)

		return nil
	})

	return nil
}
