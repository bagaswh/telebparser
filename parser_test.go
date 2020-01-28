package telebparser

import (
	"testing"
)

var path = "/media/bagaswh/72F061FAF061C547/Users/Bagas Wahyu Hidayah/Downloads/Telegram Desktop/ChatExport_02_01_2020 (1)"
var invalidPath = "/home/bagaswh"

func parse() error {
	var messageRoom MessageRoom
	Parse(path, &messageRoom, 4)

	// f, _ := os.Create("messages.json")
	// json.NewEncoder(f).Encode(messageRoom)

	return nil
}

func parseInvalidDirectoryError() error {
	var messageRoom MessageRoom
	err := Parse(invalidPath, &messageRoom, 1)
	if err != nil {
		return err
	}
	return nil
}

func TestParse(t *testing.T) {
	if err := parse(); err != nil {
		t.Error(err)
	}
}

func TestParseError(t *testing.T) {
	if err := parseInvalidDirectoryError(); err == nil {
		t.Error("Invalid directory: error must be thrown!")
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parse()
	}
}
