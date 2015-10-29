Meteor.methods({
  media: options => parseCommand(Object.assign(options, {methodType: 'media'})),
  m: options => parseCommand(Object.assign(options, {methodType: 'm'}))
})

function parseCommand (options) {
  check(options, {
    channel_id: String,
    channel_name: String,
    command: String,
    methodType: String,
    team_id: String,
    team_domain: String,
    text: String,
    token: String,
    user_id: String,
    user_name: String,
  })

  let {
    channel_name,
    methodType,
    text,
    token,
    user_name,
  } = options

  let slackToken

  // Determine which method was called to check the associated token
  if (methodType === 'media')
    slackToken = Slack.tokens.media
  else if (methodType === 'm') {
    slackToken = Slack.tokens.m
  }
  else {
    // Crash and burn if none was supplied
    return 'method-type-not-found'
  }

  let result = Api.getDefaultResponse()

  // Verify request came from Slack
  if (!token || !slackToken || token !== slackToken) {
    result.error = 'authorization-failed'

    return result
  }

  let cmd = text

  ////////// Tweak the command for Commander //////////

  // null is needed as Commander.JS strips the first two args
  let fillerForCommander = [null, null]

  // Extract the first arg which should be the command
  // Ex. ['add'], ['remove'] or ['show']
  let command = text.match(/^\w+/)

  // Remove command and the trailing whitespace from string
  cmd = text.replace(command[0] + ' ', '')

  // Use the rest of the string. It should be the title or imdb_id
  // And, pass the username and channel
  // Ex. 'Harry Potter' or 'tt1764'
  let option = {
    media: cmd,
    name: user_name,
    channel: `#${channel_name}`
  }

  // Ex: 
  // [
  //   'add',
  //   {
  //     media: 'Harry Potter',
  //     name: 'bob123',
  //     channel: '#general'
  //   }
  // ]
  command = command.concat(option)

  // Ex: 
  // [
  //    null,
  //    null,
  //    'add',
  //    {
  //      media: 'Harry Potter',
  //      name: 'bob123',
  //      channel: '#general'
  //    }
  // ]
  cmd = fillerForCommander.concat(command)

  // console.log(cmd)

  Commander.parse(cmd)

  // Let the user know their request was successful
  // return 'Success!'
}
