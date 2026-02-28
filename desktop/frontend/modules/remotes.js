import { clearChildren, escapeHTML } from "../utils/dom.js";

const asNumber = (value, fallback = 0) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const normalizeCapabilities = (value) => {
  if (!Array.isArray(value)) {
    return [];
  }
  return value
    .map((item) =>
      String(item || "")
        .trim()
        .toLowerCase(),
    )
    .filter((item) => item !== "");
};

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
  capabilities: normalizeCapabilities(
    remote.capabilities || remote.Capabilities,
  ),
});

export const normalizeRemotesPayload = (payload = {}) => {
  const remotesRaw = Array.isArray(payload.remotes)
    ? payload.remotes
    : Array.isArray(payload.Remotes)
      ? payload.Remotes
      : [];

  const warningsRaw = Array.isArray(payload.warnings)
    ? payload.warnings
    : Array.isArray(payload.Warnings)
      ? payload.Warnings
      : [];

  return {
    project: String(payload.project || payload.Project || "").trim(),
    remotes: remotesRaw.map(normalizeRemote),
    warnings: warningsRaw
      .map((item) => String(item || "").trim())
      .filter((item) => item !== ""),
  };
};

export const normalizeRemotePreset = (preset = "") => {
  const normalized = String(preset || "")
    .trim()
    .toLowerCase();
  if (["file", "files", "source", "code"].includes(normalized)) {
    return "files";
  }
  if (["media", "assets"].includes(normalized)) {
    return "media";
  }
  if (["db", "database"].includes(normalized)) {
    return "db";
  }
  if (["full", "all"].includes(normalized)) {
    return "full";
  }
  return "";
};

const renderWarnings = (container, warnings = []) => {
  if (!container) {
    return;
  }
  clearChildren(container);
  warnings.forEach((warning) => {
    const item = document.createElement("li");
    item.textContent = warning;
    container.appendChild(item);
  });
};

const renderRemotes = (container, remotes = []) => {
  if (!container) {
    return;
  }

  if (!remotes.length) {
    container.innerHTML = `
      <div class="p-8 text-center text-slate-500 border border-dashed border-[#22492f] rounded-xl">
        No remotes configured for this environment.
      </div>`;
    return;
  }

  container.innerHTML = remotes
    .map((remote) => {
      const isProd =
        remote.environment === "prod" || remote.environment === "production";
      const themeColor = isProd ? "amber" : "blue";
      const themeIcon = isProd ? "rocket_launch" : "science";
      const borderColor = isProd ? "border-amber-500/20" : "border-[#22492f]";
      const statusText = isProd ? "Protected" : "Connected";

      return `
      <div class="glass-card rounded-xl p-0 overflow-hidden group mb-6 border ${borderColor}">
        <div class="p-6 border-b border-[#22492f] flex justify-between items-start bg-gradient-to-r from-[#1a3322] to-[#1a3322]/50 relative overflow-hidden">
          ${isProd ? `<div class="absolute top-0 right-0 w-16 h-16 bg-gradient-to-bl from-amber-500/10 to-transparent pointer-events-none"></div>` : ""}
          <div class="flex items-start gap-4 z-10">
            <div class="p-3 rounded-lg bg-${themeColor}-500/10 border border-${themeColor}-500/20 text-${themeColor}-400">
              <span class="material-symbols-outlined">${themeIcon}</span>
            </div>
            <div>
              <h3 class="text-white text-lg font-semibold flex items-center gap-2">
                  ${escapeHTML(remote.name)}
                  <span class="px-2 py-0.5 rounded text-[10px] font-bold bg-${themeColor}-500/20 text-${themeColor}-400 border border-${themeColor}-500/30 uppercase tracking-wide">${statusText}</span>
              </h3>
              <div class="flex items-center gap-4 mt-1 text-xs text-slate-400 font-mono">
                  <span class="flex items-center gap-1">
                      <span class="material-symbols-outlined text-[14px]">dns</span>
                      ${escapeHTML(remote.host)}
                  </span>
                  <span class="flex items-center gap-1">
                      <span class="material-symbols-outlined text-[14px]">schedule</span>
                      Last sync: ${remote.lastSync || "never"}
                  </span>
              </div>
            </div>
          </div>
          <div class="flex items-center gap-1 group/dropdown">
            <button data-action="remote-test" data-remote="${escapeHTML(remote.name)}" class="p-1.5 text-slate-500 hover:text-white hover:bg-white/5 rounded-lg transition-all" title="Test Connection">
                <span class="material-symbols-outlined text-[20px]">wifi_tethering</span>
            </button>
          </div>
        </div>
      </div>
      <div class="p-6">
          <div class="grid grid-cols-2 gap-4 mb-6">
            <div class="bg-[#102316]/50 rounded-lg p-3 border border-[#2e573a]">
              <div class="text-xs text-slate-500 mb-1">Database Size</div>
              <div class="text-white font-mono font-medium">${remote.dbSize || "0 MB"}</div>
            </div>
            <div class="bg-[#102316]/50 rounded-lg p-3 border border-[#2e573a]">
              <div class="text-xs text-slate-500 mb-1">Media Files</div>
              <div class="text-white font-mono font-medium">${remote.mediaSize || "0 MB"}</div>
            </div>
          </div>
          <div class="flex flex-col gap-3">
            <div class="flex flex-wrap gap-3">
                <button data-action="open-sync-modal" data-remote="${escapeHTML(remote.name)}" data-preset="full" class="flex-1 px-4 py-2.5 bg-[#22492f] hover:bg-[#2e573a] border border-[#366b47] rounded-lg text-sm text-white font-medium transition-all flex items-center justify-center gap-2 group/btn">
                    <span class="material-symbols-outlined text-[18px] group-hover/btn:text-purple-400 transition-colors">all_inclusive</span>
                    Pull Everything
                </button>
                <button data-action="open-sync-modal" data-remote="${escapeHTML(remote.name)}" data-preset="db" class="flex-1 px-4 py-2.5 bg-[#22492f] hover:bg-[#2e573a] border border-[#366b47] rounded-lg text-sm text-white font-medium transition-all flex items-center justify-center gap-2 group/btn">
                    <span class="material-symbols-outlined text-[18px] group-hover/btn:text-primary transition-colors">database</span>
                    Pull Database
                </button>
                <button data-action="open-sync-modal" data-remote="${escapeHTML(remote.name)}" data-preset="media" class="flex-1 px-4 py-2.5 bg-[#22492f] hover:bg-[#2e573a] border border-[#366b47] rounded-lg text-sm text-white font-medium transition-all flex items-center justify-center gap-2 group/btn">
                    <span class="material-symbols-outlined text-[18px] group-hover/btn:text-blue-400 transition-colors">perm_media</span>
                    Pull Media
                </button>
            </div>
            <div class="flex flex-wrap gap-3">
                <button data-action="open-remote-shell" data-remote="${escapeHTML(remote.name)}" class="flex-1 px-4 py-2.5 bg-[#102316] hover:bg-[#1a3322] border border-[#2e573a] rounded-lg text-sm text-slate-300 hover:text-white font-medium transition-all flex items-center justify-center gap-2 group/btn">
                    <span class="material-symbols-outlined text-[18px] opacity-70 group-hover/btn:opacity-100">terminal</span>
                    Open SSH
                </button>
                <button data-action="open-remote-db" data-remote="${escapeHTML(remote.name)}" class="flex-1 px-4 py-2.5 bg-[#102316] hover:bg-[#1a3322] border border-[#2e573a] rounded-lg text-sm text-slate-300 hover:text-white font-medium transition-all flex items-center justify-center gap-2 group/btn">
                    <span class="material-symbols-outlined text-[18px] opacity-70 group-hover/btn:opacity-100">database</span>
                    Open Database
                </button>
                <button data-action="open-remote-sftp" data-remote="${escapeHTML(remote.name)}" class="flex-1 px-4 py-2.5 bg-[#102316] hover:bg-[#1a3322] border border-[#2e573a] rounded-lg text-sm text-slate-300 hover:text-white font-medium transition-all flex items-center justify-center gap-2 group/btn">
                    <span class="material-symbols-outlined text-[18px] opacity-70 group-hover/btn:opacity-100">folder_open</span>
                    Open SFTP
                </button>
            </div>
          </div>
          ${remote.protected ? `<div class="mt-4 flex items-center gap-2 p-2 bg-amber-900/10 border border-amber-900/30 rounded text-amber-500/80 text-xs"><span class="material-symbols-outlined text-[16px]">info</span>Syncing from Production creates a local backup automatically.</div>` : ""}
        </div>
      </div>
    `;
    })
    .join("");
};

const safeRemotes = {
  Remotes: [
    {
      Name: "Staging",
      Host: "192.168.1.45",
      Environment: "staging",
      LastSync: "2m ago",
      DbSize: "458 MB",
      MediaSize: "1.2 GB",
    },
    {
      Name: "Production",
      Host: "203.0.113.15",
      Environment: "production",
      LastSync: "1h ago",
      DbSize: "2.4 GB",
      MediaSize: "8.5 GB",
      Protected: true,
    },
  ],
};

export const createRemotesController = ({
  bridge,
  refs,
  getProject,
  getSyncConfig,
  onStatus,
  onToast,
}) => {
  const refresh = async ({ silent = false } = {}) => {
    const project = String(getProject?.() || "").trim();
    if (!project) {
      renderRemotes(refs.remotesList, []);
      renderWarnings(refs.remotesWarnings, [
        "Select an environment to load remotes.",
      ]);
      return null;
    }

    try {
      const payload = normalizeRemotesPayload(await bridge.getRemotes(project));
      renderRemotes(refs.remotesList, payload.remotes);
      renderWarnings(refs.remotesWarnings, payload.warnings);
      if (!silent) {
        onStatus(`Status: remotes loaded for ${project}`);
      }
      return payload;
    } catch (err) {
      const payload = normalizeRemotesPayload(safeRemotes);
      renderRemotes(refs.remotesList, payload.remotes);
      renderWarnings(refs.remotesWarnings, payload.warnings);
      if (!silent) {
        onStatus("Status: showing local remotes fallback");
      }
      return payload;
    }
  };

  const testRemote = async (remoteName) => {
    const project = String(getProject?.() || "").trim();
    if (!project || !remoteName) {
      return;
    }

    const btn = document.querySelector(
      `[data-action="remote-test"][data-remote="${remoteName}"]`,
    );
    const icon = btn?.querySelector(".material-symbols-outlined");

    if (icon) {
      icon.classList.add("animate-spin");
      icon.textContent = "sync";
    }
    if (btn) btn.disabled = true;

    onToast(`Checking connection to ${remoteName}...`, "info");
    onStatus(`Testing connection to ${remoteName}...`);

    try {
      const message = await bridge.testRemote(project, remoteName);
      const response = String(message || "");
      const isError = response.toLowerCase().includes("failed");
      onStatus(
        isError
          ? `Remote ${remoteName} connection failed`
          : `Remote ${remoteName} connection successful`,
      );
      onToast(
        isError
          ? `Connection to ${remoteName} failed.`
          : `Connection to ${remoteName} successful!`,
        isError ? "error" : "success",
      );
    } catch (err) {
      onStatus(`Error testing connection to ${remoteName}`);
      onToast(`Error: ${err}`, "error");
    } finally {
      if (icon) {
        icon.classList.remove("animate-spin");
        icon.textContent = "wifi_tethering";
      }
      if (btn) btn.disabled = false;
    }
  };

  const runSyncPreset = async (remoteName, preset) => {
    const normalizedPreset = normalizeRemotePreset(preset);
    const project = String(getProject?.() || "").trim();
    if (!project || !remoteName || !normalizedPreset) {
      return;
    }

    const syncConfig = getSyncConfig?.() || {};
    const message = await bridge.runRemoteSyncPreset(
      project,
      remoteName,
      normalizedPreset,
      {
        sanitize: Boolean(syncConfig.sanitize),
        excludeLogs: Boolean(syncConfig.excludeLogs),
        compress: Boolean(syncConfig.compress),
      },
    );
    const response = String(message || "");
    const isError = response.toLowerCase().includes("failed");
    onStatus(
      isError
        ? `Failed to generate sync plan for ${remoteName}`
        : `Sync plan for ${remoteName} is ready`,
    );
    onToast(
      isError
        ? `Failed to prepare sync for ${remoteName}.`
        : `Sync plan for ${remoteName} prepared.`,
      isError ? "error" : "success",
    );
  };

  const renderSyncOptions = (container, preset, optionsDef, currentConfig) => {
    if (!container) return;

    const html = (optionsDef || [])
      .map((opt) => {
        const isChecked =
          currentConfig[opt.key] !== undefined
            ? currentConfig[opt.key]
            : opt.defaultValue;

        return `
        <label class="flex items-center justify-between cursor-pointer group">
          <div>
            <div class="text-sm font-medium text-white mb-0.5">${escapeHTML(opt.label)}</div>
            <div class="text-xs text-slate-400">${escapeHTML(opt.description)}</div>
          </div>
          <div class="relative inline-block w-10 h-6 align-middle select-none transition duration-200 ease-in">
            <input
              type="checkbox"
              data-action="toggle-sync-config"
              data-preset="${escapeHTML(preset)}"
              data-config="${escapeHTML(opt.key)}"
              class="toggle-checkbox absolute block w-4 h-4 rounded-full bg-white border-4 border-slate-600 appearance-none cursor-pointer transition-all duration-300 top-1 left-1 checked:left-5 checked:bg-white checked:border-white/0"
              ${isChecked ? "checked" : ""}
            />
            <span class="toggle-label block overflow-hidden h-6 rounded-full bg-slate-700 cursor-pointer transition-colors duration-300 group-hover:bg-slate-600 ${isChecked ? "bg-primary" : ""}"></span>
          </div>
        </label>
      `;
      })
      .join("");

    container.innerHTML = html;
  };

  const toggleSyncConfig = (preset, key, currentState, onUpdate) => {
    const nextValue = !currentState[key];
    const nextState = { ...currentState, [key]: nextValue };
    onUpdate(nextState);
    onStatus(`Option "${key}" ${nextValue ? "enabled" : "disabled"}.`);
    return nextState;
  };

  return {
    refresh,
    testRemote,
    runSyncPreset,
    renderSyncOptions,
    toggleSyncConfig,
  };
};
