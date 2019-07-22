package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/jrudio/go-radarr-client"
	"github.com/jrudio/go-sonarr-client"
)

const errTokenRequired = "a token is required"

// utils.go holds network utils and function helpers

func get(query string) (*http.Response, error) {
	client := http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequest("GET", query, nil)

	if err != nil {
		return &http.Response{}, err
	}

	return client.Do(req)
}

func post(query string, body []byte) (*http.Response, error) {
	client := http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequest("POST", query, bytes.NewBuffer(body))

	if err != nil {
		return &http.Response{}, err
	}

	req.Header.Set("Content-type", "application/json")

	return client.Do(req)
}

func encodeURL(str string) (string, error) {
	u, err := url.Parse(str)

	if err != nil {
		return "", err
	}

	return u.String(), nil
}

// getCredentials grabs apikeys and auth tokens via flags or environment vars
// prioritizes flags
func getCredentials() (serviceCredentials, error) {
	// TODO: implement environment vars

	credentials := serviceCredentials{}

	flag.StringVar(&credentials.shart.token, "token", "", "token used for bot authentication")
	flag.StringVar(&credentials.radarr.url, "radarr-url", "", "url that points to your radarr app")
	flag.StringVar(&credentials.radarr.apiKey, "radarr-key", "", "api key used for radarr")
	flag.StringVar(&credentials.sonarr.url, "sonarr-url", "", "url that points to your sonarr app")
	flag.StringVar(&credentials.sonarr.apiKey, "sonarr-key", "", "api key used for sonarr")
	flag.BoolVar(&isVerbose, "verbose", false, "output more inforation")
	versionFlag = flag.Bool("version", false, "get program version")

	flag.Parse()

	// check for version flag
	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	if credentials.shart.token == "" {
		return credentials, errors.New("a token is required")
	}

	return credentials, nil
}

type discordTOML struct {
	Token string
}

type sonarrTOML struct {
	Host string
	Key  string
}

type radarrTOML struct {
	Host string
	Key  string
}

type credentialWrapper struct {
	Discord discordTOML
	Sonarr  sonarrTOML
	Radarr  radarrTOML
}

// getCredentialsTOML grabs apikeys and auth tokens via .toml file
func getCredentialsTOML(path string) (serviceCredentials, error) {
	credWrapper := credentialWrapper{}
	credentials := serviceCredentials{}

	fileBytes, err := ioutil.ReadFile(path)

	if err != nil {
		return credentials, err
	}

	if err := toml.Unmarshal(fileBytes, &credWrapper); err != nil {
		return credentials, err
	}

	credentials = copyCreds(credWrapper, credentials)

	if credentials.shart.token == "" {
		return credentials, errors.New(errTokenRequired)
	}

	return credentials, nil
}

func copyCreds(credentialFrom credentialWrapper, credentialTo serviceCredentials) serviceCredentials {
	// radarr
	credentialTo.radarr.url = credentialFrom.Radarr.Host
	credentialTo.radarr.apiKey = credentialFrom.Radarr.Key

	// sonarr
	credentialTo.sonarr.url = credentialFrom.Sonarr.Host
	credentialTo.sonarr.apiKey = credentialFrom.Sonarr.Key

	// discord
	credentialTo.shart.token = credentialFrom.Discord.Token

	return credentialTo
}

func initializeClients(credentials serviceCredentials) (clients, error) {
	services := clients{}

	radarrClient, err := radarr.New(credentials.radarr.url, credentials.radarr.apiKey)

	if err != nil {
		return services, errors.New("radarr client failed: " + err.Error())
	}

	services.radarr = radarrClient

	sonarrClient, err := sonarr.New(credentials.sonarr.url, credentials.sonarr.apiKey)

	if err != nil {
		return services, errors.New("sonarr client failed: " + err.Error())
	}

	services.sonarr = sonarrClient

	return services, nil
}

func logPrint(chanID, message string) {
	fmt.Printf("%s - channel id: %s - %s\n", time.Now().String(), chanID, message)
}
