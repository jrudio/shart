package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

func main() {
	config.Couchpotato.BuildURL()

	gin.SetMode(gin.ReleaseMode)

	// Start server
	router := gin.Default()

	api := router.Group("/v1")
	{
		api.POST("/couchpotato", endpoint.parseMediaRequest(config.Couchpotato))
		// api.POST("/sonarr", parseMediaRequest(config.Sonarr))
		// api.POST("/plex", parseMediaRequest(config.Plex))
	}

	log.WithField("host", config.Shart.Host).Info("listening for connections...")

	// Start up server to listen for commands coming from Slack
	routerErr := router.Run(config.Shart.Host)

	if routerErr != nil {
		log.Fatal(routerErr)
	}
}
