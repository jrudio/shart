_.extend(Meteor, {
  /**
  * Checks that the settings are present and returns them
  * @param {String} setting - The requested setting key
  * return {Object|Boolean}
  */
  checkSettings (setting) {
    check(setting, String)

    try {
      return (Meteor.settings && Meteor.settings[setting]) || (Meteor.settings && Meteor.settings)
    } catch (e) {
      console.error(e.message)

      return false
    }
  }
})