package main

import "encoding/json"

// Slack information required to talk to Slack
type Slack struct {
	IncomingURL string `toml:"incomingURL"`
	Token       string `toml:"token"`
	BotName     string `toml:"botName"`
}

type slackPayload struct {
	Title       string                   `json:"title"`
	Text        string                   `json:"text"`
	Attachments []slackPayloadAttachment `json:"attachments"`
}

type slackPayloadAttachment struct {
	AuthorIcon string `json:"author_icon"`
	AuthorLink string `json:"author_link"`
	AuthorName string `json:"author_name"`
	Color      string `json:"color"`
	Fallback   string `json:"fallback"`
	Fields     []struct {
		Short bool   `json:"short"`
		Title string `json:"title"`
		Value string `json:"value"`
	} `json:"fields"`
	Footer     string `json:"footer"`
	FooterIcon string `json:"footer_icon"`
	ImageURL   string `json:"image_url"`
	Pretext    string `json:"pretext"`
	Text       string `json:"text"`
	ThumbURL   string `json:"thumb_url"`
	Title      string `json:"title"`
	TitleLink  string `json:"title_link"`
	Timestamp  int    `json:"ts"`
}

func (s slackPayload) toBytes() ([]byte, error) {
	return json.Marshal(&s)
}
