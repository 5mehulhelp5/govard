export const normalizeSettingsPayload = (settings = {}) => ({
  theme: settings.theme || settings.Theme || "system",
  proxyTarget: settings.proxyTarget || settings.ProxyTarget || "",
  preferredBrowser:
    settings.preferredBrowser || settings.PreferredBrowser || "",
  codeEditor: settings.codeEditor || settings.CodeEditor || "",
  dbClientPreference:
    settings.dbClientPreference || settings.DBClientPreference || "desktop",
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
    } catch (_err) {
      applyTheme("system");
    }
  };

  const save = async () => {
    const theme = refs.themeSelect?.value || "system";
    const proxyTarget = refs.proxyTarget?.value || "";
    const preferredBrowser = refs.preferredBrowser?.value || "";
    const codeEditor = refs.codeEditor?.value || "";
    const dbClientPreference = refs.dbClientPreference?.value || "desktop";
    try {
      const message = await bridge.updateSettings(
        theme,
        proxyTarget,
        preferredBrowser,
        codeEditor,
        dbClientPreference,
      );
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

  return { toggleDrawer, load, save, reset };
};
