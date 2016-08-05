package main

import (
	"encoding/json"
	"errors"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
)

type couchPotato struct {
	Host string `toml:"host" json:"-"`
	// Url built with api key or other credentials
	FullURL    string `json:"-"`
	APIKey     string `toml:"apiKey" json:"-"`
	SlackToken string `toml:"slackToken"`
	Success    bool   `json:"success"`
}

type wantedList struct {
	Movies []struct {
		Releases []struct {
			Status  string `json:"status"`
			Quality string `json:"quality"`
			ID      string `json:"_id"`
			MediaID string `json:"media_id"`
		} `json:"releases,omitempty"`
		Title   string `json:"title"`
		MediaID string `json:"_id"`
		Info    struct {
			Plot string `json:"plot"`
			Year int    `json:"year"`
		} `json:"info"`
	} `json:"movies"`
	Total   int  `json:"total"`
	Success bool `json:"success"`
}

type couchpotatoSearch struct {
	Movies []struct {
		ActorRoles map[string]string `json:"actor_roles"`
		Actors     []string          `json:"actors"`
		Directors  []string          `json:"directors"`
		Genres     []string          `json:"genres"`
		Images     struct {
			Actors           map[string]string `json:"actors"`
			Backdrop         []string          `json:"backdrop"`
			BackdropOriginal []string          `json:"backdrop_original"`
			Poster           []string          `json:"poster"`
			PosterOriginal   []string          `json:"poster_original"`
		} `json:"images"`
		Imdb          string `json:"imdb"`
		InLibrary     bool   `json:"in_library"`
		InWanted      bool   `json:"in_wanted"`
		Mpaa          string `json:"mpaa"`
		OriginalTitle string `json:"original_title"`
		Plot          string `json:"plot"`
		Rating        struct {
			Imdb []float64 `json:"imdb"`
		} `json:"rating"`
		Released string   `json:"released"`
		Runtime  int      `json:"runtime"`
		Tagline  string   `json:"tagline"`
		Titles   []string `json:"titles"`
		TmdbID   int      `json:"tmdb_id"`
		Type     string   `json:"type"`
		ViaImdb  bool     `json:"via_imdb"`
		ViaTmdb  bool     `json:"via_tmdb"`
		Writers  []string `json:"writers"`
		Year     int      `json:"year"`
	} `json:"movies"`
	Success bool `json:"success"`
}

// parseSlackInput gets first chunk of text; that should be the command
// Execute appropriate function
// TODO: Method type may be a problem
func (c couchPotato) parseSlackInput(input string) (string, []string) {
	// take care of no input
	if input == "" {
		return "", []string{}
	}

	// lowercase that shit
	input = strings.ToLower(input)

	args := strings.Fields(input)

	cmd := args[0]

	// command only
	if len(args) == 1 {
		return cmd, []string{}
	}

	// command and it's args
	return cmd, args[1:]

}

func (c couchPotato) doAction(cmd string, args []string) (slackPayload, error) {
	if cmd == "" {
		return slackPayload{}, errors.New("user failed to supply command")
	}

	switch cmd {
	// Add
	// case "add":
	// 	formattedText = config.Couchpotato.AddMovieToWanted(args)

	// Show
	// case "show":
	// 	if args == "wanted" {
	// 		// The user wants to display the wanted list
	// 		list, listErr := config.Couchpotato.ShowWanted("", "")

	// 		if listErr != nil {
	// 			formattedText = listErr.Error()
	// 		} else {
	// 			// Format the list for Slack
	// 			formattedText = formatWanted(list)
	// 		}
	// 	} else {
	// 		// TODO: Implement showing individual media with
	// 		// expanded information
	// 		formattedText = fmt.Sprintf("Showing %v\n", args)
	// 	}

	// Remove
	// case "remove":
	// 	formattedText = config.Couchpotato.RemoveMovieFromWanted(args)

	// search
	case "search":
		title := strings.Join(args, " ")
		searchResults, err := c.Search(title)

		if err != nil {
			log.WithFields(log.Fields{
				"command": "search",
				"args":    args,
			}).Error(err)

			return slackPayload{}, errors.New("search failed")
		}

		if !searchResults.Success {
			return slackPayload{}, errors.New("search failed")
		}

		payload := slackPayload{
			Text:        "Searched for `" + title + "`:",
			Attachments: searchResults.formatSearch(),
		}

		return payload, nil
	case "test":
		testConn := c.TestConnection()

		formattedText := "Connection to CouchPotato "

		var color string

		if testConn {
			formattedText += "worked!"
			color = "good"
		} else {
			formattedText += "failed!"
			color = "bad"
		}

		payload := slackPayload{
			Attachments: []slackPayloadAttachment{
				slackPayloadAttachment{
					Color: color,
					Text:  formattedText,
				},
			},
		}

		return payload, nil
	default:
		return slackPayload{}, errors.New("command not recognized")
	}
}

func (c couchpotatoSearch) formatSearch() []slackPayloadAttachment {
	searchResultLen := len(c.Movies)

	attachments := make([]slackPayloadAttachment, searchResultLen)

	for ii, movie := range c.Movies {
		year := strconv.Itoa(movie.Year)

		if year == "" {
			year = "n/a"
		}

		attachments[ii] = slackPayloadAttachment{
			Color:          "#000",
			Title:          movie.OriginalTitle + " (" + year + ")",
			Text:           movie.Plot,
			Fallback:       "You are unable to interact with Couchpotato search",
			AttachmentType: "default",
			CallbackID:     "",
			Actions: []slackPayloadAction{
				slackPayloadAction{
					Name:  "add_wanted",
					Text:  "Add to wanted",
					Type:  "button",
					Value: "add_wanted",
					Confirm: slackPayloadConfirm{
						Title:       "Are you sure?",
						Text:        "Adding _" + movie.OriginalTitle + "_",
						OKText:      "yes",
						DismissText: "no",
					},
				},
			},
		}
	}

	return attachments
}

func (c *couchPotato) BuildURL() {
	c.FullURL = c.Host + "/api/" + c.APIKey
}

func (c couchPotato) slackToken() string {
	return c.SlackToken
}

func (c couchPotato) Search(title string) (couchpotatoSearch, error) {
	encodedTitle, err := encodeURL(title)

	if err != nil {
		return couchpotatoSearch{}, err
	}

	query := c.FullURL + "/search/?q=" + encodedTitle

	var resp *http.Response
	resp, err = get(query)

	if err != nil {
		return couchpotatoSearch{}, err
	}

	defer resp.Body.Close()

	var result couchpotatoSearch

	err = json.NewDecoder(resp.Body).Decode(&result)

	return result, err
}

// ShowWanted shows the wanted list from CouchPotato.
// startsWith can be an empty string to show the whole wanted list
// limitOffset can be passed in the form "50" or "50,30". Empty shows all
func (c couchPotato) ShowWanted(startsWith, limitOffset string) (wantedList, error) {
	query := "/media.list/?"

	// Show the wanted list
	query += "status=active"

	query += "&type=movie"

	if len(startsWith) > 0 {
		query += "&starts_with=" + startsWith
	}

	if len(limitOffset) > 0 {
		query += "&limits_offset=" + limitOffset
	}

	reqURL := c.FullURL + query

	resp, err := get(reqURL)

	if err != nil {
		return wantedList{}, err
		// return "Error: " + bodyErr.Error()
	}

	defer resp.Body.Close()

	var list wantedList

	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return wantedList{}, err
	}

	return list, nil
}

func (c couchPotato) AddMovieToWanted(mediaID string) string {
	if mediaID == "" {
		return "Error: Cannot add movie. Please provide the imdb_id"
	}

	query := "/movie.add/?identifier="

	query += mediaID

	query = c.FullURL + query

	// Parse the response
	resp, err := get(query)

	if err != nil {
		return "Error: " + err.Error()
	}

	defer resp.Body.Close()

	type movieAdd struct {
		Success bool `json:"success"`
	}

	var result movieAdd

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "Error: " + err.Error()
	}

	if !result.Success {
		return "Failed to add movie to the wanted list"
	}

	return "Successfully added movie to the wanted list"
}

func (c couchPotato) removeMovie(mediaID, fromList string) (*http.Response, error) {
	if fromList == "" {
		fromList = "all"
	}

	// Build the query
	query := "/movie.delete/?id="

	query += mediaID

	query += "&delete_from="

	query += fromList

	// Build the url
	query = c.FullURL + query

	return get(query)
}

func (c couchPotato) RemoveMovieFromWanted(mediaID string) string {
	if mediaID == "" {
		return "Error: Cannot remove movie. Please provide the media id."
	}

	resp, err := c.removeMovie(mediaID, "wanted")

	if err != nil {
		return "Error: " + err.Error()
	}

	result := struct {
		Success bool `json:"success"`
	}{}

	// Check for another error
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "Error: " + err.Error()
	}

	if !result.Success {
		return "Failed to remove movie from the wanted list"
	}

	return "Successfully removed movie from the wanted list"
}

func (c couchPotato) TestConnection() bool {
	query := c.FullURL + "/app.available"
	resp, err := get(query)

	if err != nil {
		log.WithField("couchpotato.test", c).Error(err)
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.WithField("couchpotato.test", c).Error(resp.Status)
		return false
	}

	var r couchPotato

	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		log.WithFields(log.Fields{
			"couchpotato.test": c,
			"reason":           "possibly bad api key",
		}).Error(err)
		return false
	}

	return r.Success
}
