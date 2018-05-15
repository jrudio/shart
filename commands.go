package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type d struct {
	cmds    map[string]func(channelID string, args ...string)
	discord *discordgo.Session
}

func newDiscord(session *discordgo.Session) d {
	return d{
		cmds:    map[string]func(channelID string, args ...string){},
		discord: session,
	}
}

func (discord d) addCommand(cmd string, fn func(channelID string, args ...string)) {
	discord.cmds[cmd] = fn
}

func (discord d) execute(channelID, cmd string, args ...string) {
	if fn, ok := discord.cmds[cmd]; ok {
		fn(channelID, args...)
	} else {
		if isVerbose {
			fmt.Printf("invalid command: %s\n", cmd)
		}
	}
}

func (discord d) isValid(cmd string) bool {
	_, ok := discord.cmds[cmd]

	return ok
}

func (discord d) showHelp(channelID string) {
	msg := "Here is a list of available commands: \n"

	for key := range discord.cmds {
		msg += "`" + key + "`\n"
	}

	_, err := discord.discord.ChannelMessageSend(channelID, msg)

	if err != nil {
		fmt.Printf("failed to send command list to channel %s: %v\n",
			channelID,
			err)
	}
}

func (discord d) showError(channelID, msg string) {
	_, err := discord.discord.ChannelMessageSend(channelID, msg)

	if err != nil && isVerbose {
		fmt.Printf("send message failed: %v", err)
	}
}
