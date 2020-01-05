package telebparser

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

var filepath = "C:\\Users\\Bagas Wahyu Hidayah\\Downloads\\Telegram Desktop\\ChatExport_02_01_2020 (1)"
var filepath2 = "C:\\Users\\Bagas Wahyu Hidayah\\Downloads\\Telegram Desktop\\ChatExport_02_01_2020"
var r, _ = os.Open(filepath + "\\messages.html")
var doc, _ = goquery.NewDocumentFromReader(r)
var messageEl = doc.Find(".message")
var body = messageEl.Find(".body")

func parse() error {
	var messageRoom MessageRoom
	Parse(filepath, &messageRoom)

	f, _ := os.Create("messages.json")
	json.NewEncoder(f).Encode(messageRoom)

	return nil
}

func TestParser(t *testing.T) {
	parse()
}

func BenchmarkParseFile(b *testing.B) {
	messages := make([]Message, 0, 0)
	for i := 0; i < b.N; i++ {
		parseFile(doc, &messages)
	}
}

func BenchmarkParseMessage(b *testing.B) {
	var name string
	for i := 0; i < b.N; i++ {
		parseMessage(messageEl, &name)
	}
}

func BenchmarkParseContent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parseContent(body)
	}
}
