package main

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jeffail/gabs"
	"github.com/jrudio/shart/api"
	"github.com/jrudio/shart/commands"
	"github.com/jrudio/shart/config"
	"log"
	"net/http"
)

var (
	couch            serverApi.Server
	slackToken       string
	slackIncomingUrl string
	botName          string
)

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
	defer c.String(200, "Request recieved!")

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
		txt := commands.ParseCmd(media, &couch)

		// Reply to same channel as MediaBot
		replyToChannel(channel, txt)
	}()
}

func main() {
	// TODO: Talk to CouchPotato, Sonarr, and Plex servers

	config, configErr := config.Init()

	if configErr != nil {
		log.Fatal(configErr)
	}

	/* Initialize vars */
	couch = serverApi.Server{
		Url:    config.Path("couchpotato.host").Data().(string),
		ApiKey: config.Path("couchpotato.apiKey").Data().(string),
	}

	couch.BuildUrl()
	fmt.Println("CouchPotato URL:", couch.FullUrl)

	slackToken = config.Path("slack.token").Data().(string)
	slackIncomingUrl = config.Path("slack.incomingUrl").Data().(string)
	botName = config.Path("slack.botName").Data().(string)

	/* Start server */
	router := gin.Default()

	v1 := router.Group("/v1")
	{
		v1.POST("media", parseMediaRequest)
	}

	// Start up server to listen for commands coming from Slack
	listenPort := config.Path("shart.port").Data().(string)

	if listenPort == "" {
		listenPort = "3000"
	}

	routerErr := router.Run(":" + listenPort)

	if routerErr != nil {
		log.Fatal(routerErr)
	}
}
