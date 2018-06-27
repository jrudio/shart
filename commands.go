package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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

func (discord d) getCommands() []string {
	cmdsLen := len(discord.cmds)

	cmds := make([]string, cmdsLen)

	i := 0

	for commandName := range discord.cmds {
		cmds[i] = commandName

		if i++; i > cmdsLen {
			break
		}
	}

	return cmds
}

func (discord d) showError(channelID, msg string) {
	_, err := discord.discord.ChannelMessageSend(channelID, msg)

	if err != nil && isVerbose {
		fmt.Printf("send message failed: %v", err)
	}
}

func search(commandList d, services clients) func(channelID string, args ...string) {
	return func(channelID string, args ...string) {
		argCount := len(args)

		// we must have at least 2 args: media type and the title
		if argCount < 2 {
			fmt.Printf("%s - channel id: %s - no args\n", time.Now().String(), channelID)

			output := "search requires `<movie|show> <title>`\n"

			commandList.showError(channelID, output)
			return
		}

		// we have to parse the first arg to know if we're dealing
		// with a movie or a show type search
		mediaType := args[0]

		// remove media type from args
		args = args[1:argCount]
		argCount--

		if argCount <= 0 {
			commandList.showError(channelID, "A title is required")
			return
		}

		switch mediaType {
		case "movie":
			title := strings.Join(args, " ")

			results, err := services.radarr.Search(title)

			if err != nil {
				output := fmt.Sprintf("search failed: %v", err)
				commandList.showError(channelID, output)
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
			title := strings.Join(args, " ")

			results, err := services.sonarr.Search(title)

			if err != nil {
				fmt.Printf("%v - channel id: %s - %v\n", time.Now().String(), channelID, err)
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
			output := "unknown media type: %s\n\tshould be one of `movie|show`\n"
			output += "Here is a list of available commands: \n"

			for _, commandNames := range commandList.getCommands() {
				output += "`" + commandNames + "`\n"
			}

			commandList.showError(channelID, fmt.Sprintf(output, mediaType))
		}
	}
}

func showQualityProfiles(commandList d, services clients) func(channelID string, args ...string) {
	return func(channelID string, args ...string) {
		argCount := len(args)

		// we should have 1 arg
		if argCount < 1 {
			commandList.showError(channelID, "an arg `movie|show` is required")
			return
		}

		// first arg should be 'movie' or 'show'
		mediaType := args[0]

		switch mediaType {
		case "movie":
			profiles, err := services.radarr.GetProfiles()

			if err != nil {
				errMsg := fmt.Sprintf("failed to fetch profiles from radarr: %v\n", err)
				fmt.Printf(errMsg)
				commandList.showError(channelID, errMsg)
				return
			}

			output := "Here are the available quality profiles for radarr:\n"

			for _, profile := range profiles {
				output += fmt.Sprintf("\t`id: %d` %s\n", profile.ID, profile.Name)
			}

			if _, err := commandList.discord.ChannelMessageSend(channelID, output); err != nil {
				fmt.Printf("chan id: %s - %v\n", channelID, err)
				return
			}
		case "show":
			commandList.discord.ChannelMessageSend(channelID, "unable to get profiles as `show` is not implemented")
		default:
			errMsg := fmt.Sprintf("unknown media type: %s\n", mediaType)
			fmt.Printf(errMsg)
			commandList.showError(channelID, errMsg)
		}
	}
}

func setQualityProfile(commandList d, services clients) func(channelID string, args ...string) {
	return func(channelID string, args ...string) {
		argCount := len(args)

		// we should have 2 args
		if argCount < 2 {
			commandList.showError(channelID, "need more args: `movie|show` <quality-profile-id>")
			return
		}

		// first arg should be 'movie' or 'show'
		mediaType := args[0]
		qualityProfileID := args[1]

		// a profile quality id is required
		if qualityProfileID == "" {
			commandList.showError(channelID, "a quality profile id is required\n get id via command `quality`")
			return
		}

		profileID, err := strconv.Atoi(qualityProfileID)

		if err != nil {
			output := fmt.Sprintf("failed to convert profile quality id to int: %v", err)
			fmt.Printf("channel id: %s - %s\n", channelID, output)
			commandList.showError(channelID, output)
			return
		}

		// radarr and sonarr do not have an id of 0
		if profileID == 0 {
			commandList.showError(channelID, "please enter a valid quality profile id\n use `quality` to find valid id")
			return
		}

		switch mediaType {
		case "movie":
			defaultRadarrQualityID = profileID
			output := "successfully set movie quality to `%d`"
			commandList.discord.ChannelMessageSend(channelID, fmt.Sprintf(output, defaultRadarrQualityID))
		default:
			output := "unknown media type: %s\n\tshould be one of `movie|show`"
			commandList.discord.ChannelMessageSend(channelID, fmt.Sprintf(output, mediaType))
			return
		}
	}
}

func showRootFolders(commandList d, services clients) func(channelID string, args ...string) {
	return func(channelID string, args ...string) {
		argCount := len(args)

		// we should have 1 arg
		if argCount < 1 {
			commandList.showError(channelID, "an arg `movie|show` is required")
			return
		}

		// first arg should be 'movie' or 'show'
		mediaType := args[0]

		switch mediaType {
		case "movie":
			folders, err := services.radarr.GetRootFolders()

			if err != nil {
				errMsg := fmt.Sprintf("failed to fetch folders from radarr: %v\n", err)
				fmt.Printf(errMsg)
				commandList.showError(channelID, errMsg)
				return
			}

			output := "Here are the available root folders for radarr:\n"

			for _, folder := range folders {
				output += fmt.Sprintf("\t`id: %d` - %s\n", folder.ID, folder.Path)
			}

			if _, err := commandList.discord.ChannelMessageSend(channelID, output); err != nil {
				fmt.Printf("chan id: %s - %v\n", channelID, err)
				return
			}
		case "show":
			commandList.discord.ChannelMessageSend(channelID, "unable to get folders as `show` is not implemented")
		default:
			errMsg := fmt.Sprintf("unknown media type: %s\n", mediaType)
			fmt.Printf(errMsg)
			commandList.showError(channelID, errMsg)
		}
	}
}

func setRootFolder(commandList d, services clients) func(channelID string, args ...string) {
	return func(channelID string, args ...string) {
		argCount := len(args)

		// we should have 2 args
		if argCount < 2 {
			commandList.showError(channelID, "need more args: `movie|show` <root-folder-id|folder-path>")
			return
		}

		// first arg should be 'movie' or 'show'
		mediaType := args[0]
		folderPathOrID := args[1]

		// a path or path id is required
		if folderPathOrID == "" {
			commandList.showError(channelID, "a folder path or path id is required\n get id via command `folders`")
			return
		}

		switch mediaType {
		case "movie":

			if strings.HasPrefix(folderPathOrID, "/") {
				defaultRadarrPath = folderPathOrID
			} else {
				pathID, err := strconv.Atoi(folderPathOrID)

				if err != nil {
					output := fmt.Sprintf("failed to convert path id to int: %v", err)
					fmt.Printf("channel id: %s - %s\n", channelID, output)
					commandList.showError(channelID, output)
					return
				}

				folders, err := services.radarr.GetRootFolders()

				if err != nil {
					output := fmt.Sprintf("fetch radarr root folders failed: %v", err)
					fmt.Printf("channel id: %s - %s\n", channelID, output)
					commandList.showError(channelID, output)
					return
				}

				for _, folder := range folders {
					if folder.ID == pathID {
						defaultRadarrPath = folder.Path
					}
				}

				if defaultRadarrPath == "" {
					output := "could not find stored path via id: `%s`"
					commandList.showError(channelID, fmt.Sprintf(output, folderPathOrID))
					return
				}
			}

			output := "successfully set root folder to `%s`"
			commandList.discord.ChannelMessageSend(channelID, fmt.Sprintf(output, defaultRadarrPath))
		default:
			output := "unknown media type: %s\n\tshould be one of `movie|show`"
			commandList.discord.ChannelMessageSend(channelID, fmt.Sprintf(output, mediaType))
			return
		}
	}
}

func addMedia(commandList d, services clients) func(channelID string, args ...string) {
	return func(channelID string, args ...string) {
		argCount := len(args)

		// we should have 2 args
		if argCount < 2 {
			commandList.showError(channelID, "`movie|show <tmdb-id>`")
			return
		}

		// first arg should be 'movie' or 'show'
		mediaType := args[0]
		tmdbIDstr := args[1]

		if tmdbIDstr == "" {
			commandList.showError(channelID, "a tmdb id is required\n `movie|show <tmdb-id>`")
			return
		}

		switch mediaType {
		case "movie":
			// use radarr to add movie
			tmdbID, err := strconv.Atoi(tmdbIDstr)

			if err != nil {
				output := fmt.Sprintf("failed to convert tmdb id to int: %v", err)
				fmt.Printf("channel id: %s - %s\n", channelID, output)
				commandList.showError(channelID, output)
				return
			}

			// make sure profile quality and folder path are set
			if defaultRadarrPath == "" {
				commandList.showError(channelID, "aborting... a root folder path must be set")
				commandList.showHelp(channelID)
				return
			}

			if defaultRadarrQualityID == 0 {
				commandList.showError(channelID, "aborting... a profile quality must be set")
				commandList.showHelp(channelID)
				return
			}

			requestedMovie, err := services.radarr.GetMovie(tmdbID)

			if err != nil {
				fmt.Printf("failed to add movie: %v\n", err)
				commandList.showError(channelID, fmt.Sprintf("failed fetching movie: %v", err))
				return
			}

			// tweak fields to make a proper request
			requestedMovie.AddOptions.SearchForMovie = true
			requestedMovie.Monitored = true
			requestedMovie.QualityProfileID = defaultRadarrQualityID
			requestedMovie.RootFolderPath = defaultRadarrPath

			if err := services.radarr.AddMovie(requestedMovie); err != nil {
				fmt.Printf("failed to add movie: %v\n", err)
				commandList.showError(channelID, err.Error())
				return
			}

			output := fmt.Sprintf("successfully added `%s (%d)`", requestedMovie.Title, requestedMovie.Year)
			commandList.discord.ChannelMessageSend(channelID, output)
		case "show":
			// use sonarr to add movie
			commandList.showError(channelID, "Sorry, `show` is not implemented!")
		default:
			commandList.showError(channelID, "`movie|show <tmdb-id>`")
		}
	}
}

// I believe radarr only has this feature
func discoverMedia(commandList d, services clients) func(channelID string, args ...string) {
	return func(channelID string, args ...string) {
		argCount := len(args)

		if argCount < 1 {
			commandList.showError(channelID, "need arg `movie|show`")
			return
		}

		mediaType := args[0]

		switch mediaType {
		case "movie":
			movies, err := services.radarr.DiscoverMovies()

			if err != nil {
				output := fmt.Sprintf("fetch movies failed: %v", err)

				fmt.Printf("%v - %s - %s\n", time.Now().String(), channelID, output)

				commandList.showError(channelID, output)
				return
			}

			output := "Here are your recommended movies:\n"

			for _, movie := range movies {
				output += fmt.Sprintf("\t- %s (%d): %s\n", movie.Title, movie.Year, movie.Overview)
			}

			if _, err := commandList.discord.ChannelMessageSend(channelID, output); err != nil {
				fmt.Printf("%v - %s - %v\n", time.Now().String(), channelID, err)
			}
		default:
			output := fmt.Sprintf("unknown media type: %s\nit should be `movie` or `show`", mediaType)

			fmt.Printf("%v - %s - %s", time.Now().String(), channelID, output)

			commandList.showError(channelID, output)
		}
	}
}
