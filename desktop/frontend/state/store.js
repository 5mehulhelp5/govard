const state = {
  sidebarMode: "global-services",
  environments: [],
  selectedProject: "",
  selectedService: "all",
  selectedSeverity: "all",
  logQuery: "",
  globalServices: [],
  selectedGlobalService: "caddy",
  globalLogSeverity: "all",
  globalLogQuery: "",
  liveLogsEnabled: false,
  globalLiveLogsEnabled: false,
  terminalModalOpen: false,
  syncConfigs: {},
};

export const getState = () => state;

export const setState = (patch) => {
  Object.assign(state, patch || {});
  return state;
};
