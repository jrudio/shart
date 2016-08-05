package main

type (
	server interface {
		// ParseCMDAndArgs(cmd string) string
		doAction(cmd string, args []string) (string, error)
		// formatText(cmd, result string) string
		parseSlackInput(input string) (string, []string)
		slackToken() string
		// testConnection() bool
		// BuildURL()
		// Search(title string) (server, error)
		// TODO: When I start working on Sonarr I may have to
		// make this function name more generic for Sonarr
		// AddMovieToWanted(mediaID string) string
		// RemoveMovieFromWanted(mediaID string) string
		// ShowWanted(startsWith, limitOffset string) (wantedList, error)
		// formatSearch(title string) string
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
