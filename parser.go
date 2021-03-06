package telebparser

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type InvalidDirectoryError struct {
	Directory string
	Cause     string
	Err       error
}

func (err *InvalidDirectoryError) Error() string {
	return fmt.Sprintf("[%s] Directory[%s] is not valid because [%s]", err.Err, err.Directory, err.Cause)
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
	mu sync.Mutex

	// The name of the room.
	RoomName string `json:"roomname,omitempty"`

	// List of messages in the room.
	Messages []Message `json:"messages,omitempty"`
}

// Message represents one individual message sent by user.
type Message struct {
	// ID is unique for each message.
	// It is assigned in `id` attribute.
	ID string `json:"id,omitempty"`

	// When the message is sent.
	DateMs int64 `json:"timestamp,omitempty"`

	// The name of the user who sends the message.
	SenderName string `json:"sender,omitempty"`

	// If this message is replying other message, this is the ID this message is replying to.
	// Empty string if not replying.
	ReplyToID string `json:"reply_to_id,omitempty"`

	// Type of the message.
	MessageType int `json:"messagetype,omitempty"`

	// Content is the message content.
	// This data could be anything: text, media, audio, etc.
	Content interface{} `json:"content,omitempty"`

	// The path of the media if the message is a media type.
	MediaPath string `json:"mediapath,omitempty"`

	// The path of media's thumbnail.
	MediaThumbnailPath string `json:"mediathumbnailpath,omitempty"`
}

// ParseContent parses message content of each `.message` element.
func parseContent(s *goquery.Selection) (messageType int, content, mediaPath, mediaThumbnailPath string) {
	var el *goquery.Selection
	if el = s.Find(".text"); exists(el) {
		messageType = messageTypeText
		content = getText(el)
	} else if el = s.Find(".media_wrap"); exists(el) {
		var mediaEl *goquery.Selection
		if mediaEl = el.Find(".video_file_wrap"); exists(mediaEl) {
			messageType = messageTypeVideo
			mediaThumbnailPath, _ = el.Find(".video_file").Attr("src")
		} else if mediaEl = el.Find(".photo_wrap"); exists(mediaEl) {
			messageType = messageTypePhoto
			mediaThumbnailPath, _ = el.Find(".photo").Attr("src")
		} else if mediaEl = el.Find(".sticker_wrap"); exists(mediaEl) {
			messageType = messageTypeSticker
			mediaThumbnailPath, _ = el.Find(".sticker").Attr("src")
		} else if mediaEl = el.Find(".animated_wrap"); exists(mediaEl) {
			messageType = messageTypeGIF
			mediaThumbnailPath, _ = el.Find(".animated").Attr("src")
		} else if mediaEl = el.Find(".media_voice_message"); exists(mediaEl) {
			messageType = messageTypeVoice
		} else if mediaEl = el.Find(".media_audio_file"); exists(mediaEl) {
			messageType = messageTypeAudio
		}
		mediaPath, _ = mediaEl.Attr("href")
	}
	return
}

var dateRe = regexp.MustCompile("(\\d{2})\\.(\\d{2})\\.(\\d{4}) (\\d{2})\\:(\\d{2})\\:(\\d{2})")

func getDateComponents(datestring string) (int, int, int, int, int, int) {
	dateComponents := dateRe.FindSubmatch([]byte(datestring))
	day, _ := strconv.Atoi(string(dateComponents[1]))
	month, _ := strconv.Atoi(string(dateComponents[2]))
	year, _ := strconv.Atoi(string(dateComponents[3]))
	hour, _ := strconv.Atoi(string(dateComponents[4]))
	minute, _ := strconv.Atoi(string(dateComponents[5]))
	second, _ := strconv.Atoi(string(dateComponents[6]))
	return day, month, year, hour, minute, second
}

// Caching
var timeLocationCache *time.Location
var timeLocationString string

func getTimeValue(datestring, locString string, mu *sync.Mutex) (time.Time, error) {
	mu.Lock()
	defer mu.Unlock()
	day, month, year, hour, minute, second := getDateComponents(datestring)
	if locString != timeLocationString {
		timeLocationString = locString
		timeLocation, err := time.LoadLocation(locString)
		if err != nil {
			return time.Time{}, err
		}
		timeLocationCache = timeLocation
	}
	return time.Date(year, time.Month(month), day, hour, minute, second, 0, timeLocationCache), nil
}

func toUnixEpoch(datestring string, mu *sync.Mutex) int64 {
	timeValue, err := getTimeValue(datestring, "Asia/Jakarta", mu)
	_ = timeValue
	if err != nil {
		log.Fatal(err)
	}
	return timeValue.Unix()
}

// ParseMessage parses individual `.message` element.
func parseMessage(s *goquery.Selection, mu *sync.Mutex, prevFromName *string) Message {
	var fromName string
	ID, _ := s.Attr("id")
	body := s.Find(".body")
	// Message inside `div.message` that has `joined` class is sent by previous message's sender (recursively).
	if s.HasClass("joined") {
		fromName = *prevFromName
	} else {
		fromNameEl := goquery.NewDocumentFromNode(body.Children().Get(1))
		fromName = getText(fromNameEl.Selection)
		*prevFromName = fromName
	}
	dateSent, _ := body.Find(".date").Attr("title")
	dateMs := toUnixEpoch(dateSent, mu)
	var replyToID string
	if el := body.Find(".reply_to"); exists(el) {
		href, _ := el.Find("a").Attr("href")
		replyToID = href[7:]
	}
	// content parsing
	messageType, content, mediaPath, mediaThumbnailPath := parseContent(body)
	return Message{ID, dateMs, fromName, replyToID, messageType, content, mediaPath, mediaThumbnailPath}
}

// ParseFile parses an html file.
func parseFile(doc *goquery.Document, mu *sync.Mutex, messages *[]Message) error {
	var prevFromName string
	doc.Find(".message").Each(func(i int, s *goquery.Selection) {
		if !s.HasClass("default") {
			return
		}
		message := parseMessage(s, mu, &prevFromName)
		*messages = append(*messages, message)
	})
	return nil
}

func isValidDirectory(filtered []os.FileInfo) bool {
	switch {
	case len(filtered) == 0:
		return false
	}
	return true
}

func getFiltered(root string, re *regexp.Regexp) ([]os.FileInfo, error) {
	dirs, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}
	dirsFiltered := make([]os.FileInfo, 0, 0)
	for i := range dirs {
		if dirs[i].IsDir() {
			continue
		}
		filename := dirs[i].Name()
		if re.Match([]byte(filename)) {
			dirsFiltered = append(dirsFiltered, dirs[i])
		}
	}
	if !isValidDirectory(dirsFiltered) {
		return nil, &InvalidDirectoryError{Directory: root, Cause: "directory does not contain [^messages\\d*.html] file", Err: errors.New("InvalidDirectoryError")}
	}
	return dirsFiltered, nil
}

// Parse parses whole directory into single struct.
func Parse(root string, messageRoom *MessageRoom, numCPU int) error {
	numCPU = int(math.Max(math.Min(float64(numCPU), float64(runtime.NumCPU())), 1.0))
	runtime.GOMAXPROCS(numCPU)

	re := regexp.MustCompile("^messages\\d*\\.html")
	dirsFiltered, err := getFiltered(root, re)
	if err != nil {
		return err
	}

	// parallel processing setup
	numGrs := numCPU
	var wg sync.WaitGroup
	wg.Add(numGrs)

	numDirs := len(dirsFiltered)
	// use not more than number of files' goroutines
	if numDirs < numGrs {
		numGrs = numDirs
	}

	filesPerGrs := numDirs / numGrs
	mod := numDirs % numGrs

	var first, last int
	for i := 0; i < numGrs; i++ {
		last = first + filesPerGrs + mod
		mod = 0

		go func(first int, last int) {
			defer wg.Done()

			// create local slice to store messages to avoid cache coherence
			messages := make([]Message, 0, 0)
			for j := first; j < last; j++ {
				fullpath := filepath.Join(root, dirsFiltered[j].Name())
				f, err := os.Open(fullpath)
				if err != nil {
					log.Fatal(err)
				}
				doc, err := goquery.NewDocumentFromReader(f)
				f.Close()
				parseFile(doc, &messageRoom.mu, &messages)
			}

			// messageRoom.Messages is shared, need lock
			messageRoom.mu.Lock()
			defer messageRoom.mu.Unlock()
			messageRoom.Messages = append(messageRoom.Messages, messages...)
		}(first, last)
		first = last
	}

	wg.Wait()
	return nil
}
