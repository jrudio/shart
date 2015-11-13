package commands

import (
	"fmt"
	"github.com/jrudio/shart/api"
	"strings"
)

// Get first chunk of text; that should be the command
// Execute appropriate function
// TODO: Method type may be a problem
func ParseCmd(text string, server *serverApi.Server) string {
	// Lowercase that shit
	text = strings.ToLower(text)

	// First word should be the command
	cmd := strings.Fields(text)[0]

	// Remove that cmd from string and the extra whitespace that follows it
	text = strings.Replace(text, cmd+" ", "", 1)

	// fmt.Println("Cmd:", cmd)
	// fmt.Println("Text:", text)

	var formattedText string = ""

	switch cmd {
	// Add
	// TODO: Implement this command
	case "add":
		fmt.Printf("Adding %v\n", text)
		// formattedText = AddMedia()

	// Show
	// TODO: Implement this command
	case "show":
		fmt.Printf("Showing %v\n", text)

	// Delete
	// TODO: Implement this command
	case "delete":
		fmt.Printf("Deleting %v\n", text)

	// Search
	case "search":
		fmt.Printf("Searching for %v\n", text)
		// Make the api call
		// server.Search(text) // text === media title
		txt := server.Search(text) // text === media title

		// Format that result for Slack
		formattedText = formatSearch(text, txt)

	default:
		fmt.Println("Command not recognized")
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
	var f string = "\nSearched for: *" + title + "*\n"

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
