import { createActionsController } from "./modules/actions.js"
import {
  normalizeDashboardPayload,
  projectKey,
  renderEnvironmentList,
  renderWarnings,
  setMetricText,
  syncProjectSelectors,
} from "./modules/dashboard.js"
import { createLogsController, resolveLogTarget, syncServiceSelector } from "./modules/logs.js"
import { createMetricsController } from "./modules/metrics.js"
import { createOnboardingController } from "./modules/onboarding.js"
import { createRemotesController } from "./modules/remotes.js"
import { createSettingsController } from "./modules/settings.js"
import { createShellController } from "./modules/shell.js"
import { desktopBridge } from "./services/bridge.js"
import { getState, setState } from "./state/store.js"
import { createToast } from "./ui/toast.js"
import { byId, setText } from "./utils/dom.js"

const refs = {
  status: byId("status"),
  refresh: byId("refresh"),
  statActive: byId("statActive"),
  statServices: byId("statServices"),
  statQueue: byId("statQueue"),
  statActiveHint: byId("statActiveHint"),
  statServicesHint: byId("statServicesHint"),
  statQueueHint: byId("statQueueHint"),
  metricActiveProjects: byId("metricActiveProjects"),
  metricCPU: byId("metricCPU"),
  metricMemory: byId("metricMemory"),
  metricNetRx: byId("metricNetRx"),
  metricNetTx: byId("metricNetTx"),
  metricOOM: byId("metricOOM"),
  metricsList: byId("metricsList"),
  metricsWarnings: byId("metricsWarnings"),
  remotesList: byId("remotesList"),
  remotesWarnings: byId("remotesWarnings"),
  projectPath: byId("projectPath"),
  projectRecipe: byId("projectRecipe"),
  remoteName: byId("remoteName"),
  remoteHost: byId("remoteHost"),
  remoteUser: byId("remoteUser"),
  remotePath: byId("remotePath"),
  remotePort: byId("remotePort"),
  remoteEnvironment: byId("remoteEnvironment"),
  remoteCapabilities: byId("remoteCapabilities"),
  remoteAuthMethod: byId("remoteAuthMethod"),
  remoteProtected: byId("remoteProtected"),
  envList: byId("envList"),
  envSelector: byId("envSelector"),
  logSelector: byId("logSelector"),
  logServiceSelector: byId("logServiceSelector"),
  logSeverity: byId("logSeverity"),
  logSearch: byId("logSearch"),
  logOutput: byId("logOutput"),
  toggleLive: byId("toggleLive"),
  shellUser: byId("shellUser"),
  shellCommand: byId("shellCommand"),
  warningList: byId("warningList"),
  openSettings: byId("openSettings"),
  closeSettings: byId("closeSettings"),
  settingsDrawer: byId("settingsDrawer"),
  themeSelect: byId("themeSelect"),
  proxyTarget: byId("proxyTarget"),
  preferredBrowser: byId("preferredBrowser"),
  toastContainer: byId("toastContainer"),
}

const toast = createToast(refs.toastContainer)

const setStatus = (message) => {
  setText(refs.status, message)
}

const showToast = (message, type = "success") => {
  toast.show(message, type)
}

const showSystemNotification = (title, body) => {
  if (typeof window === "undefined" || typeof window.Notification === "undefined") {
    return
  }
  if (window.Notification.permission === "granted") {
    new window.Notification(title, { body })
    return
  }
  if (window.Notification.permission === "default") {
    window.Notification.requestPermission().then((permission) => {
      if (permission === "granted") {
        new window.Notification(title, { body })
      }
    })
  }
}

if (desktopBridge.runtime?.EventsOn) {
  desktopBridge.runtime.EventsOn("operations:notification", (payload = {}) => {
    const title = String(payload.title || "Govard operation update")
    const body = String(payload.body || "").trim()
    const level = payload.level === "error" ? "error" : "success"
    showToast(body || title, level)
    showSystemNotification(title, body || title)
  })
}

const readSelection = () =>
  resolveLogTarget({
    project: getState().selectedProject,
    service: getState().selectedService,
    severity: getState().selectedSeverity,
    query: getState().logQuery,
  })

const safeDashboard = {
  ActiveEnvironments: 0,
  RunningServices: 0,
  QueuedTasks: 0,
  ActiveSummary: "No environments detected",
  ServicesSummary: "Desktop bridge unavailable",
  QueueSummary: "Queue idle",
  Environments: [],
  Warnings: ["Desktop bridge unavailable. Showing local fallback view."],
}

const loadDashboard = async () => {
  try {
    const data = await desktopBridge.getDashboard()
    return normalizeDashboardPayload(data)
  } catch (_err) {
    return normalizeDashboardPayload(safeDashboard)
  }
}

const syncProjectState = () => {
  const selectedProject = refs.envSelector?.value || refs.logSelector?.value || ""
  setState({ selectedProject })
}

const syncServiceState = () => {
  const selectedService = refs.logServiceSelector?.value || "all"
  setState({ selectedService })
}

const syncLogFiltersState = () => {
  const selectedSeverity = refs.logSeverity?.value || "all"
  const logQuery = refs.logSearch?.value || ""
  setState({ selectedSeverity, logQuery })
}

const refreshServiceSelector = () => {
  const state = getState()
  const selectedService = syncServiceSelector(
    refs.logServiceSelector,
    state.environments,
    state.selectedProject,
    state.selectedService,
  )
  setState({ selectedService })
}

const logsController = createLogsController({
  bridge: desktopBridge,
  runtime: desktopBridge.runtime,
  refs,
  readSelection,
  onStatus: setStatus,
  onToast: showToast,
})

const shellController = createShellController({
  bridge: desktopBridge,
  refs,
  readSelection,
  onStatus: setStatus,
  onToast: showToast,
})

const metricsController = createMetricsController({
  bridge: desktopBridge,
  refs,
  onStatus: setStatus,
})

const remotesController = createRemotesController({
  bridge: desktopBridge,
  refs,
  getProject: () => getState().selectedProject,
  onStatus: setStatus,
  onToast: showToast,
})

const refreshDashboard = async () => {
  setStatus("Status: syncing dashboard...")
  const dashboard = await loadDashboard()
  setMetricText(dashboard, refs)
  renderWarnings(refs.warningList, dashboard.warnings)
  renderEnvironmentList(refs.envList, dashboard.environments)

  const previousProject = getState().selectedProject
  syncProjectSelectors(
    { envSelector: refs.envSelector, logSelector: refs.logSelector },
    dashboard.environments,
    previousProject,
  )

  const selectedProject = refs.envSelector?.value || ""
  setState({ environments: dashboard.environments, selectedProject })
  if (!selectedProject && dashboard.environments.length > 0) {
    setState({ selectedProject: projectKey(dashboard.environments[0]) })
  }

  if (refs.logSelector && refs.logSelector.value !== getState().selectedProject) {
    refs.logSelector.value = getState().selectedProject
  }
  if (refs.envSelector && refs.envSelector.value !== getState().selectedProject) {
    refs.envSelector.value = getState().selectedProject
  }

  refreshServiceSelector()
  syncLogFiltersState()
  await metricsController.refresh({ silent: true })
  await remotesController.refresh({ silent: true })
  await shellController.loadShellUser()
  await logsController.refresh()
  setStatus(`Status: refreshed at ${new Date().toLocaleTimeString()}`)
}

const onboardingController = createOnboardingController({
  bridge: desktopBridge,
  refs,
  onStatus: setStatus,
  onToast: showToast,
  onProjectAdded: refreshDashboard,
})

const actionsController = createActionsController({
  bridge: desktopBridge,
  getProject: () => getState().selectedProject,
  refreshDashboard,
  onStatus: setStatus,
  onToast: showToast,
})

const settingsController = createSettingsController({
  bridge: desktopBridge,
  refs,
  onStatus: setStatus,
  onToast: showToast,
})

document.addEventListener("click", async (event) => {
  const target = event.target
  if (!(target instanceof HTMLElement)) {
    return
  }

  const action = target.dataset.action
  if (!action) {
    return
  }

  if (action === "refresh-logs") {
    await logsController.refresh()
    return
  }
  if (action === "refresh-metrics") {
    await metricsController.refresh()
    return
  }
  if (action === "browse-project") {
    await onboardingController.browseProject()
    return
  }
  if (action === "add-project") {
    await onboardingController.addProject()
    return
  }
  if (action === "refresh-remotes") {
    await remotesController.refresh()
    return
  }
  if (action === "save-remote") {
    await remotesController.saveRemote()
    return
  }
  if (action === "remote-test") {
    await remotesController.testRemote(String(target.dataset.remote || ""))
    return
  }
  if (action === "remote-sync") {
    await remotesController.runSyncPreset(String(target.dataset.remote || ""), String(target.dataset.preset || ""))
    return
  }
  if (action === "toggle-live") {
    await logsController.toggleLive()
    return
  }
  if (action === "open-shell") {
    await shellController.openShell()
    return
  }
  if (action === "reset-shell-users") {
    await shellController.resetShellUsers()
    return
  }
  if (action === "reset-settings") {
    await settingsController.reset()
    return
  }

  await actionsController.handle(action, target.dataset.env || "")
})

if (refs.refresh) {
  refs.refresh.addEventListener("click", () => {
    refreshDashboard()
  })
}

if (refs.openSettings) {
  refs.openSettings.addEventListener("click", () => settingsController.toggleDrawer(true))
}

if (refs.closeSettings) {
  refs.closeSettings.addEventListener("click", () => settingsController.toggleDrawer(false))
}

if (refs.settingsDrawer) {
  refs.settingsDrawer.addEventListener("click", (event) => {
    if (event.target === refs.settingsDrawer) {
      settingsController.toggleDrawer(false)
    }
  })
}

document.addEventListener("keydown", (event) => {
  if (event.key === "Escape") {
    settingsController.toggleDrawer(false)
  }
  if ((event.ctrlKey || event.metaKey) && event.key === ",") {
    event.preventDefault()
    settingsController.toggleDrawer(true)
  }
})

const syncProjectSelectorsFrom = async (source) => {
  if (source === "env" && refs.logSelector) {
    refs.logSelector.value = refs.envSelector?.value || ""
  }
  if (source === "log" && refs.envSelector) {
    refs.envSelector.value = refs.logSelector?.value || ""
  }
  syncProjectState()
  refreshServiceSelector()
  await shellController.loadShellUser()
  if (logsController.isLiveEnabled()) {
    await logsController.stopLive()
    await logsController.toggleLive()
  } else {
    await logsController.refresh()
  }
  await remotesController.refresh({ silent: true })
}

if (refs.envSelector) {
  refs.envSelector.addEventListener("change", async () => {
    await syncProjectSelectorsFrom("env")
  })
}

if (refs.logSelector) {
  refs.logSelector.addEventListener("change", async () => {
    await syncProjectSelectorsFrom("log")
  })
}

if (refs.logServiceSelector) {
  refs.logServiceSelector.addEventListener("change", async () => {
    syncServiceState()
    if (logsController.isLiveEnabled()) {
      await logsController.stopLive()
      await logsController.toggleLive()
      return
    }
    await logsController.refresh()
  })
}

if (refs.logSeverity) {
  refs.logSeverity.addEventListener("change", () => {
    syncLogFiltersState()
    logsController.applyFilters()
  })
}

if (refs.logSearch) {
  refs.logSearch.addEventListener("input", () => {
    syncLogFiltersState()
    logsController.applyFilters()
  })
}

if (refs.shellUser) {
  refs.shellUser.addEventListener("change", () => {
    shellController.saveShellUser()
  })
}

if (refs.themeSelect) {
  refs.themeSelect.addEventListener("change", () => {
    settingsController.save()
  })
}

if (refs.proxyTarget) {
  refs.proxyTarget.addEventListener("change", async () => {
    await settingsController.save()
  })
}

if (refs.preferredBrowser) {
  refs.preferredBrowser.addEventListener("change", () => {
    settingsController.save()
  })
}

if (window.matchMedia) {
  window.matchMedia("(prefers-color-scheme: dark)").addEventListener("change", () => {
    if ((refs.themeSelect?.value || "system") === "system") {
      settingsController.load()
    }
  })
}

setStatus("Status: ready.")
setState({ selectedService: "all", selectedSeverity: "all", logQuery: "" })
await settingsController.load()
await refreshDashboard()
metricsController.startAutoRefresh()

window.addEventListener("beforeunload", () => {
  metricsController.stopAutoRefresh()
})
