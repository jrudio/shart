package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"net/http"
)

var endpoint endpointHandlers

type endpointHandlers struct{}

func (e endpointHandlers) parseMediaRequest(srvr server) func(c *gin.Context) {
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
		c.String(http.StatusOK, "Processing...")

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
					// Text:    "failed to parse command",
					Text: err.Error(),
				}

				e.replyToChannel(errPayload)
				return
			}

			payload.Channel = channel

			// Reply to same channel as botName
			e.replyToChannel(payload)
		}()
	}
}

func (e endpointHandlers) replyToChannel(payload slackPayload) {
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

// cmdReply handles a user's reply to a command
func (e endpointHandlers) cmdReply(srvr server) func(c *gin.Context) {
	return func(c *gin.Context) {
		// parse command and any args
		cmd := c.Param("cmd")

		if cmd == "" {
			c.String(http.StatusBadRequest, "Received empty command")
		}

		args := c.Request.URL.Query()

		payload, err := srvr.doUserReply(cmd, args)

		payload.Username = config.Slack.BotName
		payload.Markdown = true

		if err != nil {
			log.WithFields(log.Fields{
				"command": cmd,
				"args":    args,
			}).Error(err)

			errPayload := slackPayload{
				// Channel: channel,
				Text: err.Error(),
			}

			c.Redirect(http.StatusTemporaryRedirect, "slack://channel")
			go e.replyToChannel(errPayload)
			return
		}

		// payload.Channel = channel

		// Reply to same channel as botName
		e.replyToChannel(payload)

		c.Redirect(http.StatusTemporaryRedirect, "slack://channel")
	}
}
