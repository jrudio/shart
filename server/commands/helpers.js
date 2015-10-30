_.extend(CommandHelpers, {
  /**
  * @param {String} Media requested
  * returns {Promise}
  */
  parseShow (media) {
    let result

    let couchpotato = new Api.CouchPotato()

    // console.log(couchpotato)

    // Determine what to show:
    // - individual media
    // - wanted list
    // - categories; recently released, in theaters, etc.
    switch (media) {
      case 'wanted':
        result = couchpotato.wantedList
        break
      case 'charts':
        result = couchpotato.charts
          .then(r => CouchPotato.Format.charts(r))
        break
      case 'test':
        result = couchpotato.isAvailable
          .then(r => CouchPotato.Format.test(r))
        break
      // Show individual media
      default:
        console.log('\"Show individual media\" is not implemented')
        result = new Promise((resolve, reject) => reject('not-implemented'))
    }

    return result
  }
})