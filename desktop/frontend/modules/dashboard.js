import { clearChildren, escapeHTML, setText } from "../utils/dom.js";

export const createDashboardController = ({ bridge, refs, onStatus }) => {
  const updateRefs = (newRefs) => {
    refs = newRefs;
  };
};

export const normalizeDashboardPayload = (data = {}) => ({
  active: data.ActiveEnvironments ?? data.active ?? 0,
  services: data.RunningServices ?? data.services ?? 0,
  queued: data.QueuedTasks ?? data.queued ?? 0,
  activeSummary: data.ActiveSummary ?? data.activeSummary ?? "",
  servicesSummary: data.ServicesSummary ?? data.servicesSummary ?? "",
  queueSummary: data.QueueSummary ?? data.queueSummary ?? "",
  environments: Array.isArray(data.Environments)
    ? data.Environments
    : Array.isArray(data.environments)
      ? data.environments
      : [],
  warnings: Array.isArray(data.Warnings)
    ? data.Warnings
    : Array.isArray(data.warnings)
      ? data.warnings
      : [],
});

export const projectKey = (env = {}) =>
  env.Project || env.project || env.Name || env.name || "";

export const domainLabel = (env = {}) =>
  env.Domain || env.domain || env.Name || env.name || projectKey(env);

const withScheme = (value) => {
  const raw = String(value || "").trim();
  if (!raw) {
    return "";
  }
  if (/^https?:\/\//i.test(raw)) {
    return raw;
  }

  const host = raw.split("/")[0].trim();
  const isLoopback = /^(localhost|127\.0\.0\.1|\[::1\])(?::\d+)?$/i.test(host);
  const scheme = isLoopback ? "http" : "https";
  return `${scheme}://${raw.replace(/^\/+/, "")}`;
};

export const localEnvironmentURL = (env = {}) => {
  const explicitURL =
    env.LocalURL ||
    env.localURL ||
    env.URL ||
    env.Url ||
    env.url ||
    "";
  const explicitResolved = withScheme(explicitURL);
  if (explicitResolved) {
    return explicitResolved;
  }

  const candidate = String(
    env.Domain ||
      env.domain ||
      env.Name ||
      env.name ||
      env.Project ||
      env.project ||
      "",
  ).trim();
  if (!candidate) {
    return "";
  }

  let host = candidate;
  if (
    !/^https?:\/\//i.test(host) &&
    !host.includes(".") &&
    !host.includes(":")
  ) {
    host = `${host}.test`;
  }

  return withScheme(host);
};

export const serviceTargets = (env = {}) => {
  const values = Array.isArray(env.ServiceTargets)
    ? env.ServiceTargets
    : Array.isArray(env.serviceTargets)
      ? env.serviceTargets
      : [];
  return values.length ? values : ["web"];
};

export const renderMetricSkeletons = (refs) => {
  const skeleton = `<div class="h-6 w-12 skeleton mb-1"></div>`;
  const hintSkeleton = `<div class="h-3 w-24 skeleton"></div>`;
  if (refs.statActive) refs.statActive.innerHTML = skeleton;
  if (refs.statServices) refs.statServices.innerHTML = skeleton;
  if (refs.statQueue) refs.statQueue.innerHTML = skeleton;
  if (refs.statActiveHint) refs.statActiveHint.innerHTML = hintSkeleton;
  if (refs.statServicesHint) refs.statServicesHint.innerHTML = hintSkeleton;
  if (refs.statQueueHint) refs.statQueueHint.innerHTML = hintSkeleton;
};

export const setMetricText = (
  { active, services, queued, activeSummary, servicesSummary, queueSummary },
  refs,
) => {
  setText(refs.statActive, String(active));
  setText(refs.statServices, String(services));
  setText(refs.statQueue, String(queued));
  setText(refs.statActiveHint, activeSummary || "No environments detected");
  setText(refs.statServicesHint, servicesSummary || "Waiting for service data");
  setText(refs.statQueueHint, queueSummary || "Queue idle");
};

export const renderWarnings = (warningList, warnings = []) => {
  if (!warningList) {
    return;
  }
  clearChildren(warningList);
  warnings.forEach((warning) => {
    const item = document.createElement("li");
    item.textContent = String(warning);
    warningList.appendChild(item);
  });
};

const ACTIVE_ENVIRONMENT_STATUSES = new Set([
  "running",
  "warning",
  "healthy",
  "up",
  "starting",
  "restarting",
  "booting",
  "syncing",
]);

const environmentServices = (env = {}) => {
  if (Array.isArray(env.Services)) return env.Services;
  if (Array.isArray(env.services)) return env.services;
  return [];
};

const classifyEnvironmentStatus = (env = {}) => {
  const status = String(env.Status || env.status || "stopped").toLowerCase();
  const active = ACTIVE_ENVIRONMENT_STATUSES.has(status);
  const services = environmentServices(env);
  const serviceCount = services.length;

  const meta = {
    status,
    active,
    iconName: "stop_circle",
    iconClass: "text-slate-500",
    iconStyle: "",
    detailClass: "text-slate-500",
    dotClass: "bg-slate-500",
    detailText: "Stopped",
    showPulseDot: false,
  };

  if (status === "running" || status === "healthy" || status === "up") {
    meta.iconName = "play_circle";
    meta.iconClass = "text-primary fill-1";
    meta.iconStyle = "style=\"font-variation-settings: 'FILL' 1;\"";
    meta.detailClass = "text-primary";
    meta.dotClass = "bg-primary";
    meta.detailText =
      serviceCount > 0 ? `Running • ${serviceCount} services` : "Running";
    meta.showPulseDot = true;
    return meta;
  }

  if (
    status === "restarting" ||
    status === "starting" ||
    status === "booting" ||
    status === "syncing"
  ) {
    meta.iconName = "sync";
    meta.iconClass = "text-blue-400";
    meta.detailClass = "text-blue-400";
    meta.dotClass = "bg-blue-400";
    meta.detailText = status === "starting" ? "Starting..." : "Restarting...";
    return meta;
  }

  if (status === "warning") {
    meta.iconName = "warning";
    meta.iconClass = "text-amber-500";
    meta.detailClass = "text-amber-500";
    meta.dotClass = "bg-amber-500";
    meta.detailText = "Warning";
    return meta;
  }

  return meta;
};

const renderEnvironmentItem = (env, { selectedProject, sidebarMode }) => {
  const key = projectKey(env);
  const domain = domainLabel(env);
  const meta = classifyEnvironmentStatus(env);
  const isSelected = sidebarMode === "environments" && key === selectedProject;

  const baseClass =
    "w-full mb-1 flex items-center gap-3 px-3 py-2.5 rounded-lg border transition-all relative overflow-hidden text-left";
  let stateClass = meta.active
    ? "border-transparent hover:bg-[#173325]/40 hover:border-primary/20"
    : "border-transparent hover:bg-[#16241b]/80 hover:border-[#2e573a] opacity-60";

  if (isSelected) {
    stateClass = "bg-[#173325] border-primary/35 shadow-[0_0_14px_rgba(13,242,89,0.12)] opacity-100";
  }

  const selectionIndicator = isSelected
    ? `<div class="absolute inset-y-0 left-0 w-1 bg-primary"></div>`
    : "";

  const titleClass = meta.active ? "text-white" : "text-slate-400";

  return `
    <button data-action="select-environment" data-env="${escapeHTML(key)}" class="${baseClass} ${stateClass}" title="Select ${escapeHTML(domain)}">
      ${selectionIndicator}
      <div class="relative shrink-0 z-10">
        <span data-action="toggle-env" data-env="${escapeHTML(key)}" class="material-symbols-outlined ${meta.iconClass} transition-colors hover:text-white text-[20px]" ${meta.iconStyle}>${meta.iconName}</span>
        ${
          meta.showPulseDot
            ? `<span class="absolute -top-0.5 -right-0.5 w-2 h-2 rounded-full bg-primary border border-[#0c1810] animate-pulse"></span>`
            : ""
        }
      </div>
      <div class="min-w-0 pointer-events-none">
        <div class="text-sm font-semibold truncate ${titleClass}">${escapeHTML(domain)}</div>
        <div class="text-[11px] ${meta.detailClass} flex items-center gap-1">
          <span class="w-1 h-1 rounded-full ${meta.dotClass}"></span>
          <span>${escapeHTML(meta.detailText)}</span>
        </div>
      </div>
    </button>
  `;
};

export const renderEnvironmentSkeletons = (container) => {
  if (!container) return;
  const globalRow = `
    <div class="w-full mt-3 mb-4 flex items-center gap-3 px-3 py-3 rounded-xl border border-[#2e573a] bg-[#13261a]">
      <div class="h-8 w-8 rounded-lg skeleton"></div>
      <div class="flex-1 space-y-2">
        <div class="h-3 w-28 skeleton"></div>
        <div class="h-2 w-20 skeleton"></div>
      </div>
    </div>
  `;
  const activeHeader = `<div class="px-1 mt-4 pb-4 text-[10px] font-semibold text-primary/80 uppercase tracking-[0.12em]">Active Environments</div>`;
  const items = Array(3)
    .fill(0)
    .map(
      () => `
    <div class="w-full mb-1 flex items-center gap-3 px-3 py-2.5 rounded-lg border border-transparent">
      <div class="h-5 w-5 rounded-full skeleton"></div>
      <div class="flex-1 space-y-2">
        <div class="h-3 w-24 skeleton"></div>
        <div class="h-2 w-12 skeleton"></div>
      </div>
    </div>
  `,
    )
    .join("");
  const inactiveHeader = `<div class="px-1 mt-4 pb-4 text-[10px] font-semibold text-slate-500 uppercase tracking-[0.12em]">Inactive Environments</div>`;
  container.innerHTML =
    globalRow +
    activeHeader +
    items +
    inactiveHeader +
    items;
};

export const renderEnvironmentList = (
  container,
  environments = [],
  selectedProject = "",
  options = {},
) => {
  if (!container) {
    return;
  }
  const sidebarMode =
    options.sidebarMode === "global-services"
      ? "global-services"
      : "environments";

  const globalSelected = sidebarMode === "global-services";
  const globalClass = globalSelected
    ? "w-full mt-3 mb-4 text-left p-3 rounded-xl bg-[#173325] border-l-4 border-primary border border-primary/25 transition-all relative overflow-hidden shadow-[0_0_16px_rgba(13,242,89,0.1)]"
    : "w-full mt-3 mb-4 text-left p-3 rounded-xl bg-[#0f1d15] border border-[#1f3d2a]/80 hover:bg-[#13261a] hover:border-primary/20 transition-all relative overflow-hidden group";
  const globalIndicator = globalSelected
    ? `<div class="absolute inset-y-0 left-0 w-1 bg-primary/80"></div>`
    : "";
  const globalIconWrapClass = globalSelected
    ? "h-9 w-9 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center text-primary"
    : "h-9 w-9 rounded-lg bg-[#13261a] border border-[#2e573a]/80 flex items-center justify-center text-[#90cba4]/70 group-hover:text-primary transition-colors";
  const globalTitleClass = globalSelected
    ? "text-white text-sm font-semibold truncate w-full text-left"
    : "text-[#d5e8dd] text-sm font-semibold truncate w-full text-left group-hover:text-white transition-colors";
  const globalSubtitleClass = globalSelected
    ? "text-xs text-[#90cba4]/60"
    : "text-xs text-[#90cba4]/40 group-hover:text-[#90cba4]/60 transition-colors";
  const globalRow = `
    <button data-action="switch-sidebar-mode" data-mode="global-services" class="${globalClass}" title="Open Global Services">
      ${globalIndicator}
      <div class="flex items-center gap-3">
        <div class="${globalIconWrapClass}">
          <span class="material-symbols-outlined text-[20px]">hub</span>
        </div>
        <div class="flex flex-col items-start min-w-0 pointer-events-none">
          <span class="${globalTitleClass}">Global Services</span>
          <span class="${globalSubtitleClass}">Shared system services</span>
        </div>
      </div>
    </button>
  `;

  const active = [];
  const inactive = [];
  environments.forEach((env) => {
    const status = classifyEnvironmentStatus(env);
    if (status.active) {
      active.push(env);
    } else {
      inactive.push(env);
    }
  });

  const renderGroup = (
    title,
    items,
    emptyText,
    tone = "text-slate-400",
    spacingClass = "mt-4 pb-4",
  ) => `
    <div class="px-1 ${spacingClass} text-[10px] font-semibold ${tone} uppercase tracking-[0.12em]">${title}</div>
    ${
      items.length
        ? items
            .map((env) =>
              renderEnvironmentItem(env, { selectedProject, sidebarMode }),
            )
            .join("")
        : `<div class="px-3 py-4 mb-1 text-xs text-slate-500">${emptyText}</div>`
    }
  `;

  if (!environments.length) {
    container.innerHTML =
      globalRow +
      renderGroup(
        "Active Environments",
        [],
        "No active environments.",
        "text-primary/80",
      ) +
      renderGroup(
        "Inactive Environments",
        [],
        "No environments detected.",
        "text-slate-400",
        "mt-4 pb-4",
      );
    return;
  }

  container.innerHTML =
    globalRow +
    renderGroup(
      "Active Environments",
      active,
      "No active environments.",
      "text-primary/80",
    ) +
    renderGroup(
      "Inactive Environments",
      inactive,
      "No inactive environments.",
      "text-slate-400",
      "mt-4 pb-4",
    );
};

const syncSingleSelector = (selector, environments, selectedProject) => {
  if (!selector) {
    return;
  }
  const previous = selectedProject || selector.value;
  selector.innerHTML = "";
  environments.forEach((env) => {
    const option = document.createElement("option");
    option.value = projectKey(env);
    option.textContent = domainLabel(env);
    selector.appendChild(option);
  });
  const exists = environments.some((env) => projectKey(env) === previous);
  selector.value = exists
    ? previous
    : environments.length
      ? projectKey(environments[0])
      : "";
};

export const syncProjectSelectors = (
  selectors,
  environments = [],
  selectedProject = "",
) => {
  syncSingleSelector(selectors.envSelector, environments, selectedProject);
  syncSingleSelector(selectors.logSelector, environments, selectedProject);
};
export const renderProjectHero = (
  refs,
  environments = [],
  selectedProject = "",
) => {
  const env = environments.find((e) => projectKey(e) === selectedProject);
  if (!env) return;

  const title = domainLabel(env);
  const status = String(env.Status || env.status || "stopped").toLowerCase();
  const url = localEnvironmentURL(env);

  if (refs.projectTitle) refs.projectTitle.textContent = title;
  if (refs.projectStatusText) {
    refs.projectStatusText.textContent =
      status.charAt(0).toUpperCase() + status.slice(1);
  }

  if (refs.projectStatusBadge) {
    const badge = refs.projectStatusBadge;
    const dot = badge.querySelector('[data-role="project-status-dot"]');
    badge.className =
      "px-3 py-1 rounded-full border text-xs font-bold uppercase tracking-wide neon-glow flex items-center gap-1.5";
    if (dot instanceof HTMLElement) {
      dot.className = "w-2 h-2 rounded-full";
    }
    if (status === "running") {
      badge.classList.add("bg-primary/20", "border-primary/30", "text-primary");
      if (dot instanceof HTMLElement) {
        dot.classList.add(
          "bg-primary",
          "shadow-[0_0_12px_rgba(13,242,89,0.9)]",
          "animate-pulse",
        );
      }
    } else if (status === "warning") {
      badge.classList.add(
        "bg-amber-500/20",
        "border-amber-500/30",
        "text-amber-500",
      );
      if (dot instanceof HTMLElement) {
        dot.classList.add("bg-amber-500");
      }
    } else {
      badge.classList.add(
        "bg-slate-500/20",
        "border-slate-500/30",
        "text-slate-400",
      );
      if (dot instanceof HTMLElement) {
        dot.classList.add("bg-slate-500");
      }
    }
  }

  if (refs.projectUrl) {
    refs.projectUrl.href = url || "#";
    refs.projectUrl.dataset.action = "env-open";
    refs.projectUrl.dataset.env = selectedProject;
    if (refs.projectUrl.classList) {
      if (url) {
        refs.projectUrl.classList.remove("hidden");
      } else {
        refs.projectUrl.classList.add("hidden");
      }
    }
  }
  if (refs.projectUrlText) {
    refs.projectUrlText.textContent = url;
  }

  if (refs.projectTechnologies) {
    const techs = Array.isArray(env.Technologies)
      ? env.Technologies
      : Array.isArray(env.technologies)
        ? env.technologies
        : [];

    if (techs.length) {
      refs.projectTechnologies.innerHTML = techs
        .map((tech) => {
          let color = "bg-blue-500";
          let shadow = "rgba(59, 130, 246, 0.5)";
          if (
            tech.toLowerCase().includes("mysql") ||
            tech.toLowerCase().includes("maria")
          ) {
            color = "bg-yellow-500";
            shadow = "rgba(234, 179, 8, 0.5)";
          }
          if (
            tech.toLowerCase().includes("redis") ||
            tech.toLowerCase().includes("cache")
          ) {
            color = "bg-red-500";
            shadow = "rgba(239, 68, 68, 0.5)";
          }
          if (tech.toLowerCase().includes("python")) {
            color = "bg-green-600";
            shadow = "rgba(22, 163, 74, 0.5)";
          }
          if (tech.toLowerCase().includes("node")) {
            color = "bg-green-500";
            shadow = "rgba(34, 197, 94, 0.5)";
          }

          return `<span class="flex items-center gap-1 bg-[#1a3322] px-2 py-0.5 rounded border border-[#2e573a]">
          <span class="w-1.5 h-1.5 rounded-full ${color}" style="box-shadow: 0 0 8px ${shadow}"></span>
          ${escapeHTML(tech)}
        </span>`;
        })
        .join("");
    } else {
      refs.projectTechnologies.innerHTML = "";
    }
  }
  if (refs.heroRestartBtn) {
    const isStopped = status !== "running" && status !== "warning";
    const action = isStopped ? "env-start" : "env-restart";
    const label = isStopped ? "Start" : "Restart";
    const icon = isStopped ? "play_arrow" : "restart_alt";

    refs.heroRestartBtn.dataset.action = action;
    refs.heroRestartBtn.dataset.env = selectedProject;
    refs.heroRestartBtn.title = `${label} Environment`;
    refs.heroRestartBtn.innerHTML = `
      <span class="material-symbols-outlined text-lg">${icon}</span>
      ${label}
    `;
  }
  if (refs.heroStopBtn) {
    refs.heroStopBtn.dataset.env = selectedProject;
    const isStopped = status !== "running" && status !== "warning";
    refs.heroStopBtn.disabled = isStopped;
    refs.heroStopBtn.title = isStopped
      ? "Environment is not running"
      : "Stop Environment";
    refs.heroStopBtn.className = isStopped
      ? "h-12 w-12 bg-[#13261a] text-slate-500 border border-[#2e573a] rounded-lg transition-all flex items-center justify-center cursor-not-allowed opacity-70"
      : "h-12 w-12 bg-red-600 text-white border border-red-500 rounded-lg hover:bg-red-500 transition-all active:scale-95 flex items-center justify-center shadow-lg shadow-red-500/20";
  }

  if (refs.envVarsList) {
    renderEnvVars(refs.envVarsList, env);
  }

  const servicesContainer = document.getElementById("activeServicesList");
  if (servicesContainer) {
    renderActiveServices(servicesContainer, env);
  }
};

const inferServiceTarget = (service = {}) => {
  const explicit = String(service.Target || service.target || "")
    .trim()
    .toLowerCase();
  if (explicit) {
    return explicit;
  }

  const name = String(service.Name || service.name || "")
    .trim()
    .toLowerCase();

  if (name.includes("php")) return "php";
  if (
    name.includes("maria") ||
    name.includes("mysql") ||
    name.includes("postgres") ||
    name.includes("database") ||
    name.includes("db")
  ) {
    return "db";
  }
  if (name.includes("opensearch")) return "opensearch";
  if (name.includes("elastic")) return "elasticsearch";
  if (name.includes("redis")) return "redis";
  if (name.includes("valkey")) return "valkey";
  if (name.includes("rabbit")) return "rabbitmq";
  if (name.includes("varnish")) return "varnish";
  if (
    name.includes("nginx") ||
    name.includes("apache") ||
    name.includes("proxy") ||
    name.includes("web")
  ) {
    return "web";
  }

  return "web";
};

export const renderActiveServices = (container, env) => {
  if (!container) return;
  const project = projectKey(env);

  const services = Array.isArray(env?.Services)
    ? env.Services
    : Array.isArray(env?.services)
      ? env.services
      : [];

  if (services.length === 0) {
    container.innerHTML = `
      <div class="p-6 text-center text-slate-400 border border-dashed border-[#2e573a] rounded-xl bg-[#1a3322]/30">
        <span class="material-symbols-outlined text-3xl mb-2 opacity-20">inventory_2</span>
        <div class="text-sm italic">No active services detected</div>
      </div>`;
    return;
  }

  container.innerHTML = services
    .map((service) => {
      const status = String(
        service.Status || service.status || "stopped",
      ).toLowerCase();
      const serviceTarget = inferServiceTarget(service);
      const isHealthy =
        status === "healthy" || status === "running" || status === "up";
      const statusColor = isHealthy ? "text-green-400" : "text-amber-400";

      let icon = "bolt";
      let iconBg = "bg-blue-500/10";
      let iconText = "text-blue-400";
      let iconBorder = "border-blue-500/20";

      const name = String(
        service.Name || service.name || "unknown",
      ).toLowerCase();
      if (name.includes("php")) {
        icon = "php";
        iconBg = "bg-indigo-500/10";
        iconText = "text-indigo-400";
        iconBorder = "border-indigo-500/20";
      } else if (
        name.includes("mysql") ||
        name.includes("db") ||
        name.includes("maria")
      ) {
        icon = "database";
        iconBg = "bg-yellow-500/10";
        iconText = "text-yellow-400";
        iconBorder = "border-yellow-500/20";
      } else if (
        name.includes("nginx") ||
        name.includes("proxy") ||
        name.includes("web")
      ) {
        icon = "language";
        iconBg = "bg-green-500/10";
        iconText = "text-green-400";
        iconBorder = "border-green-500/20";
      }

      return `
        <div class="glass-panel p-4 rounded-xl border border-[#2e573a] hover:border-primary/30 transition-all flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div class="p-2 rounded ${iconBg} ${iconText} border ${iconBorder}">
              <span class="material-symbols-outlined">${icon}</span>
            </div>
            <div>
              <h4 class="text-white font-medium text-sm">${escapeHTML(service.Name || service.name || "Service")}</h4>
              <div class="flex items-center gap-2 text-xs mt-1">
                <span class="text-slate-400">Port: ${service.Port || service.port || "N/A"}</span>
                <span class="w-1 h-1 rounded-full bg-slate-600"></span>
                <span class="${statusColor}">${escapeHTML(service.Status || service.status || "Unknown")}</span>
              </div>
            </div>
          </div>
          <div class="flex items-center gap-3">
            <button
              class="p-1.5 rounded bg-[#13261a] border border-[#2e573a] text-slate-300 hover:text-white hover:bg-[#22492f] transition-colors"
              title="View Logs"
              data-action="open-service-logs"
              data-project="${escapeHTML(project)}"
              data-service="${escapeHTML(serviceTarget)}"
            >
              <span class="material-symbols-outlined text-lg">list_alt</span>
            </button>
            <button
              class="p-1.5 rounded bg-[#13261a] border border-[#2e573a] text-slate-300 hover:text-white hover:bg-[#22492f] transition-colors"
              title="Open Terminal"
              data-action="open-service-shell"
              data-project="${escapeHTML(project)}"
              data-service="${escapeHTML(serviceTarget)}"
            >
              <span class="material-symbols-outlined text-lg">terminal</span>
            </button>
          </div>
        </div>`;
    })
    .join("");
};

export const renderEnvVars = (container, env) => {
  if (!container) return;

  const envVars = env?.EnvVars || env?.envVars || {};
  const keys = Object.keys(envVars);

  if (keys.length === 0) {
    container.innerHTML = `<div class="text-xs text-slate-500 italic">No environment variables defined</div>`;
    return;
  }

  container.innerHTML = keys
    .map((key) => {
      const value = envVars[key];
      return `
      <div data-action="copy-text" data-text="${escapeHTML(value)}" class="flex justify-between items-center group cursor-pointer hover:bg-[#22492f]/50 p-1.5 -mx-1.5 rounded transition-colors" title="Click to copy">
        <span class="text-xs text-[#90cba4] font-mono">${escapeHTML(key)}</span>
        <span class="text-xs text-white font-mono bg-[#102316] px-2 py-0.5 rounded border border-[#2e573a] break-all max-w-[60%]">${escapeHTML(value)}</span>
      </div>`;
    })
    .join("");
};
