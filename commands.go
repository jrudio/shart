package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseCMD gets first chunk of text; that should be the command
// Execute appropriate function
// TODO: Method type may be a problem
func ParseCMD(text string, srvr server) string {
	// Lowercase that shit
	text = strings.ToLower(text)

	// The text will be parsed as (search)( )(lord of the rings)
	cmdRegex := regexp.MustCompile(`(\w+)(\s)?([\w\s]+)?`)

	parsedText := cmdRegex.FindStringSubmatch(text)

	if len(parsedText) == 0 {
		return "Error: Please provide a command and an argument (e.g. search <movie-title>)"
	}

	// First parenthesis should be command
	cmd := parsedText[1]

	// Third group of parenthesis should be the media title or imdb_id
	args := parsedText[3]

	if cmd != "test" && args == "" {
		return "Error: Please enter a title or id. (e.g. search Interstellar)"
	}

	var formattedText string

	switch cmd {
	// Add
	case "add":
		formattedText = srvr.AddMovieToWanted(args)

	// Show
	case "show":
		if args == "wanted" {
			// The user wants to display the wanted list
			list, listErr := srvr.ShowWanted("", "")

			if listErr != nil {
				formattedText = listErr.Error()
			} else {
				// Format the list for Slack
				formattedText = formatWanted(list)
			}
		} else {
			// TODO: Implement showing individual media with
			// expanded information
			formattedText = fmt.Sprintf("Showing %v\n", args)
		}

	// Remove
	case "remove":
		formattedText = srvr.RemoveMovieFromWanted(args)

	// Search
	case "search":
		txt := srvr.Search(args)

		// Format that result for Slack
		formattedText = formatSearch(args, txt)
	case "test":
		test := srvr.TestConnection()

		formattedText = "Connection to CouchPotato "

		if test {
			formattedText += "worked"
		} else {
			formattedText += "failed"
		}
	default:
		formattedText = "Command not recognized"
	}

	return formattedText
}

/* Format messages for Slack */

////////////////////////////////////
// Should look like
////////////////////////////////////
// @username has added <title> (<year>)
////////////////////////////////////
// func AddMedia(media map[string]string) string {
//   username := media["username"]
//   title := media["title"]
//   year := media["year"]

//   formattedText := "@" + title + " has added " + title + " (" + year + ")"

//   return formattedText
// }

////////////////////////////////////
// Searched For: *<title>*
//
// [] <tmdb_id> <title> - <year>:
//    <plot>
// [] <tmdb_id> <title> - <year>:
//    <plot>
// [] <tmdb_id> <title> - <year>:
//    <plot>
// [] <tmdb_id> <title> - <year>:
//    <plot>
// [] <tmdb_id> <title> - <year>:
//    <plot>
////////////////////////////////////
func formatSearch(title string, result []map[string]string) string {
	var f = "\nSearched for: *" + title + "*\n"

	for ii := 0; ii < len(result); ii++ {
		// "m" for movie ya bish
		m := result[ii]

		id := m["id"]

		if id == "{}" {
			id = "no_id"
		}

		f += ":black_small_square:\t*ID: " + id + "* * " + m["title"] + "* - *" + m["year"] +
			"*:\n\t" + m["plot"] + "\n"
	}

	return f
}

////////////////////////////////////
// Showing <count> movies in your wanted list:
// 		ID: <media_id>
//
// [] <title> - <year>:
// 		ID: <media_id>
// 		<plot>
// [] <title> - <year>:
// 		ID: <media_id>
// 		<plot>
// [] <title> - <year>:
// 		ID: <media_id>
// 		<plot>
// [] <title> - <year>:
// 		ID: <media_id>
// 		<plot>
// [] <title> - <year>:
// 		<plot>
//
////////////////////////////////////

func formatWanted(list wantedList) string {
	movieCount := strconv.Itoa(list.Total)

	formattedText := "Showing *" + movieCount + "* movies from your wanted list:"

	// Newline
	formattedText += "\n"

	for _, movie := range list.Movies {
		// Add the bullet point emoji-shit
		formattedText += ":black_small_square:\t"

		// Title
		formattedText += "*" + movie.Title + " "

		// Year
		formattedText += "(" + strconv.Itoa(movie.Info.Year) + ") "

		// Divider
		formattedText += " - "

		// Media id
		formattedText += movie.MediaID + "* "

		// // Newline and Tab
		formattedText += ":\n\t"

		// Plot
		formattedText += movie.Info.Plot

		// Newline
		formattedText += "\n"
	}

	return formattedText
}
