export const normalizeSettingsPayload = (settings = {}) => ({
  theme: settings.theme || settings.Theme || "system",
  proxyTarget: settings.proxyTarget || settings.ProxyTarget || "",
  preferredBrowser:
    settings.preferredBrowser || settings.PreferredBrowser || "",
  codeEditor: settings.codeEditor || settings.CodeEditor || "",
  dbClientPreference:
    settings.dbClientPreference || settings.DBClientPreference || "pma",
});

export const applyTheme = (theme) => {
  const root = document.documentElement;
  if (!root) {
    return;
  }
  const setDarkClass = (darkEnabled) => {
    if (darkEnabled) {
      root.classList.add("dark");
    } else {
      root.classList.remove("dark");
    }
  };
  if (theme === "system") {
    const prefersDark =
      window.matchMedia &&
      window.matchMedia("(prefers-color-scheme: dark)").matches;
    setDarkClass(prefersDark);
    return;
  }
  setDarkClass(theme === "dark");
};

export const createSettingsController = ({
  bridge,
  refs,
  onStatus,
  onToast,
}) => {
  const UPDATE_BADGE_BASE_CLASS =
    "inline-flex items-center rounded-full border px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.08em]";
  const UPDATE_BADGE_IDLE_CLASS =
    "border-[#2e573a] bg-[#173325]/70 text-[#90cba4]";
  const UPDATE_BADGE_WORKING_CLASS =
    "border-blue-500/35 bg-blue-500/15 text-blue-200";
  const UPDATE_BADGE_AVAILABLE_CLASS =
    "border-amber-500/35 bg-amber-500/15 text-amber-200";
  const UPDATE_BADGE_CURRENT_CLASS = "border-primary/35 bg-primary/15 text-primary";
  const UPDATE_BADGE_INSTALLED_CLASS =
    "border-primary/45 bg-primary/20 text-primary";
  const UPDATE_BADGE_ERROR_CLASS = "border-red-500/35 bg-red-500/15 text-red-200";

  const updateState = {
    checked: false,
    checking: false,
    installing: false,
    outdated: false,
    installCompleted: false,
    failed: false,
    message: "Version check has not been run yet.",
  };

  const normalizeUpdateResult = (payload = {}) => ({
    currentVersion: String(
      payload.currentVersion || payload.CurrentVersion || "",
    ).trim(),
    latestVersion: String(payload.latestVersion || payload.LatestVersion || "").trim(),
    outdated:
      payload.outdated === undefined
        ? Boolean(payload.Outdated)
        : Boolean(payload.outdated),
    message: String(payload.message || payload.Message || "").trim(),
  });

  const normalizeErrorMessage = (err, fallback) => {
    const raw =
      typeof err?.message === "string"
        ? err.message
        : typeof err === "string"
          ? err
          : "";
    const message = raw.trim();
    return message ? message : fallback;
  };

  const normalizeCheckOptions = (options = {}) => ({
    silent:
      options && typeof options === "object" ? Boolean(options.silent) : false,
  });

  const renderUpdateSection = () => {
    if (refs.settingsUpdateStatus) {
      refs.settingsUpdateStatus.textContent = updateState.message;
    }

    if (refs.settingsUpdateBadge) {
      let badgeText = "Idle";
      let badgeToneClass = UPDATE_BADGE_IDLE_CLASS;

      if (updateState.failed) {
        badgeText = "Failed";
        badgeToneClass = UPDATE_BADGE_ERROR_CLASS;
      } else if (updateState.checking || updateState.installing) {
        badgeText = "Working";
        badgeToneClass = UPDATE_BADGE_WORKING_CLASS;
      } else if (updateState.installCompleted) {
        badgeText = "Installed";
        badgeToneClass = UPDATE_BADGE_INSTALLED_CLASS;
      } else if (updateState.outdated && updateState.checked) {
        badgeText = "Available";
        badgeToneClass = UPDATE_BADGE_AVAILABLE_CLASS;
      } else if (updateState.checked) {
        badgeText = "Current";
        badgeToneClass = UPDATE_BADGE_CURRENT_CLASS;
      }

      refs.settingsUpdateBadge.textContent = badgeText;
      refs.settingsUpdateBadge.className = `${UPDATE_BADGE_BASE_CLASS} ${badgeToneClass}`;
    }

    if (refs.checkUpdatesButton) {
      refs.checkUpdatesButton.disabled =
        updateState.checking || updateState.installing;
      refs.checkUpdatesButton.innerHTML = updateState.checking
        ? '<span class="material-symbols-outlined text-[18px]">progress_activity</span><span>Checking...</span>'
        : '<span class="material-symbols-outlined text-[18px]">sync</span><span>Check for updates</span>';
    }

    if (refs.installUpdateButton) {
      const shouldShow =
        updateState.outdated &&
        updateState.checked &&
        !updateState.installCompleted;
      refs.installUpdateButton.classList.toggle("hidden", !shouldShow);
      refs.installUpdateButton.disabled =
        updateState.checking || updateState.installing;
      refs.installUpdateButton.innerHTML = updateState.installing
        ? '<span class="material-symbols-outlined text-[18px]">install_desktop</span><span>Installing...</span>'
        : '<span class="material-symbols-outlined text-[18px]">download</span><span>Download & Install Update</span>';
    }
  };

  const updateRefs = (newRefs) => {
    refs = newRefs;
    renderUpdateSection();
  };
  const toggleDrawer = (open) => {
    if (!refs.settingsDrawer) {
      return;
    }
    if (open) {
      refs.settingsDrawer.classList.remove("hidden");
      refs.settingsDrawer.setAttribute("aria-hidden", "false");
      return;
    }
    refs.settingsDrawer.classList.add("hidden");
    refs.settingsDrawer.setAttribute("aria-hidden", "true");
  };

  const load = async () => {
    try {
      const raw = await bridge.getSettings();
      const settings = normalizeSettingsPayload(raw);
      if (refs.themeSelect) refs.themeSelect.value = settings.theme;
      if (refs.proxyTarget) refs.proxyTarget.value = settings.proxyTarget;
      if (refs.preferredBrowser)
        refs.preferredBrowser.value = settings.preferredBrowser;
      if (refs.codeEditor) refs.codeEditor.value = settings.codeEditor;
      if (refs.dbClientPreference)
        refs.dbClientPreference.value = settings.dbClientPreference;
      applyTheme(settings.theme);
      renderUpdateSection();
    } catch (_err) {
      applyTheme("system");
      renderUpdateSection();
    }
  };

  const save = async () => {
    const theme = refs.themeSelect?.value || "system";
    const proxyTarget = refs.proxyTarget?.value || "";
    const preferredBrowser = refs.preferredBrowser?.value || "";
    const codeEditor = refs.codeEditor?.value || "";
    const dbClientPreference = refs.dbClientPreference?.value || "pma";
    try {
      const message = await bridge.updateSettings({
        theme,
        proxyTarget,
        preferredBrowser,
        codeEditor,
        dbClientPreference,
      });
      applyTheme(theme);
      onStatus("Settings saved successfully.");
      onToast("Settings saved successfully.", "success");
    } catch (err) {
      const message = "Could not save settings.";
      onStatus(message);
      onToast(message, "error");
    }
  };

  const reset = async () => {
    try {
      const message = await bridge.resetSettings();
      onStatus("Settings reset to defaults.");
      onToast("Settings reset to defaults.", "success");
      await load();
    } catch (err) {
      const message = "Could not reset settings.";
      onStatus(message);
      onToast(message, "error");
    }
  };

  const checkForUpdates = async (options = {}) => {
    if (updateState.checking || updateState.installing) {
      return { skipped: true, reason: "busy" };
    }

    const { silent } = normalizeCheckOptions(options);

    updateState.checking = true;
    updateState.message = "Checking for latest version...";
    renderUpdateSection();

    try {
      const raw = await bridge.checkForUpdates();
      const result = normalizeUpdateResult(raw);

      updateState.checked = true;
      updateState.outdated = result.outdated;
      updateState.installCompleted = false;
      updateState.failed = false;

      if (result.message) {
        updateState.message = result.message;
      } else if (result.outdated) {
        updateState.message = `Update available: ${result.currentVersion} -> ${result.latestVersion}`;
      } else {
        updateState.message = `Govard Desktop is up to date (${result.currentVersion}).`;
      }

      if (!silent) {
        if (updateState.outdated) {
          onStatus("Update available.");
          onToast("A newer Govard version is available.", "info");
        } else {
          onStatus("Govard Desktop is up to date.");
          onToast("Govard Desktop is already up to date.", "success");
        }
      }
      return {
        skipped: false,
        failed: false,
        outdated: result.outdated,
        currentVersion: result.currentVersion,
        latestVersion: result.latestVersion,
        message: updateState.message,
      };
    } catch (_err) {
      updateState.checked = true;
      updateState.outdated = false;
      updateState.failed = true;
      updateState.message = normalizeErrorMessage(
        _err,
        "Could not check for updates.",
      );
      if (!silent) {
        onStatus(updateState.message);
        onToast(updateState.message, "error");
      }
      return {
        skipped: false,
        failed: true,
        outdated: false,
        currentVersion: "",
        latestVersion: "",
        message: updateState.message,
      };
    } finally {
      updateState.checking = false;
      renderUpdateSection();
    }
  };

  const installLatestUpdate = async () => {
    if (updateState.checking || updateState.installing) {
      return { ok: false, skipped: true, reason: "busy" };
    }

    updateState.installing = true;
    updateState.message = "Downloading and installing update...";
    renderUpdateSection();
    let updateInstalled = false;
    let restartFailed = false;

    try {
      await bridge.installLatestUpdate();
      updateState.outdated = false;
      updateState.installCompleted = true;
      updateState.failed = false;
      updateInstalled = true;
      updateState.message =
        "Update installed. Restart Govard Desktop to run the new version.";
      onStatus("Update installed. Restarting Govard Desktop...");
      onToast("Update installed. Restarting Govard Desktop...", "success");

      try {
        await bridge.restartDesktopApp();
      } catch (restartErr) {
        restartFailed = true;
        updateState.failed = true;
        updateState.message = normalizeErrorMessage(
          restartErr,
          "Update installed but automatic restart failed. Please restart manually.",
        );
        onStatus(updateState.message);
        onToast(updateState.message, "warning");
      }
    } catch (_err) {
      updateState.failed = true;
      updateState.message = normalizeErrorMessage(_err, "Automatic update failed.");
      onStatus(updateState.message);
      onToast(updateState.message, "error");
      return { ok: false, skipped: false, message: updateState.message };
    } finally {
      updateState.installing = false;
      renderUpdateSection();
    }

    return {
      ok: updateInstalled,
      skipped: false,
      restartFailed,
      message: updateState.message,
    };
  };

  return {
    toggleDrawer,
    load,
    save,
    reset,
    checkForUpdates,
    installLatestUpdate,
    updateRefs,
  };
};

export const renderSettingsDrawer = (container) => {
  if (!container) return;
  container.innerHTML = `
      <div
        class="drawer hidden fixed inset-0 z-[100] bg-[#0c1810]/40 backdrop-blur-sm"
        id="settingsDrawer"
        aria-hidden="true"
      >
        <div
          class="bg-[#1a3322] border-l border-[#2e573a] h-full ml-auto w-[420px] p-8 flex flex-col gap-6 shadow-2xl"
        >
          <div
            class="flex justify-between items-center mb-4 border-b border-[#2e573a] pb-4"
          >
            <h3 class="text-white text-lg font-bold">Govard Settings</h3>
            <button
              class="p-2 rounded hover:bg-white/5 text-slate-400 hover:text-white flex items-center justify-center h-9 w-9"
              id="closeSettings"
              type="button"
            >
              <span class="material-symbols-outlined">close</span>
            </button>
          </div>
          <!-- Settings fields -->
          <div class="flex flex-col gap-6 overflow-y-auto pr-2">
            <label
              class="flex flex-col gap-2 text-sm font-medium text-slate-300"
            >
              Theme
              <select
                id="themeSelect"
                class="bg-[#102316] border border-[#2e573a] rounded-lg px-4 py-3 text-white outline-none focus:border-primary/50"
              >
                <option value="system">System</option>
                <option value="light">Light</option>
                <option value="dark">Dark</option>
              </select>
            </label>
            <label
              class="flex flex-col gap-2 text-sm font-medium text-slate-300"
            >
              Proxy target
              <input
                id="proxyTarget"
                type="text"
                placeholder="govard.test"
                class="bg-[#102316] border border-[#2e573a] rounded-lg px-4 py-3 text-white outline-none focus:border-primary/50"
              />
            </label>
            <label
              class="flex flex-col gap-2 text-sm font-medium text-slate-300"
            >
              Code Editor (IDE)
              <input
                id="codeEditor"
                type="text"
                placeholder="code"
                class="bg-[#102316] border border-[#2e573a] rounded-lg px-4 py-3 text-white outline-none focus:border-primary/50"
              />
            </label>
            <label
              class="flex flex-col gap-2 text-sm font-medium text-slate-300"
            >
              Preferred browser
              <input
                id="preferredBrowser"
                type="text"
                placeholder="firefox"
                class="bg-[#102316] border border-[#2e573a] rounded-lg px-4 py-3 text-white outline-none focus:border-primary/50"
              />
            </label>
            <label
              class="flex flex-col gap-2 text-sm font-medium text-slate-300"
            >
              Database Client
              <select
                id="dbClientPreference"
                class="bg-[#102316] border border-[#2e573a] rounded-lg px-4 py-3 text-white outline-none focus:border-primary/50 appearance-none cursor-pointer"
              >
                <option value="pma">PHPMyAdmin (Proxy Container)</option>
                <option value="desktop">Local Client (e.g. BeeKeeper Studio)</option>
              </select>
            </label>
          </div>
          <div class="mt-auto flex flex-col gap-3">
            <div class="relative overflow-hidden rounded-xl border border-[#2e573a]/80 bg-[linear-gradient(155deg,rgba(13,242,89,0.12),rgba(13,242,89,0.03)_45%,rgba(12,24,16,0.96)_100%)] p-4 shadow-[0_10px_30px_rgba(0,0,0,0.28)]">
              <div class="pointer-events-none absolute -right-8 -top-8 h-24 w-24 rounded-full bg-primary/15 blur-2xl"></div>
              <div class="relative z-10">
                <div class="flex items-center justify-between gap-2">
                  <div class="flex items-center gap-2">
                    <span class="material-symbols-outlined text-primary text-[18px]"
                      >system_update_alt</span
                    >
                    <p class="text-sm font-semibold text-white">Updates</p>
                  </div>
                  <span
                    id="settingsUpdateBadge"
                    class="inline-flex items-center rounded-full border border-[#2e573a] bg-[#173325]/70 px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.08em] text-[#90cba4]"
                  >
                    Idle
                  </span>
                </div>
                <p
                  class="mt-2 text-xs leading-5 text-slate-300"
                  id="settingsUpdateStatus"
                  aria-live="polite"
                >
                Version check has not been run yet.
                </p>
                <div class="mt-3 flex flex-col gap-2">
                  <button
                    class="group inline-flex w-full items-center justify-center gap-2 rounded-xl border border-primary/35 bg-primary/10 px-4 py-2.5 text-sm font-semibold text-primary shadow-[inset_0_1px_0_rgba(144,203,164,0.2)] transition-all hover:border-primary/55 hover:bg-primary/20 active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-70"
                    data-action="check-updates"
                    id="checkUpdatesButton"
                    type="button"
                  >
                    <span class="material-symbols-outlined text-[18px]">sync</span>
                    <span>Check for updates</span>
                  </button>
                  <button
                    class="hidden inline-flex w-full items-center justify-center gap-2 rounded-xl border border-primary/45 bg-gradient-to-r from-primary/85 via-[#9dffbf] to-primary/85 px-4 py-2.5 text-sm font-bold text-[#102316] shadow-[0_10px_24px_rgba(13,242,89,0.25)] transition-all hover:brightness-105 active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-70"
                    data-action="install-update"
                    id="installUpdateButton"
                    type="button"
                  >
                    <span class="material-symbols-outlined text-[18px]">download</span>
                    <span>Download & Install Update</span>
                  </button>
                </div>
              </div>
            </div>
            <button
              class="w-full px-5 py-3 bg-[#22492f] border border-[#366b47] rounded-lg text-sm text-white hover:bg-[#2e573a] transition-all"
              data-action="reset-settings"
            >
              Reset Settings
            </button>
          </div>
        </div>
      </div>
  `;
};
