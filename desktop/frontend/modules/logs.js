import { projectKey, serviceTargets } from "./dashboard.js"

const errorPattern = /\b(error|critical|fail|failed|exception|fatal|panic)\b/i
const warnPattern = /\b(warn|warning|deprecated)\b/i

export const normalizeLogSeverity = (severity = "all") => {
  const normalized = String(severity || "all")
    .trim()
    .toLowerCase()
  if (["all", "error", "warn", "info"].includes(normalized)) {
    return normalized
  }
  return "all"
}

export const classifyLogSeverity = (line = "") => {
  const text = String(line || "")
  if (errorPattern.test(text)) {
    return "error"
  }
  if (warnPattern.test(text)) {
    return "warn"
  }
  return "info"
}

export const filterLogsText = (raw = "", severity = "all", query = "") => {
  const selectedSeverity = normalizeLogSeverity(severity)
  const normalizedQuery = String(query || "")
    .trim()
    .toLowerCase()

  const lines = String(raw || "").split("\n")
  const filtered = lines.filter((line) => {
    if (selectedSeverity !== "all" && classifyLogSeverity(line) !== selectedSeverity) {
      return false
    }
    if (normalizedQuery !== "" && !line.toLowerCase().includes(normalizedQuery)) {
      return false
    }
    return true
  })
  return filtered.join("\n").trim()
}

export const resolveLogTarget = ({ project = "", service = "all", severity = "all", query = "" } = {}) => ({
  project: String(project || "").trim(),
  service: String(service || "all").trim() || "all",
  severity: normalizeLogSeverity(severity),
  query: String(query || "").trim(),
})

export const syncServiceSelector = (selector, environments, project, selectedService = "all") => {
  if (!selector) {
    return "all"
  }
  const env = environments.find((item) => projectKey(item) === project)
  const targets = env ? serviceTargets(env) : ["web"]
  const mergedTargets = ["all", ...targets.filter((target) => target !== "all")]
  selector.innerHTML = ""
  mergedTargets.forEach((target) => {
    const option = document.createElement("option")
    option.value = target
    option.textContent = target
    selector.appendChild(option)
  })
  const hasSelected = mergedTargets.includes(selectedService)
  selector.value = hasSelected ? selectedService : mergedTargets[0]
  return selector.value
}

export const createLogsController = ({ bridge, runtime, refs, readSelection, onStatus, onToast }) => {
  let livePoll = null
  let liveEnabled = false
  let rawLogOutput = ""

  const renderFilteredOutput = () => {
    if (!refs.logOutput) {
      return
    }
    const { severity, query } = readSelection()
    const filtered = filterLogsText(rawLogOutput, severity, query)
    refs.logOutput.textContent = filtered || "No logs match the current filters."
  }

  const appendLogLine = (line) => {
    rawLogOutput = rawLogOutput ? `${rawLogOutput}\n${line}` : String(line || "")
    renderFilteredOutput()
    if (refs.logOutput) {
      refs.logOutput.scrollTop = refs.logOutput.scrollHeight
    }
  }

  const refresh = async () => {
    const { project, service } = readSelection()
    if (!project) {
      if (refs.logOutput) {
        refs.logOutput.textContent = "Select an environment to view logs."
      }
      rawLogOutput = ""
      return
    }
    if (refs.logOutput) {
      refs.logOutput.textContent = "Loading logs..."
    }
    try {
      const logs = await bridge.getLogsForService(project, service)
      rawLogOutput = String(logs || "")
      renderFilteredOutput()
    } catch (err) {
      rawLogOutput = ""
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
    applyFilters: renderFilteredOutput,
    toggleLive,
    stopLive,
    isLiveEnabled: () => liveEnabled,
  }
}
