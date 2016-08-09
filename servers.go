package main

import "net/url"

type (
	server interface {
		doAction(cmd string, args []string) (slackPayload, error)
		doUserReply(cmd string, args url.Values) (slackPayload, error)
		// formatText(cmd, result string) string
		parseSlackInput(input string) (string, []string)
		slackToken() string
	}

	sonarr struct {
		Host string `toml:"host" json:"-"`
		// Url built with api key or other credentials
		FullURL    string `json:"-"`
		APIKey     string `toml:"apiKey" json:"-"`
		SlackToken string `toml:"slackToken"`
	}

	plex struct {
		Host       string `toml:"host" json:"-"`
		FullURL    string // URL built with api key or other credentials
		Token      string `toml:"token" json:"-"`
		SlackToken string `toml:"slackToken"`
	}
)
