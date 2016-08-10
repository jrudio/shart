package main

import (
	"encoding/json"
	"errors"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
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
	// Show
	case "show":
		argsCount := len(args)

		if argsCount == 0 {
			return slackPayload{}, errors.New("At least one argument is needed for `show`")
		}

		if args[0] == "wanted" {
			// The user wants to display the wanted list
			list, err := c.ShowWanted("", "")

			if err != nil {
				return slackPayload{}, err
			}

			// Format the list for Slack
			return slackPayload{
				Text:        "Showing `wanted`:",
				Attachments: list.formatWanted(),
			}, nil
		}

		// TODO: Implement showing individual media with
		// expanded information
		return slackPayload{
			Text: "Showing information for id: `" + args[0] + "`",
			// Attachments: c.formatWanted(),
		}, nil

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
			Title:          movie.OriginalTitle + " (" + year + ")",
			Text:           movie.Plot,
			Fallback:       "You are unable to interact with Couchpotato search",
			AttachmentType: "default",
		}

		if !movie.InWanted {
			attachments[ii].Fields = []slackPayloadFields{
				slackPayloadFields{
					Value: "",
				},
				slackPayloadFields{
					Value: "<" + config.Shart.RootURL + "/couchpotato/add?imdb_id=" + movie.Imdb + "|Add to wanted>",
					Short: true,
				},
			}
		}

		if movie.InLibrary {
			attachments[ii].Actions = append(attachments[ii].Actions, slackPayloadAction{
				Text: "Movie already in library",
			})
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
	// encodedTitle, err := encodeURL(title)
	encodedTitle := url.QueryEscape(title)
	var err error

	// if err != nil {
	// 	return couchpotatoSearch{}, err
	// }

	query := c.FullURL + "/search/?q=" + encodedTitle

	var resp *http.Response
	resp, err = get(query)

	if err != nil {
		return couchpotatoSearch{}, err
	}

	defer resp.Body.Close()

	var result couchpotatoSearch

	err = json.NewDecoder(resp.Body).Decode(&result)

	if err != nil {
		var bodyBytes []byte
		bodyBytes, err = ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Println(err)
			return result, err
		}

		log.Println(string(bodyBytes))
	}

	return result, err
}

func (c couchPotato) doUserReply(cmd string, args url.Values) (slackPayload, error) {
	if cmd == "" {
		return slackPayload{}, errors.New("user failed to supply command")
	}

	switch cmd {
	case "add":
		imdbID := args.Get("imdb_id")

		movieAdded, err := c.AddMovieToWanted(imdbID)

		if err != nil {
			return slackPayload{}, err
		}

		if !movieAdded {
			return slackPayload{
				Attachments: []slackPayloadAttachment{
					slackPayloadAttachment{
						Color: "danger",
						Text:  "Failed to add movie: " + imdbID,
					},
				},
			}, nil
		}

		payload := slackPayload{
			Attachments: []slackPayloadAttachment{
				slackPayloadAttachment{
					Color: "good",
					Text:  "Added movie: " + imdbID,
				},
			},
		}

		return payload, nil
	default:
		return slackPayload{}, errors.New("unknown command")
	}
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

	query = c.FullURL + query

	resp, err := get(query)

	if err != nil {
		return wantedList{}, err
	}

	defer resp.Body.Close()

	var list wantedList

	err = json.NewDecoder(resp.Body).Decode(&list)

	return list, err
}

func (w wantedList) formatWanted() []slackPayloadAttachment {
	attachments := []slackPayloadAttachment{
		slackPayloadAttachment{},
	}

	attachments[0].Fields = make([]slackPayloadFields, w.Total)

	for ii, movie := range w.Movies {
		year := strconv.Itoa(movie.Info.Year)

		attachments[0].Fields[ii] = slackPayloadFields{
			Value: movie.Title + " (" + year + ")",
			Short: true,
		}
	}

	return attachments
}

func (c couchPotato) AddMovieToWanted(mediaID string) (bool, error) {
	if mediaID == "" {
		return false, errors.New("imdb_id required")
	}

	query := "/movie.add/?identifier="

	query += mediaID

	query = c.FullURL + query

	// Parse the response
	resp, err := get(query)

	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	type movieAdd struct {
		Success bool `json:"success"`
	}

	var result movieAdd

	err = json.NewDecoder(resp.Body).Decode(&result)

	return result.Success, err
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
