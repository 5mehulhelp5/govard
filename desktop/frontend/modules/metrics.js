import { clearChildren, escapeHTML, setText } from "../utils/dom.js"

const asNumber = (value) => {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : 0
}

export const formatMetricPercent = (value = 0) => `${asNumber(value).toFixed(1)}%`
export const formatMetricMB = (value = 0) => `${asNumber(value).toFixed(1)} MB`

const normalizeProjectMetric = (project = {}) => ({
  project: String(project.project || project.Project || "").trim(),
  status: String(project.status || project.Status || "stopped")
    .trim()
    .toLowerCase(),
  cpuPercent: asNumber(project.cpuPercent ?? project.CPUPercent),
  memoryMB: asNumber(project.memoryMB ?? project.MemoryMB),
  memoryPercent: asNumber(project.memoryPercent ?? project.MemoryPercent),
  netRxMB: asNumber(project.netRxMB ?? project.NetRxMB),
  netTxMB: asNumber(project.netTxMB ?? project.NetTxMB),
  oomKilled: Boolean(project.oomKilled ?? project.OOMKilled),
})

export const normalizeMetricsPayload = (payload = {}) => {
  const summary = payload.summary || payload.Summary || {}
  const projectsRaw = Array.isArray(payload.projects)
    ? payload.projects
    : Array.isArray(payload.Projects)
      ? payload.Projects
      : []
  const warnings = Array.isArray(payload.warnings)
    ? payload.warnings
    : Array.isArray(payload.Warnings)
      ? payload.Warnings
      : []

  return {
    updatedAt: String(payload.updatedAt || payload.UpdatedAt || "").trim(),
    summary: {
      activeProjects: asNumber(summary.activeProjects ?? summary.ActiveProjects),
      cpuPercent: asNumber(summary.cpuPercent ?? summary.CPUPercent),
      memoryMB: asNumber(summary.memoryMB ?? summary.MemoryMB),
      netRxMB: asNumber(summary.netRxMB ?? summary.NetRxMB),
      netTxMB: asNumber(summary.netTxMB ?? summary.NetTxMB),
      oomProjects: asNumber(summary.oomProjects ?? summary.OOMProjects),
    },
    projects: projectsRaw.map(normalizeProjectMetric),
    warnings: warnings.map((item) => String(item)),
  }
}

const renderMetricSummary = (refs, summary = {}) => {
  setText(refs.metricActiveProjects, String(summary.activeProjects ?? 0))
  setText(refs.metricCPU, formatMetricPercent(summary.cpuPercent))
  setText(refs.metricMemory, formatMetricMB(summary.memoryMB))
  setText(refs.metricNetRx, formatMetricMB(summary.netRxMB))
  setText(refs.metricNetTx, formatMetricMB(summary.netTxMB))
  setText(refs.metricOOM, String(summary.oomProjects ?? 0))
}

const renderMetricWarnings = (container, warnings = []) => {
  if (!container) {
    return
  }
  clearChildren(container)
  warnings.forEach((warning) => {
    const item = document.createElement("li")
    item.textContent = warning
    container.appendChild(item)
  })
}

const renderMetricProjects = (container, projects = []) => {
  if (!container) {
    return
  }
  if (!projects.length) {
    container.innerHTML = `<div class="panel__empty">No running resource metrics yet.</div>`
    return
  }

  container.innerHTML = projects
    .map((project) => {
      const statusClass = project.status === "running" ? "env__status env__status--live" : "env__status env__status--idle"
      const statusText = project.status === "running" ? "Running" : "Stopped"
      return `
      <article class="metric-card">
        <div class="metric-card__head">
          <h3>${escapeHTML(project.project || "unknown")}</h3>
          <span class="${statusClass}">${statusText}</span>
        </div>
        <p>CPU ${escapeHTML(formatMetricPercent(project.cpuPercent))}</p>
        <p>Memory ${escapeHTML(formatMetricMB(project.memoryMB))} (${escapeHTML(formatMetricPercent(project.memoryPercent))})</p>
        <p>NET RX ${escapeHTML(formatMetricMB(project.netRxMB))} | NET TX ${escapeHTML(formatMetricMB(project.netTxMB))}</p>
        <p>${project.oomKilled ? "OOM: detected" : "OOM: clear"}</p>
      </article>
    `
    })
    .join("")
}

export const createMetricsController = ({ bridge, refs, onStatus }) => {
  let refreshTimer = null

  const renderPayload = (payload) => {
    const metrics = normalizeMetricsPayload(payload)
    renderMetricSummary(refs, metrics.summary)
    renderMetricWarnings(refs.metricsWarnings, metrics.warnings)
    renderMetricProjects(refs.metricsList, metrics.projects)
    return metrics
  }

  const refresh = async ({ silent = false } = {}) => {
    try {
      const payload = await bridge.getResourceMetrics()
      const metrics = renderPayload(payload)
      if (!silent) {
        onStatus(`Status: metrics refreshed at ${new Date().toLocaleTimeString()}`)
      }
      return metrics
    } catch (err) {
      renderPayload({
        summary: {},
        projects: [],
        warnings: [`Metrics unavailable: ${err}`],
      })
      if (!silent) {
        onStatus("Status: metrics unavailable")
      }
      return null
    }
  }

  const startAutoRefresh = () => {
    if (refreshTimer) {
      clearInterval(refreshTimer)
    }
    refreshTimer = setInterval(() => {
      refresh({ silent: true })
    }, 15000)
  }

  const stopAutoRefresh = () => {
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }
  }

  return {
    refresh,
    startAutoRefresh,
    stopAutoRefresh,
  }
}
