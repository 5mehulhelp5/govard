import { projectKey, serviceTargets } from "./dashboard.js"

export const resolveLogTarget = ({ project = "", service = "web" } = {}) => ({
  project: String(project || "").trim(),
  service: String(service || "web").trim() || "web",
})

export const syncServiceSelector = (selector, environments, project, selectedService = "web") => {
  if (!selector) {
    return "web"
  }
  const env = environments.find((item) => projectKey(item) === project)
  const targets = env ? serviceTargets(env) : ["web"]
  selector.innerHTML = ""
  targets.forEach((target) => {
    const option = document.createElement("option")
    option.value = target
    option.textContent = target
    selector.appendChild(option)
  })
  const hasSelected = targets.includes(selectedService)
  selector.value = hasSelected ? selectedService : targets[0]
  return selector.value
}

export const createLogsController = ({ bridge, runtime, refs, readSelection, onStatus, onToast }) => {
  let livePoll = null
  let liveEnabled = false

  const appendLogLine = (line) => {
    if (!refs.logOutput) {
      return
    }
    const current = refs.logOutput.textContent || ""
    refs.logOutput.textContent = `${current}\n${line}`.trim()
    refs.logOutput.scrollTop = refs.logOutput.scrollHeight
  }

  const refresh = async () => {
    const { project, service } = readSelection()
    if (!project) {
      if (refs.logOutput) {
        refs.logOutput.textContent = "Select an environment to view logs."
      }
      return
    }
    if (refs.logOutput) {
      refs.logOutput.textContent = "Loading logs..."
    }
    try {
      const logs = await bridge.getLogsForService(project, service)
      refs.logOutput.textContent = logs || "No logs available."
    } catch (err) {
      refs.logOutput.textContent = `Failed to load logs: ${err}`
    }
  }

  const stopLive = async () => {
    liveEnabled = false
    if (refs.toggleLive) {
      refs.toggleLive.textContent = "Live: Off"
    }
    if (livePoll) {
      clearInterval(livePoll)
      livePoll = null
    }
    try {
      await bridge.stopLogStream()
    } catch (_err) {
      // Fallback polling mode may not have a stream to stop.
    }
  }

  const startLive = async () => {
    const { project, service } = readSelection()
    if (!project) {
      onStatus("Select an environment to stream logs.")
      return
    }

    liveEnabled = true
    if (refs.toggleLive) {
      refs.toggleLive.textContent = "Live: On"
    }

    if (bridge.startLogStreamForService && runtime?.EventsOn) {
      try {
        await bridge.startLogStreamForService(project, service)
        return
      } catch (_err) {
        // Fall back to polling.
      }
    }
    await refresh()
    livePoll = setInterval(refresh, 2000)
  }

  const toggleLive = async () => {
    if (liveEnabled) {
      await stopLive()
      return
    }
    await startLive()
  }

  if (runtime?.EventsOn) {
    runtime.EventsOn("logs:line", appendLogLine)
    runtime.EventsOn("logs:status", (message) => {
      onStatus(message)
      onToast(message, "success")
    })
    runtime.EventsOn("logs:error", (message) => {
      onStatus(message)
      onToast(message, "error")
    })
  }

  return {
    refresh,
    toggleLive,
    stopLive,
    isLiveEnabled: () => liveEnabled,
  }
}

