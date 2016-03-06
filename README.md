Shart
===

#####Built with Go
#####Shart stains everywhere!

##Prerequisite

- [A Slack Team](https://slack.com/)
- Slack Incoming WebHook
- Slack Slash Command

####Slack Incoming Webhook
- Create a Slack [incoming webhook](https://slack.com/services/new/incoming-webhook)
- Choose a channel (Doesn't matter)
- Copy the `WebHook URL`
- Save integration

Save the `url` somewhere safe as you need it when you start up shart

####Slack Slash Command
- Create a Slack [slash command](https://slack.com/services/new/slash-commands): `m`
- Point the `url` to: `http://<your-ip-address>:4040/v1/m`
- Keep the `method` as `POST`
- Copy the `token`
- Save the integration

Save the `token` somewhere safe as you need it when you start up shart

###Installation

If you have [Go](https://golang.org/) installed you can `go get github.com/jrudio/shart`

then

`go install github.com/jrudio/shart`

###Alternatively
You can download the [pre-compiled version in the releases section](https://github.com/jrudio/shart/releases) of this repo.

####Only tested on Mac OS X 10.10.5 (Yosemite)

###Start Up

To sucessfully startup shart, the following flags are required:

- `slack-token`
- `slack-url`
- `couchpotato-url`
- `couchpotato-apikey`

The following are optional:

- `bot-name` (Default is `MediaBot`)
- `host` (Default listens on port `4040`)

```bash
shart -couchpotato-url "192.168.1.5:5050" -couchpotato-api "abc1234" -slack-token "abc1234" \
  -slack-url "https://hooks.slack.com/services/xxxxx/xxxx/xxxx" -bot-name "ShartBot" \
  -host ":4040"
```

####CouchPotato is the only server implemented right now

What Currently Works:

- `search <title>`
- `show wanted`
- `add <imdb_id>`
- `remove <media_id>`
- `test` (Tests connection to CouchPotato)


Planned Commands:

- `show <tmdb_id>`

####To add a movie:

1. `/m search <movie>`

  This will display the results along with it's `imdb_id`

2. `/m add <imdb_id>`

  It will then respond saying it was added successfully or that it failed.

####To remove a movie from your wanted list:

1. `/m show wanted`

  This will show your wanted list along with the necessary `media_id` needed to remove it

2. `/m remove <media_id>`

  Then a response of success or failed will be displayed
