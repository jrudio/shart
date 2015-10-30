Shart
====

Control your CouchPotao, Sonarr, and Plex servers via Slack

##Prerequisites
- [Meteor](https://www.meteor.com)
- [A Slack team](https://slack.com/)

##To Get Started

#####Slack
Slack Incoming WebHook:

- Create a Slack [incoming webhook](https://slack.com/services/new/incoming-webhook)
- Choose a channel (Doesn't matter)
- Copy the `WebHook URL`
- Save integration
- Paste the `WebHook URL` into this project's `example_settings.json` under `slack.hooks.incoming`

Slack commands:

- Create a Slack [slash command](https://slack.com/services/new/slash-commands): `media`
- Point the `url` to: `http://<your-ip-address>:3000/methods/media`
- Keep the `method` as `POST`
- Copy the `token`
- Save the integration
- Paste the `token` into `slack.tokens.media`

If you want the shorthand command `/m` as well, do the above again, but change the `url` to:

`http://<your-ip-address>:3000/methods/m`

and the `token` destination to `slack.tokens.m`

######Meteor - Development version

- Make sure you change the appropriate variables in `example_settings.json`
- Start up meteor: `meteor --settings <path/to/settings.json>`

######Meteor - Production version

***TODO: Finish this section!***

- `npm i -g mup`
- Create a new folder (outside of this project) to init mup

#####Test the Meteor server -> CouchPotato connection*

- In your `example_settings.json`, change `couchpotato.apiKey` & `couchpotato.host` to match your CouchPotato's `apiKey` & `host`

- In Slack type: `/media show test`

- You should get a response back like in the example below

**\*This is the only test command I have setup at the moment**

*Files and folders that need to be loaded first are prepended with an underscore*

######A simple CouchPotato connection status check
![couch query](.gifs/couch_query.gif)
