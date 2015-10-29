Meteor.startup( () => {
  Slack = {}

  let settings = Helpers.checkSettings('slack')

  _.extend(Slack, {
    tokens: {
      media: settings.tokens.media,
      m: settings.tokens.m,
    },
    hooks: {
      incoming: settings.hooks.incoming
    },
    respondAsMediaBot (text, channel, cb) {
      // console.log(text)

      let url = Slack.hooks.incoming

      let payload = {
        channel,
        text,
        username: 'Media'
      }

      payload = JSON.stringify(payload)

      let params = {payload}

      HTTP.post(url, {
        params
      }, cb)
    }
  })
})
