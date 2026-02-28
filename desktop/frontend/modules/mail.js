const defaultMailpitURL = "https://mail.govard.test";

const sanitizeProxyTarget = (value = "") =>
  String(value || "")
    .trim()
    .replace(/^https?:\/\//i, "")
    .replace(/\/.*$/, "")
    .replace(/\.+$/g, "")
    .replace(/^\.+/g, "");

export const normalizeMailpitURL = (value = "") => {
  const target = sanitizeProxyTarget(value);
  if (!target) {
    return defaultMailpitURL;
  }
  if (target.startsWith("mail.")) {
    return `https://${target}`;
  }
  return `https://mail.${target}`;
};

export const createMailController = ({ bridge, refs, onStatus, onToast }) => {
  let lastURL = defaultMailpitURL;

  const applyURL = (url) => {
    if (refs.mailFrame && refs.mailFrame.src !== url) {
      refs.mailFrame.src = url;
    }
    if (refs.mailLocation) {
      refs.mailLocation.textContent = url;
    }
  };

  const refresh = async ({ silent = false } = {}) => {
    try {
      const raw = await bridge.getMailpitURL();
      lastURL = normalizeMailpitURL(raw);
    } catch (_err) {
      lastURL = defaultMailpitURL;
    }

    applyURL(lastURL);
    if (!silent) {
      onStatus(`Status: Mailpit inbox ready at ${lastURL}`);
    }
    return lastURL;
  };

  const openExternal = async () => {
    const url = await refresh({ silent: true });
    try {
      const message = await bridge.quickActionForProject(
        "open-mail-client",
        "",
      );
      onStatus(message);
      onToast(message, "success");
      return;
    } catch (_err) {
      if (typeof window !== "undefined" && typeof window.open === "function") {
        window.open(url, "_blank", "noopener");
        onStatus(`Opened Mailpit: ${url}`);
      }
    }
  };

  return {
    refresh,
    openExternal,
  };
};
