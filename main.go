package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jrudio/go-radarr-client"
	sonarr "github.com/jrudio/go-sonarr-client"
)

const (
	// keyword is the trigger for our program to listen to
	keyword = "shart"
)

var (
	keywordLen  = 0
	commandList commands
	isVerbose   bool
)

type commands interface {
	execute(channelID, cmd string, args ...string)
	isValid(cmd string) bool
	showHelp(channelID string)
	showError(channelID string, msg string)
	addCommand(cmd string, fn func(channelID string, args ...string))
}

type shartCredentials struct {
	token string
}

type radarrCredentials struct {
	url    string
	apiKey string
}

type sonarrCredentials struct {
	url    string
	apiKey string
}

type serviceCredentials struct {
	shart  shartCredentials
	radarr radarrCredentials
	sonarr sonarrCredentials
}

type clients struct {
	// TODO: maybe add discord here as well?
	radarr radarr.Client
	sonarr *sonarr.Sonarr
}

func checkErrAndExit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	credentials, err := getCredentials()

	if err != nil {
		fmt.Printf("need credentials: %v\n", err)
		os.Exit(1)
	}

	if keyword == "" {
		fmt.Println("a keyword (or trigger) is required for shart to work")
		os.Exit(1)
	}

	services, err := initializeClients(credentials)

	checkErrAndExit(err)

	discord, err := discordgo.New("Bot " + credentials.shart.token)

	checkErrAndExit(err)

	// get keyword length
	keywordLen = len(keyword)

	commandList := newDiscord(discord)

	commandList = addCommands(commandList, services)

	discord.AddHandler(onMsgCreate(commandList))

	err = discord.Open()

	checkErrAndExit(err)

	defer discord.Close()

	fmt.Println("bot is listening...")

	ctrlC := make(chan os.Signal, 1)

	signal.Notify(ctrlC, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	<-ctrlC
}

func onMsgCreate(commandList commands) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		if isVerbose {
			fmt.Println(m.Content)
		}

		messageLen := len(m.Content)

		if messageLen < keywordLen {
			return
		}

		// our keyword was not triggered -- ignore
		if keyword != m.Content[:keywordLen] {
			return
		}

		// user triggered keyword so lets see what subcommand was requested
		if messageLen > keywordLen {
			// user has a subcommand

			args := strings.Split(m.Content, " ")
			argCount := len(args)

			// remove the keyword
			args = args[1:argCount]

			argCount--

			subcommand := args[0]

			if !commandList.isValid(subcommand) {
				// let user know that command wasn't valid
				commandList.showError(m.ChannelID, "invalid command")
				return
			}

			// remove the subcommand
			args = args[1:argCount]

			commandList.execute(m.ChannelID, subcommand, args...)
		} else {
			// it's only the keyword so return a list of subcommands
			commandList.showHelp(m.ChannelID)
		}

		// TODO: maybe keep track of user and their subsequent commands
		// so multiple users don't mess each other up

		// fmt.Println(m.Content)
	}
}

func addCommands(commandList d, services clients) d {
	commandList.addCommand("search", func(channelID string, args ...string) {
		argCount := len(args)

		if argCount <= 0 {
			fmt.Printf("%s - channel id: %s - no args\n", time.Now().String(), channelID)
			commandList.showHelp(channelID)
			return
		}

		// we have to parse the first arg to know if we're dealing
		// with a movie or a show type search
		switch args[0] {
		case "movie":
			args = args[1:argCount]
			argCount--

			if argCount <= 0 {
				commandList.showError(channelID, "A title is required")
				return
			}

			title := strings.Join(args, " ")

			results, err := services.radarr.Search(title)

			if err != nil {
				commandList.showError(channelID, err.Error())
				return
			}

			resultCount := len(results)
			formattedResults := "No results found"

			if resultCount > 0 {
				formattedResults = "Here are your search results for `" + title + "`:\n"

				for _, result := range results {
					formattedResults += "- " + result.Title + " " + strconv.Itoa(result.Year) + " (" + strconv.Itoa(result.TmdbID) + ")\n"
				}
			}

			commandList.discord.ChannelMessageSend(channelID, formattedResults)

			return
		case "show":
			args = args[1:argCount]

			argCount--

			if argCount <= 0 {
				commandList.showError(channelID, "A title is required")
				return
			}

			title := strings.Join(args, " ")

			results, err := services.sonarr.Search(title)

			if err != nil {
				commandList.showError(channelID, err.Error())
				return
			}

			resultCount := len(results)
			formattedResults := "No results found"

			if resultCount > 0 {
				formattedResults = "Here are your search results for `" + title + "`:\n"

				for _, result := range results {
					formattedResults += "- " + result.Title + " " + strconv.Itoa(result.Year) + " (" + result.ImdbID + ")\n"
				}
			}

			commandList.discord.ChannelMessageSend(channelID, formattedResults)
		default:
			// unknown type
			msg := "Unknown media type"

			commandList.discord.ChannelMessageSend(channelID, msg)
			commandList.showHelp(channelID)
		}
	})

	// clear deletes messages in a channel -- user can delete x messages
	commandList.addCommand("clear", func(channelID string, args ...string) {
		argCount := len(args)
		messageLimit := 0

		if argCount > 0 {
			// make sure arg is an int
			limit, err := strconv.Atoi(args[0])

			if err != nil {
				fmt.Printf("%v - clear command - channel id %s - failed because arg: %v\n",
					time.Now().String(),
					channelID,
					err)

				return
			}

			messageLimit = limit
		}

		messages, err := commandList.discord.ChannelMessages(channelID, messageLimit, "", "", "")

		if err != nil {
			fmt.Printf("failed to retrieve message ids: %v\n", err)
			return
		}

		messageIDs := make([]string, len(messages))

		for i, message := range messages {
			messageIDs[i] = message.ID
		}

		if err := commandList.discord.ChannelMessagesBulkDelete(channelID, messageIDs); err != nil {
			fmt.Printf("failed to delete messages: %v\n", err)
		}
	})

	return commandList
}
