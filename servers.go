package main

import (
	"encoding/json"
	"github.com/jeffail/gabs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type (
	server interface {
		BuildURL()
		Search(title string) []map[string]string
		TestConnection() bool
		// TODO: When I start working on Sonarr I may have to
		// make this function name more generic for Sonarr
		AddMovieToWanted(mediaID string) string
		RemoveMovieFromWanted(mediaID string) string
		ShowWanted(startsWith, limitOffset string) (wantedList, error)
	}

	couchPotato struct {
		Host    string `toml:"host"`
		FullURL string // Url built with api key or other credentials
		APIKey  string `toml:"apiKey"`
		Success bool   `json:"success"`
	}

	sonarr struct {
		Host    string `toml:"host"`
		FullURL string // Url built with api key or other credentials
		APIKey  string `toml:"apiKey"`
	}

	plex struct {
		Host    string `toml:"host"`
		FullURL string // URL built with api key or other credentials
		Token   string `toml:"token"`
	}

	wantedList struct {
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
)

func encodeURL(str string) (string, error) {
	u, err := url.Parse(str)

	if err != nil {
		return "", err
	}

	return u.String(), nil
}

func request(reqURL string) ([]byte, error) {
	// Send the request
	resp, respErr := http.Get(reqURL)

	// Check for an error
	if respErr != nil {
		return nil, respErr
	}

	// Parse the response
	body, readBodyErr := ioutil.ReadAll(resp.Body)

	// Check for another error
	if readBodyErr != nil {
		return nil, readBodyErr
	}

	// Close the reader
	resp.Body.Close()

	return body, nil
}

func (c *couchPotato) BuildURL() {
	c.FullURL = c.Host + "/api/" + c.APIKey
}

func (c *couchPotato) Search(title string) []map[string]string {
	encodedTitle, encodeErr := encodeURL(title)

	if encodeErr != nil {
		log.Fatal(encodeErr)
	}

	query := "/search/?q=" + encodedTitle

	url := c.FullURL + query

	resp, reqErr := http.Get(url)

	if reqErr != nil {
		log.Fatal(reqErr)
	}

	defer resp.Body.Close()

	// TODO: Create a function for the following. I am repeating
	// this in TestConnection()
	body, _ := ioutil.ReadAll(resp.Body)

	bBody := string(body)

	searchResult, parseErr := gabs.ParseJSON([]byte(bBody))

	if parseErr != nil {
		log.Fatal(parseErr)
	}

	paths := map[string]string{
		"id":    "movies.imdb",
		"title": "movies.original_title",
		"year":  "movies.year",
		"plot":  "movies.plot",
	}

	searchResultLength, _ := searchResult.ArrayCountP("movies")

	result := make([]map[string]string, searchResultLength)

	// Extract title year and plot
	// Display result so I can figure out how to manipulate it
	for ii := 0; ii < searchResultLength; ii++ {
		id, _ := searchResult.ArrayElementP(ii, paths["id"])
		title, _ := searchResult.ArrayElementP(ii, paths["title"])
		year, _ := searchResult.ArrayElementP(ii, paths["year"])
		plot, _ := searchResult.ArrayElementP(ii, paths["plot"])

		info := map[string]string{
			"id":    id.String(),
			"title": title.Data().(string),
			"year":  year.String(),
			"plot":  plot.Data().(string),
		}

		result[ii] = info
	}
	return result
}

// ShowWanted shows the wanted list from CouchPotato.
// startsWith can be an empty string to show the whole wanted list
// limitOffset can be passed in the form "50" or "50,30". Empty shows all
func (c *couchPotato) ShowWanted(startsWith, limitOffset string) (wantedList, error) {
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

	body, bodyErr := request(reqURL)

	if bodyErr != nil {
		return wantedList{}, bodyErr
		// return "Error: " + bodyErr.Error()
	}

	var list wantedList

	unmarshalErr := json.Unmarshal(body, &list)

	if unmarshalErr != nil {
		return wantedList{}, unmarshalErr
		// return "Error: " + unmarshalErr.Error()
	}

	// txt := ""
	// for _, movie := range list.Movies {
	// 	txt += movie.Title + ": " + movie.MediaID + "\n"
	// fmt.Println(movie.Title + ": " + movie.MediaID)

	// if len(movie.Releases) == 0 {
	// 	continue
	// }

	// for _, releases := range movie.Releases {
	// 	fmt.Println(movie.Title + ": " + releases.ID)
	// }
	// }

	// return txt
	return list, nil
}

func (c *couchPotato) AddMovieToWanted(mediaID string) string {
	if mediaID == "" {
		return "Error: Cannot add movie. Please provide the imdb_id"
	}

	query := "/movie.add/?identifier="

	query += mediaID

	reqURL := c.FullURL + query

	// Parse the response
	body, readBodyErr := request(reqURL)

	if readBodyErr != nil {
		return "Error: " + readBodyErr.Error()
	}

	type movieAdd struct {
		Success bool `json:"success"`
	}

	var result movieAdd

	unmarshallErr := json.Unmarshal(body, &result)

	if unmarshallErr != nil {
		return "Error: " + unmarshallErr.Error()
	}

	if !result.Success {
		return "Failed to add movie to the wanted list"
	}

	return "Successfully added movie to the wanted list"
}

func (c *couchPotato) removeMovie(mediaID, fromList string) ([]byte, error) {
	if fromList == "" {
		fromList = "all"
	}

	// Build the query
	query := "/movie.delete/?id="

	query += mediaID

	query += "&delete_from="

	query += fromList

	// Build the url
	reqURL := c.FullURL + query

	body, bodyErr := request(reqURL)

	if bodyErr != nil {
		return nil, bodyErr
	}

	// Convert from bytes to string
	// bodyStr := string(body)

	// If all is good, return that struct
	return body, nil
}

func (c *couchPotato) RemoveMovieFromWanted(mediaID string) string {
	if mediaID == "" {
		return "Error: Cannot remove movie. Please provide the media id."
	}

	body, bodyErr := c.removeMovie(mediaID, "wanted")

	if bodyErr != nil {
		return "Error: " + bodyErr.Error()
	}

	type mRemove struct {
		Success bool `json:"success"`
	}

	var result mRemove

	// Unmarshall body into a struct
	unmarshallErr := json.Unmarshal(body, &result)

	// Check for another error
	if unmarshallErr != nil {
		return "Error: " + unmarshallErr.Error()
	}

	if !result.Success {
		return "Failed to remove movie from the wanted list"
	}

	return "Successfully removed movie from the wanted list"
}

func (c couchPotato) TestConnection() bool {
	query := "/app.available"
	resp, err := http.Get(c.FullURL + query)

	if err != nil {
		log.Println("Test Connection: " + err.Error())
		return false
	}

	defer resp.Body.Close()

	body, readBodyErr := ioutil.ReadAll(resp.Body)

	if readBodyErr != nil {
		log.Println("Response Body: " + readBodyErr.Error())
		return false
	}

	// Change type to string
	newBody := string(body)

	var r couchPotato

	// Make usable via Go
	_err2 := json.Unmarshal([]byte(newBody), &r)

	if _err2 != nil {
		log.Println(_err2)
		return false
	}

	// fmt.Println(r.Success)

	return r.Success
}
