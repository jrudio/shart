let settings = Meteor.checkSettings('couchpotato')

_.extend(CouchPotato, {
  apiKey: settings.apiKey,
  host: settings.host,
  Format: {}
})

// User-facing strings for results from api calls
_.extend(CouchPotato.Format, {
  test (result) {
    let txt = 'CouchPotato connection test'

    if (result === true)
      txt = `:white_check_mark:\t${txt} successful!`
    else
      txt = `:exclamation:\t${txt} failed!`

    return txt
  }
})