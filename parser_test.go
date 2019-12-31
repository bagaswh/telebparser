package telebparser

import (
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"testing"
)

func parse() error {
	filepath := "C:\\Users\\Bagas Wahyu Hidayah\\Downloads\\Telegram Desktop\\ChatExport_30_12_2019"
	dirs, err := ioutil.ReadDir(filepath)
	if err != nil {
		return err
	}

	messages := make([]Message, 0, 100)

	re := regexp.MustCompile("^messages\\d*\\.html")
	for i := range dirs {
		if dirs[i].IsDir() {
			continue
		}

		filename := dirs[i].Name()
		fullpath := path.Join(filepath, filename)
		if re.Match([]byte(filename)) {
			f, err := os.Open(fullpath)
			if err != nil {
				return err
			}

			Parse(f, &messages)
			// fmt.Println(messages)

			f.Close()
		}

	}

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
