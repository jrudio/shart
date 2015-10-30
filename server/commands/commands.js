// Load commander.js
Commander = Meteor.npmRequire('commander')

// React to request
Commander
  .command('add <media>')
  .action(options => {
    check(options, {
      media: String,
      name: String,
      channel: String
    })

    let {
      media,
      name,
      channel
    } = options

    // Construct the user-readable message
    let msg = `@${name} has added ${media}`

    Slack.respondAsMediaBot(msg, channel)
  })

Commander
  .command('remove <media>')
  .action(options => {
    check(options, {
      media: String,
      name: String,
      channel: String
    })

    let {
      media,
      name,
      channel
    } = options

    // Construct the user-readable message
    let msg = `@${name} has removed ${media}`

    Slack.respondAsMediaBot(msg, channel)
  })
  
Commander
  .command('show <media>')
  .action(options => {
    check(options, {
      media: String,
      name: String,
      channel: String
    })

    let {
      media,
      name,
      channel
    } = options

    // Determine whether to show individual media, a list of media, etc
    CommandHelpers.parseShow(media)
      .then(text => {
        console.log('Response:')
        console.log(text)
        Slack.respondAsMediaBot(text, channel)
      })
      .catch(e => console.error(e.message))
  })