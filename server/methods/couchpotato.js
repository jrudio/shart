let settings = Meteor.checkSettings('couchpotato')

_.extend(CouchPotato, {
  apiKey: settings.apiKey,
  host: settings.host,
  Format: {}
})

// User-facing strings for results from api calls
_.extend(CouchPotato.Format, {
  test (result) {
    check(result, {
      success: Boolean
    })

    let {success} = result

    let txt = 'CouchPotato connection test'

    if (success === true)
      txt = `:white_check_mark:\t${txt} successful!`
    else
      txt = `:exclamation:\t${txt} failed!`

    return txt
  },
  charts (result) {
    check(result, {
      count: Number,
      ignored: Array,
      charts: Array,
      success: Boolean
    })

    const {
      count,
      ignore,
      charts,
      success
    } = result

    let txt = ''

    if (!count)
      return `:x:\tNo charts returned!`

    // The response text should show up in Slack as such (ignore the forward slashes):
    ////////////////////////////////
    // 
    // [Chart Name]
    //  [] Title - Year
    //  [] Title - Year
    //  [] Title - Year
    //  [] Title - Year
    // [Chart Name]
    //  [] Title - Year
    //  [] Title - Year
    //  [] Title - Year
    //  [] Title - Year
    // 
    ////////////////////////////////
    // 
    // [Blu-ray.com - New Releases]
    //  [] Harry Potter - 2005
    //  [] The Hobbit - 2011
    //  [] Interstellar - 2014
    // 
    ////////////////////////////////

    // Iterate through the charts prop to construct the response text
    charts.forEach(_chart => {
      txt += `*${_chart.name}*\n`

      _chart.list.forEach(_list => {
        let statusBox = ':black_small_square:'
        let {info} = _list

        // TODO: Clean up the commented code below

        // The following commented code is irrelevant as any media that is
        // already downloaded or in the wanted list will not be present in the charts

        // let mediaStatus = Api.CouchPotato.getStatus(info.in_library, info.in_wanted)

        // switch (mediaStatus) {
        //   case 'wanted':
        //     statusBox = ':clipboard:'
        //     break
        //   case 'downloaded':
        //     statusBox = ':white_check_mark:'
        //     break
        //   default:
        //     statusBox = ':black_small_square:'
        // }

        txt += `${statusBox}\t${_list.title} - ${info.year}\n`
      })
    })

    return txt
  }
})