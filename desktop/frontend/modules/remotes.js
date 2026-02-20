import { clearChildren, escapeHTML } from "../utils/dom.js"

const asNumber = (value, fallback = 0) => {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : fallback
}

const normalizeCapabilities = (value) => {
  if (!Array.isArray(value)) {
    return []
  }
  return value
    .map((item) => String(item || "").trim().toLowerCase())
    .filter((item) => item !== "")
}

const normalizeRemote = (remote = {}) => ({
  name: String(remote.name || remote.Name || "").trim(),
  host: String(remote.host || remote.Host || "").trim(),
  user: String(remote.user || remote.User || "").trim(),
  path: String(remote.path || remote.Path || "").trim(),
  port: asNumber(remote.port ?? remote.Port, 22),
  environment: String(remote.environment || remote.Environment || "staging")
    .trim()
    .toLowerCase(),
  protected: Boolean(remote.protected ?? remote.Protected),
  authMethod: String(remote.authMethod || remote.AuthMethod || "keychain")
    .trim()
    .toLowerCase(),
  capabilities: normalizeCapabilities(remote.capabilities || remote.Capabilities),
})

export const normalizeRemotesPayload = (payload = {}) => {
  const remotesRaw = Array.isArray(payload.remotes)
    ? payload.remotes
    : Array.isArray(payload.Remotes)
      ? payload.Remotes
      : []

  const warningsRaw = Array.isArray(payload.warnings)
    ? payload.warnings
    : Array.isArray(payload.Warnings)
      ? payload.Warnings
      : []

  return {
    project: String(payload.project || payload.Project || "").trim(),
    remotes: remotesRaw.map(normalizeRemote),
    warnings: warningsRaw.map((item) => String(item || "").trim()).filter((item) => item !== ""),
  }
}

export const normalizeRemotePreset = (preset = "") => {
  const normalized = String(preset || "")
    .trim()
    .toLowerCase()
  if (["file", "files", "source", "code"].includes(normalized)) {
    return "files"
  }
  if (["media", "assets"].includes(normalized)) {
    return "media"
  }
  if (["db", "database"].includes(normalized)) {
    return "db"
  }
  if (["full", "all"].includes(normalized)) {
    return "full"
  }
  return ""
}

const renderWarnings = (container, warnings = []) => {
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

const renderRemotes = (container, remotes = []) => {
  if (!container) {
    return
  }
  if (!remotes.length) {
    container.innerHTML = `<div class="panel__empty">No remotes configured yet.</div>`
    return
  }

  container.innerHTML = remotes
    .map((remote) => {
      const capabilities = remote.capabilities.length ? remote.capabilities.join(", ") : "none"
      const safety = remote.protected ? "protected" : "writable"
      return `
      <article class="remote-card">
        <div class="remote-card__head">
          <h3>${escapeHTML(remote.name)}</h3>
          <span class="env__status ${remote.protected ? "env__status--idle" : "env__status--live"}">${escapeHTML(
            remote.environment,
          )}</span>
        </div>
        <p>${escapeHTML(`${remote.user}@${remote.host}:${remote.port}`)}</p>
        <p>${escapeHTML(remote.path)}</p>
        <p>Auth ${escapeHTML(remote.authMethod)} | ${escapeHTML(safety)} | Caps ${escapeHTML(capabilities)}</p>
        <div class="remote-card__actions">
          <button class="button button--ghost" data-action="remote-test" data-remote="${escapeHTML(remote.name)}">Test</button>
          <button class="button button--ghost" data-action="remote-sync" data-remote="${escapeHTML(remote.name)}" data-preset="files">Plan Files</button>
          <button class="button button--ghost" data-action="remote-sync" data-remote="${escapeHTML(remote.name)}" data-preset="media">Plan Media</button>
          <button class="button button--ghost" data-action="remote-sync" data-remote="${escapeHTML(remote.name)}" data-preset="db">Plan DB</button>
          <button class="button button--ghost" data-action="remote-sync" data-remote="${escapeHTML(remote.name)}" data-preset="full">Plan Full</button>
        </div>
      </article>
    `
    })
    .join("")
}

const readRemoteInput = (refs) => ({
  name: refs.remoteName?.value || "",
  host: refs.remoteHost?.value || "",
  user: refs.remoteUser?.value || "",
  path: refs.remotePath?.value || "",
  port: asNumber(refs.remotePort?.value, 22),
  environment: refs.remoteEnvironment?.value || "staging",
  capabilities: refs.remoteCapabilities?.value || "files,media,db,deploy",
  authMethod: refs.remoteAuthMethod?.value || "keychain",
  protected: Boolean(refs.remoteProtected?.checked),
})

const validateRemoteInput = (input) => {
  if (!input.name.trim()) {
    return "Remote name is required"
  }
  if (!input.host.trim()) {
    return "Remote host is required"
  }
  if (!input.user.trim()) {
    return "Remote user is required"
  }
  if (!input.path.trim()) {
    return "Remote path is required"
  }
  if (input.port <= 0 || input.port > 65535) {
    return "Remote port must be between 1 and 65535"
  }
  return ""
}

export const createRemotesController = ({ bridge, refs, getProject, onStatus, onToast }) => {
  const refresh = async ({ silent = false } = {}) => {
    const project = String(getProject?.() || "").trim()
    if (!project) {
      renderRemotes(refs.remotesList, [])
      renderWarnings(refs.remotesWarnings, ["Select an environment to load remotes."])
      return null
    }

    try {
      const payload = normalizeRemotesPayload(await bridge.getRemotes(project))
      renderRemotes(refs.remotesList, payload.remotes)
      renderWarnings(refs.remotesWarnings, payload.warnings)
      if (!silent) {
        onStatus(`Status: remotes loaded for ${project}`)
      }
      return payload
    } catch (err) {
      renderRemotes(refs.remotesList, [])
      renderWarnings(refs.remotesWarnings, [`Remotes unavailable: ${err}`])
      if (!silent) {
        onStatus("Status: remotes unavailable")
      }
      return null
    }
  }

  const saveRemote = async () => {
    const project = String(getProject?.() || "").trim()
    if (!project) {
      onStatus("Select an environment before saving remotes.")
      onToast("Select an environment before saving remotes.", "warning")
      return
    }

    const input = readRemoteInput(refs)
    const validationError = validateRemoteInput(input)
    if (validationError) {
      onStatus(validationError)
      onToast(validationError, "warning")
      return
    }

    const message = await bridge.addRemote(
      project,
      input.name,
      input.host,
      input.user,
      input.path,
      input.port,
      input.environment,
      input.capabilities,
      input.authMethod,
      input.protected,
    )

    const response = String(message || "")
    const isError = response.toLowerCase().includes("failed")
    onStatus(response || "Remote saved")
    onToast(response || "Remote saved", isError ? "error" : "success")
    await refresh({ silent: true })
  }

  const testRemote = async (remoteName) => {
    const project = String(getProject?.() || "").trim()
    if (!project || !remoteName) {
      return
    }

    const message = await bridge.testRemote(project, remoteName)
    const response = String(message || "")
    const isError = response.toLowerCase().includes("failed")
    onStatus(isError ? `Status: remote ${remoteName} test failed` : `Status: remote ${remoteName} test finished`)
    onToast(response || `Remote test finished for ${remoteName}`, isError ? "error" : "success")
  }

  const runSyncPreset = async (remoteName, preset) => {
    const normalizedPreset = normalizeRemotePreset(preset)
    const project = String(getProject?.() || "").trim()
    if (!project || !remoteName || !normalizedPreset) {
      return
    }

    const message = await bridge.runRemoteSyncPreset(project, remoteName, normalizedPreset)
    const response = String(message || "")
    const isError = response.toLowerCase().includes("failed")
    onStatus(
      isError
        ? `Status: sync plan ${normalizedPreset} for ${remoteName} failed`
        : `Status: sync plan ${normalizedPreset} for ${remoteName} ready`,
    )
    onToast(response || `Sync plan generated for ${remoteName}`, isError ? "error" : "success")
  }

  return {
    refresh,
    saveRemote,
    testRemote,
    runSyncPreset,
  }
}
