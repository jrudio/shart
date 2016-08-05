package main

import (
	"errors"
	"flag"
	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"os"
)

type shartConfig struct {
	Slack       `toml:"slack"`
	Couchpotato couchPotato `toml:"couchpotato"`
	Sonarr      sonarr      `toml:"sonarr"`
	Plex        plex        `toml:"plex"`
	Shart       struct {
		Host string `toml:"host"`
	} `toml:"shart"`
}

var config shartConfig

func init() {
	// set log level
	showDebug := flag.Bool("debug", false, "log debugging information")

	if *showDebug {
		log.SetLevel(log.DebugLevel)
	}

	// get config values
	_, err := toml.DecodeFile("config.toml", &config)

	if err != nil {
		// likely file not found error, so write the default config to file
		log.Warn(err)
		log.Info("creating default config file...")

		var bytesWritten int
		bytesWritten, err = writeDefaultConfig("config.toml")

		if err != nil {
			// failing here means we should exit program
			log.WithError(err).Fatal("failed to write config")
		}

		log.WithField("bytes written", bytesWritten).Info("wrote the default config")

		log.Info("please edit the config.toml file")

		os.Exit(1)
	}

	// check required values
	var hasErrs bool

	if config.Slack.IncomingWebhook == "" {
		log.WithField("slack.incomingWebhook", "empty").Error("missing required arg")
		hasErrs = true
	} else if config.Slack.IncomingWebhook == "http://hooks.slack.com/services" {
		log.WithField("slack.incomingWebhook", config.Slack.IncomingWebhook).Error("invalid webhook - please go to https://slack.com/services/new/incoming-webhook to create a valid webhook")
		hasErrs = true
	}

	// check optional values and display missing as warning
	if config.Couchpotato.Host == "" {
		log.WithField("couchpotato", config.Couchpotato).Warn("missing arg")
	}

	if config.Sonarr.Host == "" || config.Sonarr.APIKey == "" {
		log.WithField("sonarr", config.Sonarr).Warn("missing args")
	}

	if config.Plex.Host == "" || config.Plex.Token == "" {
		log.WithField("plex", config.Plex).Warn("missing args")
	}

	if hasErrs {
		os.Exit(1)
	}

	// default to listen on localhost:4040 if 'shart.host' is not found
	if config.Shart.Host == "" {
		config.Shart.Host = ":4040"
	}

	// default botname to 'ShartBot'
	if config.Slack.BotName == "" {
		config.Slack.BotName = "ShartBot"
	}

	// build the api base urls here
	config.Couchpotato.BuildURL()
}

func writeDefaultConfig(filename string) (int, error) {
	file, err := os.Create(filename)

	if err != nil {
		return 0, err
	}

	defer file.Close()

	// grab default config
	var defaultConfig []byte
	defaultConfig, err = Asset("config.default.toml")

	if err != nil {
		return 0, err
	}

	var bytesWritten int
	bytesWritten, err = file.Write(defaultConfig)

	if err != nil {
		return bytesWritten, err
	}

	if bytesWritten < 1 {
		return bytesWritten, errors.New("failed to write the default config file")
	}

	return bytesWritten, nil
}
