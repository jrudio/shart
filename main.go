package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	config.Couchpotato.BuildURL()

	gin.SetMode(gin.ReleaseMode)

	// Start server
	router := gin.Default()

	v1 := router.Group("/v1")
	{
		v1.POST("/couchpotato", parseMediaRequest(config.Couchpotato))
		// v1.POST("/sonarr", parseMediaRequest(config.Sonarr))
		// v1.POST("/plex", parseMediaRequest(config.Plex))
	}

	log.WithField("host", config.Shart.Host).Info("listening for connections...")

	// Start up server to listen for commands coming from Slack
	routerErr := router.Run(config.Shart.Host)

	if routerErr != nil {
		log.Fatal(routerErr)
	}
}

func parseMediaRequest(srvr server) func(c *gin.Context) {
	return func(c *gin.Context) {
		// Get form data from POST request
		token := c.PostForm("token")

		// But first check the request was from Slack
		if token != srvr.slackToken() {
			c.String(http.StatusUnauthorized, "Not Authorized")
			return
		}

		channel := c.PostForm("channel_name")
		media := c.PostForm("text")

		// Ack the user
		// c.String(http.StatusOK, "Request received!")

		go func() {
			cmd, args := srvr.parseSlackInput(media)

			payload, err := srvr.doAction(cmd, args)

			if err != nil {
				log.WithFields(log.Fields{
					"command": cmd,
					"args":    args,
				}).Error(err)

				errPayload := slackPayload{
					Channel: channel,
					Text:    "failed to parse command",
				}

				replyToChannel(errPayload)
				return
			}

			// text = srvr.formatText(cmd, text)
			payload.Channel = channel

			// Reply to same channel as botName
			replyToChannel(payload)
		}()
	}
}

func replyToChannel(payload slackPayload) {
	payload.Username = config.Slack.BotName
	payload.Markdown = true

	payloadBytes, err := payload.toBytes()

	if err != nil {
		log.Error(err)
		return
	}

	// Send request
	var resp *http.Response
	resp, err = post(config.Slack.IncomingWebhook, payloadBytes)

	if err != nil {
		log.Error(err)
		return
	}

	if err = resp.Body.Close(); err != nil {
		log.WithFields(log.Fields{
			"channel": payload.Channel,
			"payload": payload,
		}).Error(err)
	}
}
