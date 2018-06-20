Shart
===

Use Discord to manage your Radarr and Sonarr applications

Commands:

- `search <title>` (for new media)
- `clear` (remove messages if there's too much clutter)
- `add <tmdb-id>` to be monitored
- `quality` to retrieve avilable quality profiles
- `folders` to retrieve avilable root folders
- `set-quality <profile-id>` to set quality profile to make a valid add request
- `set-folder <folder-path-or-id>` to set folder path make a valid add request


Install
===

- Install the Go compiler
- Install go [dep](https://github.com/golang/dep)
- clone this project
- run `dep ensure` in project folder
- run `go build -o shart`
- run `./shart -token <discord-token> -radarr-url http://192.168.1.15:7878 -radarr-key abc123 -sonarr-url http://192.168.1.15:8989 -sonarr-key abc123`


To get a discord token go to `https://discordapp.com/developers/applications/me` 
- click `New App`
- fill out required information
- click save
- click `create a bot user`
- click on `generate oauth2 url`
- check `send messages` and `manage messages`
- copy and go to url
- authorize bot to access your discord server
- go back to `https://discordapp.com/developers/applications/me` 
- click on `token` to retrieve discord token

Usage
===

This bot will respond to the trigger word `shart`

if you would like to search for movies use: `shart search movie sicario`

the bot will respond with: 

```
Here are your search results for sicario:
    - Sicario 2015 (273481)
    - Sicario 1994 (95700)
    - Sicario: Day of the Soldado 2018 (400535)
```

use the id in parenthesis to add that movie

`shart add movie 400535`

you must set a default quality profile id and root folder path for both radarr and sonarr

`shart set-quality movie 3`

`shart set-quality show 4`

`shart set-folder movie 3` or `shart set-folder movie /home/user1/movies`

`shart set-folder show 2` or `shart set-folder show /home/user1/shows`

otherwise you will get both of these errors:

`aborting... a root folder path must be set`

`aborting... a profile quality must be set`

once you set those adding a movie will give you a success message: `successfully added Sicario: Day of the Soldado - (2018)`
