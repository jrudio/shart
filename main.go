package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jeffail/gabs"
	"github.com/jrudio/shart/api"
	"github.com/jrudio/shart/commands"
	"log"
	"net/http"
)

var (
	couchPotato      server.CouchPotato
	slackToken       string
	slackIncomingUrl string
	botName          string = "CouchPotatoBot"
	host             string
)

func initVars() []error {
	flag.StringVar(&couchPotato.Url, "couchpotato-url", "", "couchpotato url")
	flag.StringVar(&couchPotato.ApiKey, "couchpotato-apikey", "", "couchpotato api key")
	flag.StringVar(&slackToken, "slack-token", "", "slack token used to authorize bot")
	flag.StringVar(&slackIncomingUrl, "slack-url", "", "slack url to send our messages to")
	flag.StringVar(&host, "host", ":4040", "host is the address you want shart to listen on")
	flag.StringVar(&botName, "bot-name", "MediaBot", "bot name is the name of the bot posting to your slack channel")

	flag.Parse()

	requiredArgs := map[string]string{
		"couchpotato-url":    couchPotato.Url,
		"couchpotato-apikey": couchPotato.ApiKey,
		"slack-token":        slackToken,
		"slack-url":          slackIncomingUrl,
	}

	argLen := len(requiredArgs)

	var err []error

	for key, arg := range requiredArgs {
		if arg == "" {
			err = append(err, errors.New(key+" is required"))
			continue
		}

		if argLen == 1 && len(err) == 0 {
			// fmt.Println("Args satisfied. Appending nil to []err")
			err = append(err, nil)
		}

		argLen--
	}

	// fmt.Println(err)

	return err
}

func main() {
	// No errors means initErr[0] == nil
	if initErr := initVars(); initErr[0] != nil {
		// We have more than one error present
		for _, err := range initErr {
			// Display multiple errors
			fmt.Println(err.Error())
		}

		return
	}

	couchPotato.BuildUrl()
	fmt.Println("CouchPotato URL:", couchPotato.FullUrl)

	// fmt.Println(couchPotato.RemoveMovieFromWanted("4a9cedabd75d4c0499616b42e57afeb6"))
	// fmt.Println(couchPotato.AddMovieToWanted("tt2869728"))
	// fmt.Println(couchPotato.AddMovieToWanted("Ride Along 2", "tt2869728"))
	// list, listErr := couchPotato.ShowWanted("", "")

	// if listErr != nil {
	// 	fmt.Println(listErr.Error())
	// 	return
	// }

	// fmt.Println(commands.FormatWanted(list))
	// fmt.Println(couchPotato.Search("Ride Along 2"))
	// return

	/* Start server */
	router := gin.Default()

	v1 := router.Group("/v1")
	{
		v1.POST("media", parseMediaRequest)
		v1.POST("m", parseMediaRequest)
	}

	// Start up server to listen for commands coming from Slack
	routerErr := router.Run(host)

	if routerErr != nil {
		log.Fatal(routerErr)
	}
}

func replyToChannel(channel, text string) {
	// Use gabs to generate json
	payload := gabs.New()

	payload.Set("#"+channel, "channel")
	payload.Set(botName, "username")
	payload.Set(text, "text")

	payloadBuffer := []byte(payload.String())

	// Send request
	resp, postErr := http.Post(slackIncomingUrl, "application/json", bytes.NewBuffer(payloadBuffer))

	if postErr != nil {
		log.Fatal(postErr)
	}

	defer resp.Body.Close()

	fmt.Printf("\nReply to channel status code: %v", resp.Status)
}

func parseMediaRequest(c *gin.Context) {
	// Ack the user
	defer c.String(200, "Request received!")

	// Destructure Post data
	token := c.PostForm("token")

	// But first check the request was from Slack
	if token != slackToken {
		c.String(403, "Not Authorized")
		return
	}

	channel := c.PostForm("channel_name")
	media := c.PostForm("text")

	go func() {
		// Parse <media> to get the requested commands, titles, etc
		txt := commands.ParseCmd(media, &couchPotato)

		// Reply to same channel as MediaBot
		replyToChannel(channel, txt)
	}()
}
