package main

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	config.couchPotato.BuildURL()

	gin.SetMode(gin.ReleaseMode)

	// Start server
	router := gin.Default()

	v1 := router.Group("/v1")
	{
		v1.POST("media", parseMediaRequest)
		v1.POST("m", parseMediaRequest)
	}

	// Start up server to listen for commands coming from Slack
	routerErr := router.Run(config.Shart.Host)

	if routerErr != nil {
		log.Fatal(routerErr)
	}
}

func replyToChannel(channel, text string) {
	// Use gabs to generate json
	payload := slackPayload{
		Title: "New message from " + config.Slack.BotName,
		Text:  text,
	}

	payloadBytes, err := payload.toBytes()

	if err != nil {
		log.Error(err)
		return
	}

	// Send request
	var resp *http.Response
	resp, err = http.Post(config.Slack.IncomingURL, "application/json", bytes.NewBuffer(payloadBytes))

	if err != nil {
		log.Error(err)
		return
	}

	defer resp.Body.Close()

	fmt.Printf("\nReply to channel status code: %v", resp.Status)
}

func parseMediaRequest(c *gin.Context) {
	// Ack the user

	// Destructure Post data
	token := c.PostForm("token")

	fmt.Println("token:", token)

	// But first check the request was from Slack
	if token != config.Slack.Token {
		c.String(403, "Not Authorized")
		return
	}

	channel := c.PostForm("channel_name")
	media := c.PostForm("text")

	c.String(200, "Request received!")

	go func() {
		// Parse <media> to get the requested commands, titles, etc
		txt := ParseCMD(media, &config.couchPotato)

		// Reply to same channel as MediaBot
		replyToChannel(channel, txt)
	}()
}
