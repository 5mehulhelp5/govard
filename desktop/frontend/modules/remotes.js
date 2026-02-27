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
      <div class="flex items-center justify-between pb-2">
        <h3 class="text-white text-lg font-semibold flex items-center gap-2">Connected Remotes</h3>
        <button data-action="open-onboarding" class="px-4 py-2 bg-[#22492f] text-white rounded-lg text-sm font-medium hover:bg-[#2e573a] transition-colors border border-[#366b47] flex items-center gap-2">
          <span class="material-symbols-outlined text-lg">add_link</span>
          Connect New Remote
        </button>
      </div>
      <div class="p-8 text-center text-slate-500 border border-dashed border-[#22492f] rounded-xl">
        No remotes configured for this environment.
      </div>`;
    return;
  }

  const header = `
    <div class="flex items-center justify-between pb-2">
      <h3 class="text-white text-lg font-semibold flex items-center gap-2">Connected Remotes</h3>
      <button data-action="open-onboarding" class="px-4 py-2 bg-[#22492f] text-white rounded-lg text-sm font-medium hover:bg-[#2e573a] transition-colors border border-[#366b47] flex items-center gap-2 shadow-sm shadow-[#102316]">
        <span class="material-symbols-outlined text-lg">add_link</span>
        Connect New Remote
      </button>
    </div>
  `;

  container.innerHTML =
    header +
    remotes
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
          <div class="relative">
            <button class="text-slate-400 hover:text-white transition-colors">
              <span class="material-symbols-outlined">more_vert</span>
            </button>
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
          <div class="flex flex-wrap gap-3">
              <button data-action="open-sync-modal" data-remote="${escapeHTML(remote.name)}" data-preset="db" class="flex-1 px-4 py-2.5 bg-[#22492f] hover:bg-[#2e573a] border border-[#366b47] rounded-lg text-sm text-white font-medium transition-all flex items-center justify-center gap-2 group/btn">
                  <span class="material-symbols-outlined text-[18px] group-hover/btn:text-primary transition-colors">database</span>
                  Pull Database
              </button>
              <button data-action="open-sync-modal" data-remote="${escapeHTML(remote.name)}" data-preset="media" class="flex-1 px-4 py-2.5 bg-[#22492f] hover:bg-[#2e573a] border border-[#366b47] rounded-lg text-sm text-white font-medium transition-all flex items-center justify-center gap-2 group/btn">
                  <span class="material-symbols-outlined text-[18px] group-hover/btn:text-blue-400 transition-colors">perm_media</span>
                  Pull Media
              </button>
              <button data-action="open-sync-modal" data-remote="${escapeHTML(remote.name)}" data-preset="full" class="flex-1 px-4 py-2.5 bg-[#22492f] hover:bg-[#2e573a] border border-[#366b47] rounded-lg text-sm text-white font-medium transition-all flex items-center justify-center gap-2 group/btn">
                  <span class="material-symbols-outlined text-[18px] group-hover/btn:text-purple-400 transition-colors">all_inclusive</span>
                  Pull Everything
              </button>
              <button data-action="remote-test" data-remote="${escapeHTML(remote.name)}" class="px-3 py-2.5 bg-[#102316] hover:bg-[#1a3322] border border-[#2e573a] rounded-lg text-slate-400 hover:text-white transition-all" title="Test Connection">
                  <span class="material-symbols-outlined text-[18px]">wifi_tethering</span>
              </button>
          </div>
          ${remote.protected ? `<div class="mt-4 flex items-center gap-2 p-2 bg-amber-900/10 border border-amber-900/30 rounded text-amber-500/80 text-xs"><span class="material-symbols-outlined text-[16px]">info</span>Syncing from Production creates a local backup automatically.</div>` : ""}
        </div>
      </div>
    `;
      })
      .join("");
};

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
});

const validateRemoteInput = (input) => {
  if (!input.name.trim()) {
    return "Remote name is required";
  }
  if (!input.host.trim()) {
    return "Remote host is required";
  }
  if (!input.user.trim()) {
    return "Remote user is required";
  }
  if (!input.path.trim()) {
    return "Remote path is required";
  }
  if (input.port <= 0 || input.port > 65535) {
    return "Remote port must be between 1 and 65535";
  }
  return "";
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

  const saveRemote = async () => {
    const project = String(getProject?.() || "").trim();
    if (!project) {
      onStatus("Select an environment before saving remotes.");
      onToast("Select an environment before saving remotes.", "warning");
      return;
    }

    const input = readRemoteInput(refs);
    const validationError = validateRemoteInput(input);
    if (validationError) {
      onStatus(validationError);
      onToast(validationError, "warning");
      return;
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
    );

    const response = String(message || "");
    const isError = response.toLowerCase().includes("failed");
    onStatus(response || "Remote saved");
    onToast(response || "Remote saved", isError ? "error" : "success");
    await refresh({ silent: true });
  };

  const testRemote = async (remoteName) => {
    const project = String(getProject?.() || "").trim();
    if (!project || !remoteName) {
      return;
    }

    const message = await bridge.testRemote(project, remoteName);
    const response = String(message || "");
    const isError = response.toLowerCase().includes("failed");
    onStatus(
      isError
        ? `Status: remote ${remoteName} test failed`
        : `Status: remote ${remoteName} test finished`,
    );
    onToast(
      response || `Remote test finished for ${remoteName}`,
      isError ? "error" : "success",
    );
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
        ? `Status: sync plan ${normalizedPreset} for ${remoteName} failed`
        : `Status: sync plan ${normalizedPreset} for ${remoteName} ready`,
    );
    onToast(
      response || `Sync plan generated for ${remoteName}`,
      isError ? "error" : "success",
    );
  };

  const syncSyncConfigUI = (syncConfig) => {
    const {
      syncToggleSanitize: s,
      syncToggleExcludeLogs: e,
      syncToggleCompress: c,
    } = refs;
    const items = [
      { btn: s, key: "sanitize" },
      { btn: e, key: "excludeLogs" },
      { btn: c, key: "compress" },
    ];

    items.forEach(({ btn, key }) => {
      if (!btn) return;
      const enabled = syncConfig[key];
      const span = btn.querySelector("span");

      if (enabled) {
        btn.classList.remove("bg-[#102316]", "border-slate-700");
        btn.classList.add("bg-primary/20", "border-primary/30");
        if (span) {
          span.classList.remove("left-0.5", "bg-slate-500");
          span.classList.add("right-0.5", "bg-primary");
        }
      } else {
        btn.classList.add("bg-[#102316]", "border-slate-700");
        btn.classList.remove("bg-primary/20", "border-primary/30");
        if (span) {
          span.classList.add("left-0.5", "bg-slate-500");
          span.classList.remove("right-0.5", "bg-primary");
        }
      }
    });
  };

  const toggleSyncConfig = async (key, currentState, onUpdate) => {
    const nextValue = !currentState[key];
    const nextState = { ...currentState, [key]: nextValue };
    onUpdate(nextState);
    syncSyncConfigUI(nextState);
    onStatus(`Status: sync configuration '${key}' set to ${nextValue}`);
  };

  return {
    refresh,
    saveRemote,
    testRemote,
    runSyncPreset,
    syncSyncConfigUI,
    toggleSyncConfig,
  };
};
