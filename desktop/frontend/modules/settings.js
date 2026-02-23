export const normalizeSettingsPayload = (settings = {}) => ({
  theme: settings.theme || settings.Theme || "system",
  proxyTarget: settings.proxyTarget || settings.ProxyTarget || "",
  preferredBrowser:
    settings.preferredBrowser || settings.PreferredBrowser || "",
});

export const applyTheme = (theme) => {
  const root = document.documentElement;
  if (!root) {
    return;
  }
  if (theme === "system") {
    const prefersDark =
      window.matchMedia &&
      window.matchMedia("(prefers-color-scheme: dark)").matches;
    root.dataset.theme = prefersDark ? "dark" : "light";
    return;
  }
  root.dataset.theme = theme;
};

export const createSettingsController = ({
  bridge,
  refs,
  onStatus,
  onToast,
}) => {
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
      applyTheme(settings.theme);
    } catch (_err) {
      applyTheme("system");
    }
  };

  const save = async () => {
    const theme = refs.themeSelect?.value || "system";
    const proxyTarget = refs.proxyTarget?.value || "";
    const preferredBrowser = refs.preferredBrowser?.value || "";
    try {
      const message = await bridge.updateSettings(
        theme,
        proxyTarget,
        preferredBrowser,
      );
      applyTheme(theme);
      onStatus(message);
      onToast(message, "success");
    } catch (err) {
      const message = `Failed to save settings: ${err}`;
      onStatus(message);
      onToast(message, "error");
    }
  };

  const reset = async () => {
    try {
      const message = await bridge.resetSettings();
      onStatus(message);
      onToast(message, "success");
      await load();
    } catch (err) {
      const message = `Failed to reset settings: ${err}`;
      onStatus(message);
      onToast(message, "error");
    }
  };

  return { toggleDrawer, load, save, reset };
};
