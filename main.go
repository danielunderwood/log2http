package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/nxadm/tail"
)

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

	client := NewDiscordWebhook(url)
	defer client.Close()

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
				fields = append(fields, Field{Name: name, Value: match[i], Inline: true})
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
			client.MessageQueue <- message
		}
	}
}
