package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	radarr "github.com/jrudio/go-radarr-client"
	sonarr "github.com/jrudio/go-sonarr-client"
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

func clearMessages(commandList d, services clients) func(channelID string, args ...string) {
	return func(channelID string, args ...string) {
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
			commandList.showError(channelID, err.Error())
		}
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
					// can't display movie summary because of Discord's 2000 character limit
					formattedResults += "- " + result.Title + " (" + strconv.Itoa(result.Year) + ") `" + strconv.Itoa(result.TmdbID) + "`\n"
				}
			}

			commandList.discord.ChannelMessageSend(channelID, formattedResults)
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
					// can't display movie summary because of Discord's 2000 character limit
					formattedResults += "- " + result.Title + " (" + strconv.Itoa(result.Year) + ") `" + strconv.Itoa(result.TvdbID) + "`\n"
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
			profiles, err := services.sonarr.GetProfiles()

			if err != nil {
				errMsg := fmt.Sprintf("failed to fetch profiles from sonarr: %v\n", err)
				fmt.Printf(errMsg)
				commandList.showError(channelID, errMsg)
				return
			}

			output := "Here are the available quality profiles for sonarr:\n"

			for _, profile := range profiles {
				output += fmt.Sprintf("\t`id: %d` %s\n", profile.ID, profile.Name)
			}

			if _, err := commandList.discord.ChannelMessageSend(channelID, output); err != nil {
				fmt.Printf("chan id: %s - %v\n", channelID, err)
				return
			}

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
		case "show":
			defaultSonarrQualityID = profileID
			output := "successfully set series quality to `%d`"
			commandList.discord.ChannelMessageSend(channelID, fmt.Sprintf(output, defaultSonarrQualityID))
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
			folders, err := services.sonarr.GetRootFolders()

			if err != nil {
				errMsg := fmt.Sprintf("failed to fetch folders from sonarr: %v\n", err)
				fmt.Printf(errMsg)
				commandList.showError(channelID, errMsg)
				return
			}

			output := "Here are the available root folders for sonarr:\n"

			for _, folder := range folders {
				output += fmt.Sprintf("\t`id: %d` - %s\n", folder.ID, folder.Path)
			}

			if _, err := commandList.discord.ChannelMessageSend(channelID, output); err != nil {
				fmt.Printf("chan id: %s - %v\n", channelID, err)
				return
			}
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
		case "show":
			if strings.HasPrefix(folderPathOrID, "/") {
				defaultSonarrPath = folderPathOrID
			} else {
				pathID, err := strconv.Atoi(folderPathOrID)

				if err != nil {
					output := fmt.Sprintf("failed to convert path id to int: %v", err)
					fmt.Printf("channel id: %s - %s\n", channelID, output)
					commandList.showError(channelID, output)
					return
				}

				folders, err := services.sonarr.GetRootFolders()

				if err != nil {
					output := fmt.Sprintf("fetch sonarr root folders failed: %v", err)
					fmt.Printf("channel id: %s - %s\n", channelID, output)
					commandList.showError(channelID, output)
					return
				}

				for _, folder := range folders {
					if folder.ID == pathID {
						defaultSonarrPath = folder.Path
					}
				}

				if defaultSonarrPath == "" {
					output := "could not find stored path via id: `%s`"
					commandList.showError(channelID, fmt.Sprintf(output, folderPathOrID))
					return
				}
			}

			output := "successfully set root folder to `%s`"
			commandList.discord.ChannelMessageSend(channelID, fmt.Sprintf(output, defaultSonarrPath))

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
		mediaID := args[1]

		if mediaID == "" {
			commandList.showError(channelID, "a tmdb/tvdb id is required\n `movie|show <id>`")
			return
		}

		switch mediaType {
		case "movie":
			// use radarr to add movie
			tmdbID, err := strconv.Atoi(mediaID)

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

			if errors := services.radarr.AddMovie(requestedMovie); errors != nil {
				output := ""
				logOutput := ""

				for _, err := range errors {
					if err == radarr.ErrorMovieExists {
						output += "`" + requestedMovie.Title + " (" + strconv.Itoa(requestedMovie.Year) + ")` is already added"
					}

					logOutput += err.Error() + "\n"

				}

				fmt.Printf("failed to add movie - channel id: %s - %s\n", channelID, logOutput)
				commandList.discord.ChannelMessageSend(channelID, output)
				return
			}

			output := fmt.Sprintf("successfully added `%s (%d)`", requestedMovie.Title, requestedMovie.Year)
			commandList.discord.ChannelMessageSend(channelID, output)
		case "show":
			// use sonarr to add movie
			tvdbID, err := strconv.Atoi(mediaID)

			if err != nil {
				output := fmt.Sprintf("failed to convert tvdb id to int: %v", err)
				fmt.Printf("channel id: %s - %s\n", channelID, output)
				commandList.showError(channelID, output)
				return
			}

			// make sure profile quality and folder path are set
			if defaultSonarrPath == "" {
				commandList.showError(channelID, "aborting... a root folder path must be set")
				commandList.showHelp(channelID)
				return
			}

			if defaultSonarrQualityID == 0 {
				commandList.showError(channelID, "aborting... a profile quality must be set")
				commandList.showHelp(channelID)
				return
			}

			// commandList.showError(channelID, "Sorry, `show` is not implemented!")

			requestedShow, err := services.sonarr.GetSeriesFromTVDB(tvdbID)

			if err != nil {
				fmt.Printf("failed to add show: %v\n", err)
				commandList.showError(channelID, fmt.Sprintf("failed fetching show: %v", err))
				return
			}

			// tweak fields to make a proper request
			requestedShow.AddOptions.SearchForMissingEpisodes = true
			requestedShow.Monitored = true
			requestedShow.QualityProfileID = defaultSonarrQualityID
			requestedShow.Path = defaultSonarrPath + requestedShow.Title

			if errors := services.sonarr.AddSeries(*requestedShow); errors != nil {
				output := ""
				logOutput := ""

				for _, err := range errors {
					if err == sonarr.ErrorSeriesExists {
						output += "`" + requestedShow.Title + " (" + strconv.Itoa(requestedShow.Year) + ")` is already added"
					}

					logOutput += err.Error() + "\n"
				}

				fmt.Printf("failed to add show - channel id: %s - %s\n", channelID, logOutput)
				commandList.discord.ChannelMessageSend(channelID, output)
				return
			}

			output := fmt.Sprintf("successfully added `%s (%d)`", requestedShow.Title, requestedShow.Year)
			commandList.discord.ChannelMessageSend(channelID, output)

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

func showLibrary(commandList d, services clients) func(channelID string, args ...string) {
	return func(channelID string, args ...string) {
		// command: library <movie|show> [missing, downloaded, or missing] <page-number>
		// page number is optional -- w/o page number we'll show the first page of results
		//
		// examples:
		// library movie
		// library movie 3
		// library movie missing
		// library movie missing 3
		// library movie downloaded 6

		argCount := len(args)

		if argCount < 1 {
			commandList.showError(channelID, "need arg `movie|show`")
			return
		}

		mediaType := args[0]

		args = args[1:argCount]
		argCount--

		// args should be: "", "1" (page), "monitored" (a filter type), "monitored 2" (a filter type + page number)

		page := "1"
		pageSize := "40"

		switch mediaType {
		case "movie":
			options := radarr.GetMovieOptions{
				Page:     page,
				PageSize: pageSize,
				SortKey:  "sortTitle",
				SortDir:  "asc",
			}

			if argCount > 0 {
				// check for a page number
				// if successful there's no filter
				if _, err := strconv.Atoi(args[0]); err != nil {
					// we could not convert so we most likely have a filter
					filter := args[0]
					filterValue := "true"

					switch filter {
					case "monitored":
						filter = "monitored"
					case "downloaded":
						filter = "downloaded"
					case "missing":
						filter = "downloaded"
						filterValue = "false"
					case "released":
						filter = "status"
						filterValue = "released"
					case "announced":
						filter = "status"
						filterValue = "announced"
					case "cinemas":
						filter = "status"
						filterValue = "inCinemas"
					default:
						commandList.showError(channelID, fmt.Sprintf("unknown filter `%s` for command `library movie`", filter))
						return
					}

					options.FilterKey = filter
					options.FilterValue = filterValue
					options.FilterType = "equal"

					// check for page number
					if argCount > 1 {
						if _, err := strconv.Atoi(args[1]); err == nil {
							page = args[1]
							options.Page = page
						}
					}

				} else {
					// we converted the argument to a number
					// so we have a page number
					page = args[0]
					options.Page = page
				}
			}

			movies, err := services.radarr.GetMovies(options)

			if err != nil {
				output := fmt.Sprintf("fetch movies from radarr failed: %v", err)

				commandList.showError(channelID, output)
				logPrint(channelID, output)
				return
			}

			movieCount := len(movies)
			// the output preface has 35 chars
			output := fmt.Sprintf("showing %d movies on page %s:\n\n",
				movieCount, page)
			titleLen := 0
			yearLen := 0

			// no movies but there is a page argument
			if movieCount < 1 && page != "" {
				output += "uh oh! try going back a page!"
			} else if movieCount < 1 && page == "" {
				output += "add some movies to your library! :smile:"
			}

			for _, movie := range movies {
				yearStr := strconv.Itoa(movie.Year)

				// '<title> (2000) - downloaded\n'
				//  <-x-->|<------26 chars ---->|
				//
				// x == title; average char length is 14
				// if not downloaded subtract 13 chars
				// if downloaded total would be 82
				// not downloaded total is 69 for each movie
				//
				// we can average ~50 movies before we need to go to the next page
				output += movie.Title + " (" + yearStr + ") "

				if movie.Downloaded {
					output += " - `downloaded`"
				}

				output += "\n"

				titleLen += len(movie.Title)
				yearLen += len(yearStr)
			}

			if isVerbose {
				// get the average movie + year length to determine how many movies we can show in discord
				// without going over the 2000 char limit
				fmt.Printf("movie count: %d\n", movieCount)
				fmt.Printf("\ttotal title length: %d\n\ttotal year length: %d\n", titleLen, yearLen)

				averageTitleLen := 0
				averageYearLen := 0

				if movieCount != 0 {
					averageTitleLen = titleLen / movieCount
					averageYearLen = yearLen / movieCount
				}

				fmt.Printf("\taverage title length: %d\n\taverage year length: %d\n", averageTitleLen, averageYearLen)
				fmt.Printf("\ttotal message length: %d\n", len(output))
			}

			if _, err := commandList.discord.ChannelMessageSend(channelID, output); err != nil {
				fmt.Printf("message sent to discord failed: %v\n", err)
				commandList.discord.ChannelMessageSend(channelID, fmt.Sprintf("could not reply back: %v", err))
			}
		case "show":
			output := "`library show` not implemented"
			if _, err := commandList.discord.ChannelMessageSend(channelID, output); err != nil {
				fmt.Printf("message sent to discord failed: %v\n", err)
				commandList.discord.ChannelMessageSend(channelID, fmt.Sprintf("could not reply back: %v", err))
			}
		default:
			output := "unknown command"

			logPrint(channelID, output)
		}
	}
}
