package main

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
