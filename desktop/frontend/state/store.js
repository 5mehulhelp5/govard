const state = {
  environments: [],
  selectedProject: "",
  selectedService: "all",
  selectedSeverity: "all",
  logQuery: "",
  liveLogsEnabled: false,
  syncConfig: {
    sanitize: true,
    excludeLogs: true,
    compress: false,
  },
};

export const getState = () => state;

export const setState = (patch) => {
  Object.assign(state, patch || {});
  return state;
};
