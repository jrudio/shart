Api.Url = class Url {
  constructor (host) {
    // The constructor will check whether an end-slash is needed
    check(host, String)

    this.baseUrl = this.modifyUrlForEndSlash(host)

    // Good to go!
    return this
  }
  modifyUrlForEndSlash (host) {
    let regex = {
      endSlash: /\/$/
    }

    let hasEndSlash = regex.endSlash.test(host)
    let baseUrl = host

    // Determine whether to add a slash or not before adding the api key
    if (!hasEndSlash) baseUrl += '/'

    return baseUrl
  }
  fetch (query) {
    return new Promise( (resolve, reject) => {
      HTTP.get(query, (error, result) => {
        if (error)
          return reject(error)
        else
          return resolve(result.data)
      })
    })

    // return fetch(query).then(result => result.json())
  }
}
