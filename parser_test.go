package telebparser

import (
	"fmt"
	"testing"
)

func parse() error {
	filepath := "C:\\Users\\Bagas Wahyu Hidayah\\Downloads\\Telegram Desktop\\ChatExport_30_12_2019"

	var messageRoom MessageRoom
	Parse(filepath, &messageRoom)
	fmt.Println(messageRoom)

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
