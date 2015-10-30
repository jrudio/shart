Api.DEFAULT_RESPONSE = {
  error: null,
  message: null,
  success: false
}

Api.getDefaultResponse = () => {
  let r = {}

  _.extend(r, Api.DEFAULT_RESPONSE)

  return r
}