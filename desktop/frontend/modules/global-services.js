import { escapeHTML, setText } from "../utils/dom.js";

const globalServiceIcons = {
  caddy: "shield",
  mail: "mail",
  pma: "database",
  portainer: "deployed_code",
  dnsmasq: "dns",
};

const ACTIVE_STATUSES = new Set([
  "running",
  "restarting",
  "starting",
  "healthy",
  "up",
]);

const BULK_START_ENABLED_CLASS =
  "h-10 min-w-[118px] px-3 bg-primary text-background-dark rounded-xl text-xs font-bold uppercase tracking-[0.08em] hover:bg-primary/90 transition-all active:scale-95 inline-flex items-center justify-center gap-1.5 shadow-[0_8px_22px_rgba(13,242,89,0.18)] ring-1 ring-primary/30 whitespace-nowrap disabled:opacity-70 disabled:cursor-not-allowed disabled:active:scale-100";
const BULK_START_DISABLED_CLASS =
  "h-10 min-w-[118px] px-3 bg-[#13261a] text-slate-400 border border-[#2e573a] rounded-xl text-xs font-bold uppercase tracking-[0.08em] transition-all inline-flex items-center justify-center gap-1.5 whitespace-nowrap disabled:opacity-70 disabled:cursor-not-allowed";
const BULK_RESTART_ENABLED_CLASS =
  "h-10 min-w-[118px] px-3 bg-primary text-background-dark rounded-xl text-xs font-bold uppercase tracking-[0.08em] hover:bg-primary/90 transition-all active:scale-95 inline-flex items-center justify-center gap-1.5 shadow-[0_8px_22px_rgba(13,242,89,0.18)] ring-1 ring-primary/30 whitespace-nowrap disabled:opacity-70 disabled:cursor-not-allowed disabled:active:scale-100";
const BULK_RESTART_DISABLED_CLASS =
  "h-10 min-w-[118px] px-3 bg-[#13261a] text-slate-400 border border-[#2e573a] rounded-xl text-xs font-bold uppercase tracking-[0.08em] transition-all inline-flex items-center justify-center gap-1.5 whitespace-nowrap disabled:opacity-70 disabled:cursor-not-allowed";
const BULK_STOP_ENABLED_CLASS =
  "h-10 min-w-[118px] px-3 bg-red-600 text-white border border-red-500 rounded-xl text-xs font-bold uppercase tracking-[0.08em] hover:bg-red-500 transition-all active:scale-95 inline-flex items-center justify-center gap-1.5 shadow-[0_8px_24px_rgba(239,68,68,0.25)] ring-1 ring-red-400/30 whitespace-nowrap disabled:opacity-70 disabled:cursor-not-allowed disabled:active:scale-100";
const BULK_STOP_DISABLED_CLASS =
  "h-10 min-w-[118px] px-3 bg-[#2b1414] text-red-300/60 border border-red-900/40 rounded-xl text-xs font-bold uppercase tracking-[0.08em] transition-all inline-flex items-center justify-center gap-1.5 whitespace-nowrap disabled:opacity-70 disabled:cursor-not-allowed";
const BULK_PULL_CLASS =
  "h-10 min-w-[118px] px-3 bg-[#22492f] text-white border border-[#366b47] rounded-xl text-xs font-bold uppercase tracking-[0.08em] hover:bg-[#2e573a] transition-all active:scale-95 inline-flex items-center justify-center gap-1.5 shadow-[0_8px_20px_rgba(34,73,47,0.35)] ring-1 ring-[#3e7d53]/40 whitespace-nowrap disabled:opacity-70 disabled:cursor-not-allowed disabled:active:scale-100";

const isServiceActive = (service = {}) =>
  ACTIVE_STATUSES.has(String(service.status || "").trim().toLowerCase()) ||
  Boolean(service.running);

const isStopLikeState = (service = {}) => {
  const status = String(service.status || "").trim().toLowerCase();
  const state = String(service.state || "").trim().toLowerCase();
  return (
    status === "stopped" ||
    status === "exited" ||
    status === "dead" ||
    state.includes("stopped") ||
    state.includes("exited") ||
    state.includes("dead")
  );
};

const hasRoutingImpact = (service = {}) =>
  (service.id === "caddy" || service.id === "dnsmasq") &&
  !isServiceActive(service) &&
  isStopLikeState(service);

const normalizeGlobalService = (service = {}) => ({
  id: String(service.id || service.ID || "").trim().toLowerCase(),
  name: String(service.name || service.Name || "").trim() || "Unknown",
  composeService: String(
    service.composeService || service.ComposeService || "",
  ).trim(),
  containerName: String(
    service.containerName || service.ContainerName || "",
  ).trim(),
  status: String(service.status || service.Status || "missing")
    .trim()
    .toLowerCase(),
  state: String(service.state || service.State || "unknown").trim(),
  health: String(service.health || service.Health || "unknown")
    .trim()
    .toLowerCase(),
  statusText: String(service.statusText || service.StatusText || "").trim(),
  running: Boolean(service.running ?? service.Running),
  openable: Boolean(service.openable ?? service.Openable),
  url: String(service.url || service.URL || "").trim(),
});

const placeDnsmasqAfterCaddy = (services = []) => {
  const caddyIndex = services.findIndex((item) => item.id === "caddy");
  const dnsmasqIndex = services.findIndex((item) => item.id === "dnsmasq");
  if (caddyIndex < 0 || dnsmasqIndex < 0 || dnsmasqIndex === caddyIndex + 1) {
    return services;
  }

  const reordered = [...services];
  const [dnsmasq] = reordered.splice(dnsmasqIndex, 1);
  const nextCaddyIndex = reordered.findIndex((item) => item.id === "caddy");
  reordered.splice(nextCaddyIndex + 1, 0, dnsmasq);
  return reordered;
};

export const normalizeGlobalServicesSnapshot = (payload = {}) => {
  const servicesRaw = Array.isArray(payload.services)
    ? payload.services
    : Array.isArray(payload.Services)
      ? payload.Services
      : [];
  return {
    active: Number(payload.active ?? payload.Active ?? 0) || 0,
    total:
      Number(payload.total ?? payload.Total ?? servicesRaw.length) ||
      servicesRaw.length,
    summary:
      String(payload.summary || payload.Summary || "").trim() ||
      "Global services status unavailable",
    warnings: Array.isArray(payload.warnings)
      ? payload.warnings
      : Array.isArray(payload.Warnings)
        ? payload.Warnings
        : [],
    services: placeDnsmasqAfterCaddy(
      servicesRaw.map(normalizeGlobalService).filter((item) => item.id),
    ),
  };
};

const statusChipClass = (status = "missing") => {
  if (status === "running") {
    return "bg-primary/20 border-primary/30 text-primary";
  }
  if (status === "restarting") {
    return "bg-amber-500/20 border-amber-500/30 text-amber-400";
  }
  if (status === "paused") {
    return "bg-amber-500/20 border-amber-500/30 text-amber-500";
  }
  if (status === "missing") {
    return "bg-slate-500/20 border-slate-500/30 text-slate-400";
  }
  return "bg-red-500/20 border-red-500/30 text-red-400";
};

const formatStatusLabel = (status = "missing") => {
  const normalized = String(status || "missing").trim().toLowerCase();
  return normalized.charAt(0).toUpperCase() + normalized.slice(1);
};

const renderServiceCard = (service, selectedService) => {
  const selected = service.id === selectedService;
  const icon = globalServiceIcons[service.id] || "widgets";
  const statusClass = statusChipClass(service.status);
  const isActive = isServiceActive(service);
  const primaryAction = isActive ? "restart" : "start";
  const primaryLabel = isActive ? "Restart" : "Start";
  const primaryIcon = isActive ? "restart_alt" : "play_arrow";
  const showRoutingWarning = hasRoutingImpact(service);
  const routingWarning = showRoutingWarning
    ? `<div class="mt-2 rounded-md border border-amber-500/40 bg-amber-500/10 px-2 py-1.5 text-[10px] text-amber-300 font-medium flex items-start gap-1.5">
          <span class="material-symbols-outlined text-[13px] leading-none mt-[1px]">warning</span>
          <span>Routing warning: ${escapeHTML(service.name)} is stopped. Proxy/domain routing may fail.</span>
        </div>`
    : "";
  const rowClass = selected
    ? "bg-transparent border border-primary/40 shadow-[0_0_0_1px_rgba(13,242,89,0.25)]"
    : "bg-transparent border border-[#2e573a] hover:border-primary/30 hover:bg-[#22492f]/20";
  const serviceName = escapeHTML(service.name);
  const containerName = escapeHTML(service.containerName);
  const statusLabel = escapeHTML(formatStatusLabel(service.status));

  return `
    <article
      data-action="global-service-select-log"
      data-service="${service.id}"
      class="rounded-xl border ${rowClass} p-3 transition-all cursor-pointer"
      title="Select ${serviceName} logs"
    >
      <div class="flex items-start justify-between gap-3">
        <div class="min-w-0">
          <div class="flex items-center gap-2">
            <span class="material-symbols-outlined text-primary text-[18px]">${icon}</span>
            <h4 class="text-sm font-semibold text-white truncate">${serviceName}</h4>
          </div>
          <p class="text-[11px] text-slate-400 mt-1 truncate">${containerName}</p>
        </div>
        <span class="px-2 py-1 rounded-md border text-[10px] font-bold uppercase tracking-wide shrink-0 ${statusClass}">
          ${statusLabel}
        </span>
      </div>
      ${routingWarning}
      <div class="mt-3 grid grid-cols-3 gap-1.5">
        <button
          data-action="global-service-primary"
          data-service="${service.id}"
          data-operation="${primaryAction}"
          data-loading-label="${primaryAction === "restart" ? "Restarting..." : "Starting..."}"
          class="h-9 rounded-lg text-[10px] font-bold bg-primary text-background-dark hover:bg-primary/90 transition-all active:scale-95 flex items-center justify-center gap-1 disabled:opacity-70 disabled:cursor-not-allowed disabled:active:scale-100"
        >
          <span class="material-symbols-outlined text-[14px]">${primaryIcon}</span>
          ${primaryLabel}
        </button>
        <button
          data-action="global-service-stop"
          data-service="${service.id}"
          data-loading-label="Stopping..."
          class="${isActive ? "h-9 rounded-lg text-[10px] font-bold bg-red-600 text-white border border-red-500 hover:bg-red-500 transition-all active:scale-95 flex items-center justify-center gap-1 disabled:opacity-70 disabled:cursor-not-allowed disabled:active:scale-100" : "h-9 rounded-lg text-[10px] font-bold bg-[#13261a] text-slate-500 border border-[#2e573a] cursor-not-allowed opacity-70 flex items-center justify-center gap-1"}"
          ${isActive ? "" : "disabled"}
        >
          <span class="material-symbols-outlined text-[14px] fill-1" style="font-variation-settings: &quot;FILL&quot; 1">stop</span>
          Stop
        </button>
        <button
          data-action="global-service-open"
          data-service="${service.id}"
          data-loading-label="Opening..."
          class="${service.openable ? "h-9 rounded-lg text-[10px] font-bold bg-[#22492f] text-white border border-[#366b47] hover:bg-[#2e573a] transition-all active:scale-95 flex items-center justify-center gap-1 disabled:opacity-70 disabled:cursor-not-allowed disabled:active:scale-100" : "h-9 rounded-lg text-[10px] font-bold border border-[#2a4231] text-slate-500 cursor-not-allowed bg-[#13261a] opacity-70 flex items-center justify-center gap-1"}"
          ${service.openable ? "" : "disabled"}
        >
          <span class="material-symbols-outlined text-[14px]">open_in_new</span>
          Open
        </button>
      </div>
    </article>
  `;
};

export const renderGlobalServices = (
  container,
  services = [],
  selectedService = "",
) => {
  if (!container) {
    return;
  }
  if (!Array.isArray(services) || services.length === 0) {
    container.innerHTML = `
      <div class="rounded-xl border border-dashed border-[#2e573a] bg-[#102316] p-4 text-sm text-slate-400">
        Global services data unavailable.
      </div>
    `;
    return;
  }

  container.innerHTML = services
    .map((service) => renderServiceCard(service, selectedService))
    .join("");
};

const statusStripClass = (service = {}) => {
  const status = String(service.status || "").trim().toLowerCase();
  if (isServiceActive(service)) {
    return {
      chip: "border-primary/30 bg-primary/10 text-primary",
      dot: "bg-primary shadow-[0_0_8px_rgba(13,242,89,0.85)]",
    };
  }
  if (status === "restarting" || status === "starting" || status === "paused") {
    return {
      chip: "border-amber-500/35 bg-amber-500/15 text-amber-300",
      dot: "bg-amber-400",
    };
  }
  if (status === "missing") {
    return {
      chip: "border-slate-500/30 bg-slate-500/10 text-slate-300",
      dot: "bg-slate-400",
    };
  }
  return {
    chip: "border-red-500/35 bg-red-500/10 text-red-300",
    dot: "bg-red-400",
  };
};

const renderStatusStrip = (container, services = []) => {
  if (!container) {
    return;
  }
  if (!Array.isArray(services) || services.length === 0) {
    container.innerHTML =
      '<span class="inline-flex items-center gap-1 rounded-md border border-[#2e573a] bg-[#13261a] px-2 py-1 text-[10px] text-slate-400">Loading services...</span>';
    return;
  }
  container.innerHTML = services
    .map((service, index) => {
      const tone = statusStripClass(service);
      const name = escapeHTML(service.name);
      const label = escapeHTML(formatStatusLabel(service.status));
      const icon = globalServiceIcons[service.id] || "widgets";
      return `<span class="global-status-chip inline-flex items-center gap-1.5 rounded-md border px-2 py-1 text-[10px] font-medium ${tone.chip}" style="--chip-order:${index}">
          <span class="w-1.5 h-1.5 rounded-full ${tone.dot}"></span>
          <span class="material-symbols-outlined text-[11px] leading-none opacity-90">${icon}</span>
          <span class="text-white/95">${name}</span>
          <span class="opacity-85">${label}</span>
        </span>`;
    })
    .join("");
};

const applyButtonState = (button, className, enabled) => {
  if (!(button instanceof HTMLElement)) {
    return;
  }
  button.className = className;
  button.disabled = !enabled;
};

const withButtonLoading = async (buttonLike, fallbackLabel, operation) => {
  const hasHTMLElement = typeof HTMLElement !== "undefined";
  const hasHTMLButtonElement = typeof HTMLButtonElement !== "undefined";
  const isButtonElement =
    hasHTMLButtonElement && buttonLike instanceof HTMLButtonElement;
  const isHTMLElement = hasHTMLElement && buttonLike instanceof HTMLElement;
  const button =
    isButtonElement
      ? buttonLike
      : isHTMLElement
        ? buttonLike.closest("button")
        : null;

  if (!(hasHTMLButtonElement && button instanceof HTMLButtonElement)) {
    return operation();
  }

  if (button.dataset.busy === "true") {
    return null;
  }

  const previousHTML = button.innerHTML;
  const previousDisabled = button.disabled;
  const previousAriaBusy = button.getAttribute("aria-busy");
  const loadingLabel =
    String(button.dataset.loadingLabel || fallbackLabel || "Processing...")
      .trim() || "Processing...";

  button.dataset.busy = "true";
  button.disabled = true;
  button.setAttribute("aria-busy", "true");
  button.innerHTML = `<span class="material-symbols-outlined text-[14px] animate-spin">progress_activity</span>${loadingLabel}`;

  try {
    return await operation();
  } finally {
    delete button.dataset.busy;
    if (!button.isConnected) {
      return;
    }
    button.disabled = previousDisabled;
    if (previousAriaBusy === null) {
      button.removeAttribute("aria-busy");
    } else {
      button.setAttribute("aria-busy", previousAriaBusy);
    }
    button.innerHTML = previousHTML;
  }
};

export const createGlobalServicesController = ({
  bridge,
  runtime,
  refs,
  getState,
  setState,
  onStatus,
  onToast,
}) => {
  const updateRefs = (nextRefs) => {
    refs = nextRefs;
  };

  let liveEnabled = false;
  let pollTimer = null;
  let rawLogOutput = "";

  const buildEmptyLogMessage = () => {
    const state = getState();
    const selectedID = String(state.selectedGlobalService || "")
      .trim()
      .toLowerCase();
    if (!selectedID) {
      return "Select a global service to view logs.";
    }
    if (selectedID === "dnsmasq") {
      return "DNSMasq is running but does not emit stdout logs by default.";
    }
    const selectedService = (state.globalServices || []).find(
      (item) => item.id === selectedID,
    );
    const serviceName = selectedService?.name || selectedID;
    return `No logs available for ${serviceName}.`;
  };

  const clearPoll = () => {
    if (pollTimer) {
      clearInterval(pollTimer);
      pollTimer = null;
    }
  };

  const renderLogOutput = () => {
    if (!refs.globalLogOutput) {
      return;
    }
    const trimmed = String(rawLogOutput || "").trim();
    refs.globalLogOutput.textContent = trimmed || buildEmptyLogMessage();
    refs.globalLogOutput.scrollTop = refs.globalLogOutput.scrollHeight;
  };

  const setActionFeedback = (message, tone = "info") => {
    const toneMap = {
      success: {
        icon: "check_circle",
        iconClass: "material-symbols-outlined text-[14px] text-primary",
        textClass: "rounded-xl border border-primary/25 bg-primary/10 px-3 py-2.5 flex items-center gap-2 text-xs text-primary/95",
      },
      warning: {
        icon: "warning",
        iconClass: "material-symbols-outlined text-[14px] text-amber-300",
        textClass: "rounded-xl border border-amber-500/30 bg-amber-500/10 px-3 py-2.5 flex items-center gap-2 text-xs text-amber-200",
      },
      error: {
        icon: "error",
        iconClass: "material-symbols-outlined text-[14px] text-red-300",
        textClass: "rounded-xl border border-red-500/30 bg-red-500/10 px-3 py-2.5 flex items-center gap-2 text-xs text-red-200",
      },
      info: {
        icon: "info",
        iconClass: "material-symbols-outlined text-[14px] text-[#90cba4]",
        textClass: "rounded-xl border border-[#2e573a] bg-[#0f2015]/80 px-3 py-2.5 flex items-center gap-2 text-xs text-slate-300",
      },
    };
    const toneConfig = toneMap[tone] || toneMap.info;

    if (refs.globalActionFeedback) {
      refs.globalActionFeedback.className = toneConfig.textClass;
      refs.globalActionFeedback.classList.remove("global-feedback-ping");
      if (typeof requestAnimationFrame === "function") {
        requestAnimationFrame(() => {
          if (refs.globalActionFeedback) {
            refs.globalActionFeedback.classList.add("global-feedback-ping");
          }
        });
      } else {
        refs.globalActionFeedback.classList.add("global-feedback-ping");
      }
    }
    if (refs.globalActionFeedbackIcon) {
      refs.globalActionFeedbackIcon.className = toneConfig.iconClass;
      refs.globalActionFeedbackIcon.textContent = toneConfig.icon;
    }
    setText(refs.globalActionFeedbackText, message || "Ready for global operations.");
  };

  const syncBulkActionButtons = (snapshot) => {
    const total = Number(snapshot.total || 0);
    const active = Number(snapshot.active || 0);
    const allRunning = total > 0 && active >= total;
    const anyRunning = active > 0;

    applyButtonState(refs.globalBulkStart, allRunning ? BULK_START_DISABLED_CLASS : BULK_START_ENABLED_CLASS, !allRunning);
    applyButtonState(refs.globalBulkRestart, anyRunning ? BULK_RESTART_ENABLED_CLASS : BULK_RESTART_DISABLED_CLASS, anyRunning);
    applyButtonState(refs.globalBulkStop, anyRunning ? BULK_STOP_ENABLED_CLASS : BULK_STOP_DISABLED_CLASS, anyRunning);
    applyButtonState(refs.globalBulkPull, BULK_PULL_CLASS, true);
  };

  const renderSummary = (snapshot) => {
    const services = Array.isArray(snapshot.services) ? snapshot.services : [];
    const total = Number(snapshot.total || services.length);
    const active = Number(
      snapshot.active || services.filter((service) => isServiceActive(service)).length,
    );
    const runningSafe = Math.max(0, Math.min(active, total || active));
    const percent = total > 0 ? Math.round((runningSafe / total) * 100) : 0;
    const hasRoutingWarning = services.some((service) => hasRoutingImpact(service));
    const offlineServices = Math.max(total - runningSafe, 0);
    const defaultSummary =
      total > 0
        ? `${runningSafe}/${total} global services running`
        : "No global services detected";

    setText(
      refs.globalServicesSummary,
      snapshot.summary || defaultSummary,
    );
    setText(refs.globalServiceCount, `${runningSafe}/${total} running`);
    setText(refs.globalServiceHealthPercent, `${percent}%`);

    if (refs.globalServiceHealthBar instanceof HTMLElement) {
      refs.globalServiceHealthBar.style.width = `${percent}%`;
      refs.globalServiceHealthBar.className =
        hasRoutingWarning || percent < 35
          ? "h-full rounded-full bg-gradient-to-r from-red-500 via-red-400 to-amber-300 transition-all duration-500"
          : percent < 100
            ? "h-full rounded-full bg-gradient-to-r from-amber-500 via-amber-300 to-primary transition-all duration-500"
            : "h-full rounded-full bg-gradient-to-r from-primary/70 via-primary to-[#7dffad] transition-all duration-500";
    }

    if (refs.globalServiceHealthLabel instanceof HTMLElement) {
      if (hasRoutingWarning) {
        refs.globalServiceHealthLabel.className =
          "mt-2 inline-flex items-center gap-1.5 rounded-md border border-amber-500/35 bg-amber-500/10 px-2 py-1 text-[10px] font-semibold text-amber-200";
        setText(refs.globalServiceHealthLabelIcon, "warning");
        setText(refs.globalServiceHealthLabelText, "Routing degraded");
      } else if (percent >= 100 && total > 0) {
        refs.globalServiceHealthLabel.className =
          "mt-2 inline-flex items-center gap-1.5 rounded-md border border-primary/25 bg-primary/10 px-2 py-1 text-[10px] font-semibold text-primary";
        setText(refs.globalServiceHealthLabelIcon, "task_alt");
        setText(refs.globalServiceHealthLabelText, "All systems nominal");
      } else if (runningSafe === 0 && total > 0) {
        refs.globalServiceHealthLabel.className =
          "mt-2 inline-flex items-center gap-1.5 rounded-md border border-red-500/35 bg-red-500/10 px-2 py-1 text-[10px] font-semibold text-red-200";
        setText(refs.globalServiceHealthLabelIcon, "error");
        setText(refs.globalServiceHealthLabelText, "Service mesh offline");
      } else {
        refs.globalServiceHealthLabel.className =
          "mt-2 inline-flex items-center gap-1.5 rounded-md border border-amber-500/35 bg-amber-500/10 px-2 py-1 text-[10px] font-semibold text-amber-200";
        setText(refs.globalServiceHealthLabelIcon, "monitor_heart");
        setText(
          refs.globalServiceHealthLabelText,
          `${offlineServices} service${offlineServices === 1 ? "" : "s"} need attention`,
        );
      }
    }

    renderStatusStrip(refs.globalServiceStatusStrip, services);
    syncBulkActionButtons({ active: runningSafe, total });

    if (hasRoutingWarning) {
      setActionFeedback(
        "Routing guard triggered: Caddy Proxy or DNSMasq is stopped.",
        "warning",
      );
    } else if (percent >= 100 && total > 0) {
      setActionFeedback(
        "All global services are healthy. Use Restart All for safe rolling refresh.",
        "success",
      );
    } else if (total > 0) {
      setActionFeedback(
        `${offlineServices} service${offlineServices === 1 ? "" : "s"} are offline. Start All can recover quickly.`,
        "warning",
      );
    } else {
      setActionFeedback("Global services are not available yet.", "info");
    }
  };

  const renderSnapshot = () => {
    const state = getState();
    const services = state.globalServices || [];
    const selected = state.selectedGlobalService || "";
    renderGlobalServices(refs.globalServicesList, services, selected);
    const selectedService = services.find((item) => item.id === selected);
    setText(
      refs.globalLogServiceName,
      selectedService ? selectedService.name : "Select service",
    );
  };

  const ensureSelectedService = () => {
    const state = getState();
    const services = state.globalServices || [];
    if (!services.length) {
      setState({ selectedGlobalService: "" });
      return "";
    }
    const selected = state.selectedGlobalService || "";
    if (services.some((item) => item.id === selected)) {
      return selected;
    }
    const preferred = services.find((item) => isServiceActive(item)) || services[0];
    setState({ selectedGlobalService: preferred.id });
    return preferred.id;
  };

  const refresh = async ({ silent = false } = {}) => {
    if (!silent && refs.globalServicesList) {
      refs.globalServicesList.innerHTML = `
        <div class="rounded-xl border border-dashed border-[#2e573a] bg-[#102316] p-4 text-sm text-slate-400">Loading global services...</div>
      `;
    }
    if (!silent) {
      setActionFeedback("Refreshing global services snapshot...", "info");
    }
    try {
      const snapshot = normalizeGlobalServicesSnapshot(
        await bridge.getGlobalServices(),
      );
      setState({ globalServices: snapshot.services });
      ensureSelectedService();
      renderSummary(snapshot);
      renderSnapshot();
      if (snapshot.warnings?.length) {
        onStatus(`Global services warnings: ${snapshot.warnings.join(" | ")}`);
      }
      return snapshot;
    } catch (err) {
      onStatus(`Failed to load global services: ${err}`);
      setActionFeedback(`Failed to load global services: ${err}`, "error");
      if (refs.globalServicesList) {
        refs.globalServicesList.innerHTML = `
          <div class="rounded-xl border border-dashed border-red-500/40 bg-red-500/10 p-4 text-sm text-red-300">
            Failed to load global services.
          </div>
        `;
      }
      return null;
    }
  };

  const appendLogLine = (line) => {
    const value = String(line || "").trim();
    if (!value) {
      return;
    }
    rawLogOutput = rawLogOutput ? `${rawLogOutput}\n${value}` : value;
    renderLogOutput();
  };

  const refreshLogs = async () => {
    const serviceID = String(getState().selectedGlobalService || "").trim();
    if (!serviceID) {
      rawLogOutput = "";
      renderLogOutput();
      return;
    }
    if (refs.globalLogOutput) {
      refs.globalLogOutput.textContent = "Loading logs...";
    }
    try {
      rawLogOutput = String(await bridge.getGlobalServiceLogs(serviceID, 300) || "");
      renderLogOutput();
    } catch (err) {
      rawLogOutput = `Failed to load logs: ${err}`;
      renderLogOutput();
    }
  };

  const stopLive = async ({ skipBridge = false } = {}) => {
    liveEnabled = false;
    setState({ globalLiveLogsEnabled: false });
    clearPoll();
    setText(refs.globalToggleLive, "Live: Off");
    setActionFeedback("Live log stream paused.", "info");
    if (skipBridge) {
      return;
    }
    try {
      await bridge.stopGlobalServiceLogStream();
    } catch (_err) {
      // Ignore: polling fallback mode may not hold stream state.
    }
  };

  const startLive = async () => {
    const serviceID = String(getState().selectedGlobalService || "").trim();
    if (!serviceID) {
      onStatus("Select a global service to stream logs.");
      setActionFeedback("Select a service first to stream logs.", "warning");
      return;
    }
    liveEnabled = true;
    setState({ globalLiveLogsEnabled: true });
    setText(refs.globalToggleLive, "Live: On");
    const serviceName =
      getState().globalServices.find((item) => item.id === serviceID)?.name ||
      serviceID;
    setActionFeedback(`Streaming live logs for ${serviceName}.`, "info");

    if (bridge.startGlobalServiceLogStream && runtime?.EventsOn) {
      try {
        await bridge.startGlobalServiceLogStream(serviceID);
        return;
      } catch (_err) {
        // Fall through to polling fallback.
      }
    }

    await refreshLogs();
    clearPoll();
    pollTimer = setInterval(refreshLogs, 2000);
  };

  const toggleLive = async () => {
    if (liveEnabled) {
      await stopLive();
      return;
    }
    await startLive();
  };

  const selectService = async (serviceID) => {
    const normalized = String(serviceID || "").trim().toLowerCase();
    if (!normalized) {
      return;
    }
    const state = getState();
    if (!state.globalServices.some((item) => item.id === normalized)) {
      return;
    }
    setState({ selectedGlobalService: normalized });
    renderSnapshot();
    if (liveEnabled) {
      await stopLive();
      await startLive();
      return;
    }
    await refreshLogs();
  };

  const runServiceAction = async (action, serviceID, triggerButton = null) => {
    const normalized = String(serviceID || "").trim().toLowerCase();
    if (!normalized) {
      return;
    }

    const actions = {
      start: bridge.startGlobalService,
      stop: bridge.stopGlobalService,
      restart: bridge.restartGlobalService,
      open: bridge.openGlobalService,
    };
    const fn = actions[action];
    if (!fn) {
      return;
    }

    const loadingLabelByAction = {
      start: "Starting...",
      stop: "Stopping...",
      restart: "Restarting...",
      open: "Opening...",
    };
    const actionVerbByAction = {
      start: "Starting",
      stop: "Stopping",
      restart: "Restarting",
      open: "Opening",
    };
    const serviceName =
      getState().globalServices.find((item) => item.id === normalized)?.name ||
      normalized;

    await withButtonLoading(
      triggerButton,
      loadingLabelByAction[action],
      async () => {
        setActionFeedback(
          `${actionVerbByAction[action] || "Processing"} ${serviceName}...`,
          "info",
        );
        try {
          const message = await fn(normalized);
          onStatus(message || `${action} ${normalized} completed`);
          onToast(message || `${action} ${normalized} completed`, "success");
          await refresh({ silent: true });
          setActionFeedback(message || `${serviceName} ${action} completed.`, "success");
        } catch (err) {
          onStatus(`${action} ${normalized} failed: ${err}`);
          onToast(`${action} ${normalized} failed: ${err}`, "error");
          setActionFeedback(
            `${actionVerbByAction[action] || "Action"} ${serviceName} failed: ${err}`,
            "error",
          );
        }
      },
    );
  };

  const runBulkAction = async (action, triggerButton = null) => {
    const actions = {
      start: bridge.startGlobalServices,
      stop: bridge.stopGlobalServices,
      restart: bridge.restartGlobalServices,
      pull: bridge.pullGlobalServices,
    };
    const fn = actions[action];
    if (!fn) {
      return;
    }

    const loadingLabelByAction = {
      start: "Starting All...",
      stop: "Stopping All...",
      restart: "Restarting All...",
      pull: "Pulling All...",
    };
    const actionVerbByAction = {
      start: "Starting",
      stop: "Stopping",
      restart: "Restarting",
      pull: "Pulling",
    };

    await withButtonLoading(triggerButton, loadingLabelByAction[action], async () => {
      setActionFeedback(
        `${actionVerbByAction[action] || "Processing"} all global services...`,
        "info",
      );
      try {
        const message = await fn();
        onStatus(message || `Global ${action} completed`);
        onToast(message || `Global ${action} completed`, "success");
        await refresh({ silent: true });
        setActionFeedback(
          message || `Global ${action} completed successfully.`,
          "success",
        );
      } catch (err) {
        onStatus(`Global ${action} failed: ${err}`);
        onToast(`Global ${action} failed: ${err}`, "error");
        setActionFeedback(`Global ${action} failed: ${err}`, "error");
      }
    });
  };

  const clearLogs = async () => {
    await stopLive();
    rawLogOutput = "";
    renderLogOutput();
    onStatus("Global service logs cleared.");
    setActionFeedback("Global service logs cleared.", "info");
  };

  if (runtime?.EventsOn) {
    runtime.EventsOn("global-logs:line", appendLogLine);
    runtime.EventsOn("global-logs:status", (message) => {
      onStatus(String(message || "").trim());
    });
    runtime.EventsOn("global-logs:error", (message) => {
      const text = String(message || "").trim();
      if (!text) {
        return;
      }
      onStatus(text);
      onToast(text, "error");
    });
  }

  return {
    refresh,
    refreshLogs,
    stopLive,
    toggleLive,
    clearLogs,
    selectService,
    runServiceAction,
    runBulkAction,
    updateRefs,
  };
};
