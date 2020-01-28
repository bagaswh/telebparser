package telebparser

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"testing"
)

var path string
var numCPU int
var showDebug bool
var printJson bool

func init() {
	flag.StringVar(&path, "path", "", "Root path of the backup files")
	flag.IntVar(&numCPU, "numcpu", runtime.NumCPU(), "Number of CPUs")
	flag.BoolVar(&showDebug, "showdebug", true, "Show debug informations")
	flag.BoolVar(&printJson, "printjson", false, "Print messages JSON output")
	// flag.Parse()
}

func parse() error {
	var messageRoom MessageRoom
	Parse(path, &messageRoom, numCPU)

	if showDebug {
		fmt.Printf("parsed %d messages\n", len(messageRoom.Messages))
	}

	if printJson {
		f, _ := os.Create("messages.json")
		json.NewEncoder(f).Encode(messageRoom)
	}

	return nil
}

func TestParse(t *testing.T) {
	if err := parse(); err != nil {
		t.Error(err)
	}
}

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parse()
	}
}
