Api.CouchPotato = class _CouchPotato extends Api.Url {
  constructor () {
    let {host} = CouchPotato

    super(host)

    this.apiKey = CouchPotato.apiKey

    // Tweak base url for CouchPotato
    this.buildApiUrl()

    return this
  }
  buildApiUrl () {
    if (this.baseUrl && this.apiKey)
      this.baseUrl += 'api/' + this.apiKey + '/'

    return this
  }
  static getStatus (inLibrary, inWanted) {
    let status

    if (!inLibrary && !inWanted) {
      // In neither library or wanted

      status = 'notFound'

    } else if (inLibrary && !inWanted) {
      // In Library

      status = 'downloaded'

    } else if (!inLibrary && inWanted && inWanted.status === 'active') {
      // Waiting on the wanted list

      status = 'wanted'
    }

    return status
  }
  hasSettings () {
    return this.apiKey && this.baseUrl
  }
  getMediaById (id) {
    if (typeof id !== 'string') throw new Error('id-not-string')

    if (!this.hasSettings) throw new Error('insufficient-settings')

    let query = this.baseUrl + 'media.get/?id=' + id

    return this.fetch(query)
  }
  get wantedList () {
    if (!this.hasSettings) throw new Error('insufficient-settings')

    let query = this.baseUrl + 'media.list/?status=active'

    return this.fetch(query)
  }
  get charts () {
    if (!this.hasSettings) throw new Error('insufficient-settings')

    let query = this.baseUrl + 'charts.view'

    return this.fetch(query)
  }
  get isAvailable () {
    if (!this.hasSettings) throw new Error('insufficient-settings')

    let query = this.baseUrl + 'app.available'

    return this.fetch(query)
  }
  search (title) {
    if (!this.hasSettings) throw new Error('insufficient-settings')

    let query = this.baseUrl + 'search/?q=' + encodeURIComponent(title)

    return this.fetch(query)
  }
  addToWanted (options) {
    let {
      profile_id,
      title,
      identifier,
      category_id,
      force_readd
    } = options

    if (!title || !identifier) throw new Error('no-title-or-identifier-supplied')

    if (!this.hasSettings) throw new Error('insufficient-settings')

    let url = this.baseUrl + 'movie.add/'

    url += identifier && '?identifier=' + identifier
    url += title && '&title=' + title

    return this.fetch(url)
  }
  removeFromWanted (id) {
    if (typeof id !== 'string') throw new Error('no-id-supplied')

    if (!this.hasSettings) throw new Error('insufficient-settings')

    let url = this.baseUrl + 'movie.delete/'

    url += '?id=' + id
    url += '&delete_from=wanted'

    return this.fetch(url)
      .then(r => r.success)
  }
}
