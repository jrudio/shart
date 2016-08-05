package main

import (
	"strconv"
)

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
