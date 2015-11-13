package serverApi

import (
	"encoding/json"
	// "fmt"
	"github.com/jeffail/gabs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type (
	Server struct {
		Url     string
		FullUrl string // Url built with api key or other credentials
		ApiKey  string
	}

	CouchPotato struct {
		Server
		Success bool `json:"success"`
	}

	Sonarr struct {
		Server
	}

	Plex struct {
		Server
	}
)

func EncodeUrl(str string) (string, error) {
	u, err := url.Parse(str)

	if err != nil {
		return "", err
	}

	return u.String(), nil
}

func (s *Server) BuildUrl() {
	s.FullUrl = s.Url + "api/" + s.ApiKey
}

// func Search() map[string]string {}
// Method type may be a problem
func (s *Server) Search(title string) []map[string]string {
	encodedTitle, encodeErr := EncodeUrl(title)

	if encodeErr != nil {
		log.Fatal(encodeErr)
	}

	query := "/search/?q=" + encodedTitle

	url := s.FullUrl + query

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
		"id":    "movies.tmdb_id",
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

func (c CouchPotato) TestConnection() bool {
	query := "/app.available"
	resp, err := http.Get(c.FullUrl + query)

	// False by default
	var result bool = false

	if err == nil {
		body, _ := ioutil.ReadAll(resp.Body)

		// Change type to string
		newBody := string(body)

		var r CouchPotato

		// Make usable via Go
		_err2 := json.Unmarshal([]byte(newBody), &r)

		if _err2 != nil {
			panic(_err2)
		}

		// fmt.Println(r.Success)

		result = r.Success
	}

	defer resp.Body.Close()

	return result
}
