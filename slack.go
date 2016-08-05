package main

import "encoding/json"

// Slack information required to talk to Slack
type Slack struct {
	IncomingWebhook string `toml:"incomingWebhook"`
	BotName         string `toml:"botName"`
}

type slackPayload struct {
	Channel     string                   `json:"channel"`
	Title       string                   `json:"title"`
	Text        string                   `json:"text"`
	Username    string                   `json:"username"`
	Markdown    bool                     `json:"mrkdwn"`
	Attachments []slackPayloadAttachment `json:"attachments"`
}

type slackPayloadAttachment struct {
	Actions        []slackPayloadAction `json:"actions"`
	AttachmentType string               `json:"attachment_type"`
	AuthorIcon     string               `json:"author_icon"`
	AuthorLink     string               `json:"author_link"`
	AuthorName     string               `json:"author_name"`
	CallbackID     string               `json:"callback_id"`
	Color          string               `json:"color"`
	Fallback       string               `json:"fallback"`
	Fields         []struct {
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

type slackPayloadAction struct {
	Name    string              `json:"name"`
	Text    string              `json:"text"`
	Type    string              `json:"type"`
	Value   string              `json:"value"`
	Style   string              `json:"style"`
	Confirm slackPayloadConfirm `json:"confirm"`
}

type slackPayloadConfirm struct {
	Title       string `json:"title"`
	Text        string `json:"text"`
	OKText      string `json:"ok_text"`
	DismissText string `json:"dismiss_text"`
}

func (s slackPayload) toBytes() ([]byte, error) {
	return json.Marshal(&s)
}
