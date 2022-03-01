package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	"github.com/nxadm/tail"
)

type DiscordMessage struct {
	Content string `json:"content"`
}

func main() {
	var file, expression, url, sourceName string
	flag.StringVar(&file, "file", "", "File to read. Required")
	flag.StringVar(&url, "url", "", "URL to POST to. Must be supplied or in environment")
	flag.StringVar(&expression, "regexp", ".*", "Expression to match, defaults to `.*`")
	flag.StringVar(&sourceName, "sourceName", "", "Name of source. Defaults to hostname")

	flag.Parse()

	if len(file) == 0 {
		fmt.Println("Usage: -file FILENAME -regexp EXPRESSION")
		os.Exit(1)
	}
	if len(url) == 0 {
		url = os.Getenv("URL")
	}
	if len(url) == 0 {
		fmt.Println("ERROR: URL must be supplied as -url or in URL environment")
		os.Exit(1)
	}
	if len(sourceName) == 0 {
		sourceName, _ = os.Hostname()
	}

	t, err := tail.TailFile(file, tail.Config{Follow: true, ReOpen: true, Poll: true})
	if err != nil {
		panic(err)
	}
	for line := range t.Lines {
		matched, err := regexp.MatchString(expression, line.Text)
		if err != nil {
			fmt.Println("ERROR", err)
			continue
		}
		if matched {
			fmt.Println("Matched", line.Text)
			message := DiscordMessage{Content: fmt.Sprintf("[%s][%s] matched line: %s", sourceName, file, line.Text)}
			body, err := json.Marshal(message)
			if err != nil {
				fmt.Print("Failed to marshal JSON", err)
				continue
			}

			resp, err := http.Post(
				url,
				"application/json",
				bytes.NewReader(body),
			)

			if err != nil {
				fmt.Println("ERROR", err)
				continue
			}

			if resp.StatusCode != 204 {
				body, _ := ioutil.ReadAll(resp.Body)
				fmt.Println("Request failed", resp.StatusCode, string(body))
			}
		}
	}
}
