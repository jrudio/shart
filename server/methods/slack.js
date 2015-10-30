let settings = Meteor.checkSettings('slack')

_.extend(Slack, {
  tokens: {
    media: settings.tokens.media,
    m: settings.tokens.m,
  },
  hooks: {
    incoming: settings.hooks.incoming
  },
  respondAsMediaBot (text, channel, cb) {
    let url = Slack.hooks.incoming

    let payload = {
      channel,
      text,
      username: 'MediaBot'
    }

    payload = JSON.stringify(payload)

    let params = {payload}

    HTTP.post(url, {
      params
    }, cb)
  }
})
