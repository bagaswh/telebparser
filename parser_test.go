package telebparser

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func parse() error {
	filepath := "C:\\Users\\Bagas Wahyu Hidayah\\Downloads\\Telegram Desktop\\ChatExport_30_12_2019"

	var messageRoom MessageRoom
	Parse(filepath, &messageRoom)

	namesCount := map[string]int{}
	for _, v := range messageRoom.Messages {
		_, ok := namesCount[v.SenderName]
		if !ok {
			namesCount[v.SenderName] = 0
		}
		namesCount[v.SenderName]++
	}

	for k, v := range namesCount {
		fmt.Println(k, v)
	}

	f, _ := os.Create("messages.json")
	json.NewEncoder(f).Encode(messageRoom)

	return nil
}

func TestParser(t *testing.T) {
	parse()
}

func BenchmarkParser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parse()
	}
}
