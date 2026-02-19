const state = {
  environments: [],
  selectedProject: "",
  selectedService: "web",
  liveLogsEnabled: false,
}

export const getState = () => state

export const setState = (patch) => {
  Object.assign(state, patch || {})
  return state
}

