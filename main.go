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

type Author struct {
	Name string `json:"name"`
}

type Provider struct {
	Name string `json:"name"`
}

type Field struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// https://discord.com/developers/docs/resources/channel#embed-object
type Embed struct {
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Author      Author   `json:"author,omitempty"`
	Provider    Provider `json:"provider,omitempty"`
	Fields      []Field  `json:"fields,omitempty"`
}

type DiscordMessage struct {
	Content string  `json:"content,omitempty"`
	Embeds  []Embed `json:"embeds,omitempty"`
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

	re := regexp.MustCompile(expression)

	t, err := tail.TailFile(file, tail.Config{Follow: true, ReOpen: true, Poll: true})
	if err != nil {
		panic(err)
	}
	for line := range t.Lines {

		fields := make([]Field, 0, len(re.SubexpNames())+2)
		fields = append(fields, Field{Name: "source", Value: sourceName})
		fields = append(fields, Field{Name: "file", Value: file})

		// This allows to use groups within the regex to find things like
		// nix run .#log2http -- -file fakefile -regexp '(?P<host>\w+) sshd\[\d+\]: Accepted publickey for (?P<user>\w+) from (?P<source>[\d\.]+)'
		// and use them as fields
		match := re.FindStringSubmatch(line.Text)
		if len(match) == 0 {
			continue
		}
		for i, name := range re.SubexpNames() {
			if i != 0 && name != "" {
				fields = append(fields, Field{Name: name, Value: match[i]})
			}
		}

		if err != nil {
			fmt.Println("ERROR", err)
			continue
		}
		if len(match) > 0 {
			fmt.Println("Matched", line.Text)
			message := DiscordMessage{
				Embeds: []Embed{
					Embed{
						Author:      Author{Name: fmt.Sprintf("%s on %s", file, sourceName)},
						Description: fmt.Sprintf("```\n%s\n```", line.Text),
						Fields:      fields,
					},
				}}
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
