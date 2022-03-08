package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/nxadm/tail"
)

func main() {
	var file, expression, url, sourceName, dedupeUri string
	flag.StringVar(&file, "file", "", "File to read. Required")
	flag.StringVar(&url, "url", "", "URL to POST to. Must be supplied or in environment")
	flag.StringVar(&expression, "regexp", ".*", "Expression to match, defaults to `.*`")
	flag.StringVar(&sourceName, "sourceName", "", "Name of source. Defaults to hostname")
	flag.StringVar(&dedupeUri, "dedupe", "", "URI for deduplication, such as bloom:///path/to/filter.bin. Defaults to no deduplication")

	flag.Parse()

	if len(file) == 0 {
		fmt.Println("Usage: -file FILENAME -regexp EXPRESSION -dedupe DEDUPE_URI")
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

	var dedupe Deduplicator
	if len(dedupeUri) == 0 {
		dedupe = NewNullDeduplicator()
	} else if strings.HasPrefix(dedupeUri, "bloom://") {
		// TODO These parameters should probably be tunable by the user
		dedupe = NewBloomFilterDeduplicator(dedupeUri[len("bloom://"):], 10000, 0.01)
	} else {
		fmt.Println("ERROR: Unable to parse deduplication URI")
		os.Exit(1)
	}
	// This is some weird go thing about how interfaces are stored in memory
	// Just comparing to nil probably won't work (it still has a type), so you have to use reflection to check the value
	if dedupe == nil || reflect.ValueOf(dedupe).IsNil() {
		fmt.Println("ERROR: Unable to create deduplicator")
		os.Exit(1)
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
		dedupeKey := []byte(line.Text)
		exists, err := dedupe.Exists(dedupeKey)
		if err != nil {
			fmt.Println("ERROR", err)
			continue
		}
		if exists {
			fmt.Println("Already exists!", line.Text)
		}
		if len(match) > 0 && !exists {
			fmt.Println("Matched", line.Text)
			message := DiscordMessage{
				Embeds: []Embed{
					{
						Author:      Author{Name: fmt.Sprintf("%s on %s", file, sourceName)},
						Description: fmt.Sprintf("```\n%s\n```", line.Text),
						Fields:      fields,
					},
				}}

			client.MessageQueue <- message
			dedupe.Add(dedupeKey)
		}
	}
}
