import { clearChildren, escapeHTML, setText } from "../utils/dom.js"

export const normalizeDashboardPayload = (data = {}) => ({
  active: data.ActiveEnvironments ?? data.active ?? 0,
  services: data.RunningServices ?? data.services ?? 0,
  queued: data.QueuedTasks ?? data.queued ?? 0,
  activeSummary: data.ActiveSummary ?? data.activeSummary ?? "",
  servicesSummary: data.ServicesSummary ?? data.servicesSummary ?? "",
  queueSummary: data.QueueSummary ?? data.queueSummary ?? "",
  environments: Array.isArray(data.Environments) ? data.Environments : Array.isArray(data.environments) ? data.environments : [],
  warnings: Array.isArray(data.Warnings) ? data.Warnings : Array.isArray(data.warnings) ? data.warnings : [],
})

export const projectKey = (env = {}) => env.Project || env.project || env.Name || env.name || ""

export const domainLabel = (env = {}) => env.Domain || env.domain || env.Name || env.name || projectKey(env)

export const serviceTargets = (env = {}) => {
  const values = Array.isArray(env.ServiceTargets) ? env.ServiceTargets : Array.isArray(env.serviceTargets) ? env.serviceTargets : []
  return values.length ? values : ["web"]
}

export const setMetricText = ({ active, services, queued, activeSummary, servicesSummary, queueSummary }, refs) => {
  setText(refs.statActive, String(active))
  setText(refs.statServices, String(services))
  setText(refs.statQueue, String(queued))
  setText(refs.statActiveHint, activeSummary || "No environments detected")
  setText(refs.statServicesHint, servicesSummary || "Waiting for service data")
  setText(refs.statQueueHint, queueSummary || "Queue idle")
}

export const renderWarnings = (warningList, warnings = []) => {
  if (!warningList) {
    return
  }
  clearChildren(warningList)
  warnings.forEach((warning) => {
    const item = document.createElement("li")
    item.textContent = String(warning)
    warningList.appendChild(item)
  })
}

export const renderEnvironmentList = (container, environments = []) => {
  if (!container) {
    return
  }
  if (!environments.length) {
    container.innerHTML = `<div class="panel__empty">No Govard environments detected.</div>`
    return
  }
  container.innerHTML = environments
    .map((env) => {
      const key = projectKey(env)
      const domain = domainLabel(env)
      const framework = env.Framework || env.framework || "unknown"
      const php = env.PHP || env.php || "n/a"
      const database = env.Database || env.database || "n/a"
      const services = Array.isArray(env.Services) && env.Services.length ? env.Services.join(", ") : "Base stack"
      const running = String(env.Status || env.status || "stopped").toLowerCase() === "running"
      const statusText = running ? "Running" : "Stopped"
      const statusClass = running ? "env__status--live" : "env__status--idle"
      return `
      <article class="env-card">
        <div class="env-card__info">
          <h3>${escapeHTML(domain)}</h3>
          <p>${escapeHTML(framework)} | PHP ${escapeHTML(php)} | ${escapeHTML(database)}</p>
          <p>${escapeHTML(services)}</p>
        </div>
        <div class="env-card__actions">
          <span class="env__status ${statusClass}">${statusText}</span>
          <button class="button button--ghost" data-action="toggle-env" data-env="${escapeHTML(key)}">
            ${running ? "Stop" : "Start"}
          </button>
          <button class="button button--ghost" data-action="open-env" data-env="${escapeHTML(key)}">Open</button>
        </div>
      </article>
    `
    })
    .join("")
}

const syncSingleSelector = (selector, environments, selectedProject) => {
  if (!selector) {
    return
  }
  const previous = selectedProject || selector.value
  selector.innerHTML = ""
  environments.forEach((env) => {
    const option = document.createElement("option")
    option.value = projectKey(env)
    option.textContent = domainLabel(env)
    selector.appendChild(option)
  })
  const exists = environments.some((env) => projectKey(env) === previous)
  selector.value = exists ? previous : environments.length ? projectKey(environments[0]) : ""
}

export const syncProjectSelectors = (selectors, environments = [], selectedProject = "") => {
  syncSingleSelector(selectors.envSelector, environments, selectedProject)
  syncSingleSelector(selectors.logSelector, environments, selectedProject)
}

