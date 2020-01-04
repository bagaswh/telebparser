package telebparser

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/bagaswh/telebparser/utils"
)

type InvalidDirectoryError struct {
	Directory string
	Cause     string
	Err       error
}

func (err *InvalidDirectoryError) Error() string {
	return fmt.Sprintf("[%s] Directory[%s] is not valid because it [%s]", err.Err, err.Directory, err.Cause)
}

// Constants of message types.
const (
	messageTypeText    = iota
	messageTypePhoto   = iota
	messageTypeSticker = iota
	messageTypeVideo   = iota
	messageTypeGIF     = iota
	messageTypeAudio   = iota
	messageTypeVoice   = iota
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

// ParseContent parses message content of each `.message` element.
func parseContent(s *goquery.Selection, isJoined bool) (messageType int, content, mediaPath, mediaThumbnailPath string) {
	var el *goquery.Selection
	var indexToFind int
	if isJoined {
		indexToFind = 1
	} else {
		indexToFind = 2
	}
	el = goquery.NewDocumentFromNode(s.Children().Get(indexToFind)).Selection
	var mediaThumbnailElementIndex int
	switch {
	case el.HasClass("text"):
		messageType = messageTypeText
		content = utils.GetText(el)
	case el.HasClass("media_wrap"):
		mediaEl := goquery.NewDocumentFromNode(el.Children().Get(0)).Selection
		switch {
		case mediaEl.HasClass("video_file_wrap"):
			messageType = messageTypeVideo
			mediaThumbnailElementIndex = 2
		case mediaEl.HasClass("photo_wrap"):
			messageType = messageTypePhoto
			mediaThumbnailElementIndex = 0
		case mediaEl.HasClass("sticker_wrap"):
			messageType = messageTypeSticker
			mediaThumbnailElementIndex = 0
		case mediaEl.HasClass("animated_wrap"):
			messageType = messageTypeGIF
			mediaThumbnailElementIndex = 1
		case mediaEl.HasClass("media_voice_message"):
			messageType = messageTypeVoice
		case mediaEl.HasClass("media_audio_file"):
			messageType = messageTypeAudio
		}
		mediaPath, _ = mediaEl.Attr("href")
		mediaThumbnailPath, _ = goquery.NewDocumentFromNode(mediaEl.Children().Get(mediaThumbnailElementIndex)).Attr("src")
	}
	return
}

// ParseMessage parses individual `.message` element.
func parseMessage(s *goquery.Selection, prevFromName *string) Message {
	var fromName string

	ID, _ := s.Attr("id")

	body := s.Find(".body")

	// Message inside `div.message` that has `joined` class is sent by previous message's sender (recursively).
	if s.HasClass("joined") {
		fromName = *prevFromName
	} else {
		fromNameEl := goquery.NewDocumentFromNode(body.Children().Get(1))
		fromName = utils.GetText(fromNameEl.Selection)
		*prevFromName = fromName
	}

	dateSent, _ := body.Find(".date").Attr("title")

	var replyToID string
	if el := body.Find(".reply_to"); utils.Exists(el) {
		href, _ := el.Find("a").Attr("href")
		replyToID = href[7:]
	}

	// content parsing
	messageType, content, mediaPath, mediaThumbnailPath := parseContent(body, s.HasClass("joined"))

	return Message{ID, dateSent, fromName, replyToID, messageType, content, mediaPath, mediaThumbnailPath}
}

// ParseFile parses an html file.
func parseFile(doc *goquery.Document, messages *[]Message) error {
	var prevFromName string
	doc.Find(".message").Each(func(i int, s *goquery.Selection) {
		if !s.HasClass("default") {
			return
		}
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

// Parse parses whole directory into single struct.
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
