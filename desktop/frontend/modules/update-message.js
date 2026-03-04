const isRedundantVersionTransitionMessage = (
  message,
  currentVersion,
  latestVersion,
) => {
  const normalized = String(message || "").trim();
  if (!normalized) {
    return false;
  }

  const lower = normalized.toLowerCase();
  const fromVersion = String(currentVersion || "").trim();
  const toVersion = String(latestVersion || "").trim();

  const hasArrow = normalized.includes("->") || normalized.includes("→");
  const hasCurrentLatestWords = lower.includes("current") && lower.includes("latest");
  const hasUpdateHint = lower.includes("update available") || lower.includes("new version");
  const hasFromVersion = fromVersion ? normalized.includes(fromVersion) : false;
  const hasToVersion = toVersion ? normalized.includes(toVersion) : false;

  if (hasCurrentLatestWords && (hasFromVersion || hasToVersion || hasUpdateHint)) {
    return true;
  }

  if (hasArrow && (hasUpdateHint || hasFromVersion || hasToVersion)) {
    return true;
  }

  return false;
};

const messageAlreadyContainsVersionTransition = (message, currentVersion, latestVersion) => {
  const normalized = String(message || "").trim();
  if (!normalized) {
    return false;
  }

  const fromVersion = String(currentVersion || "").trim();
  const toVersion = String(latestVersion || "").trim();

  const hasArrow = normalized.includes("->") || normalized.includes("→");
  const hasFromVersion = fromVersion ? normalized.includes(fromVersion) : false;
  const hasToVersion = toVersion ? normalized.includes(toVersion) : false;
  return hasArrow || (hasFromVersion && hasToVersion);
};

const appendVersionTransition = (message, currentVersion, latestVersion) => {
  const fromVersion = String(currentVersion || "").trim();
  const toVersion = String(latestVersion || "").trim();
  if (!fromVersion || !toVersion) {
    return message;
  }

  const normalized = String(message || "").trim();
  if (!normalized) {
    return `${fromVersion} -> ${toVersion}`;
  }

  const trimmed = normalized.replace(/[.\s]+$/, "");
  return `${trimmed} (${fromVersion} -> ${toVersion}).`;
};

export const formatUpdateMessage = (
  { message, currentVersion, latestVersion },
  options = {},
) => {
  const includeVersionTransition =
    options && typeof options === "object"
      ? Boolean(options.includeVersionTransition)
      : false;
  const normalizedMessage = String(message || "").trim();
  if (
    normalizedMessage &&
    !isRedundantVersionTransitionMessage(
      normalizedMessage,
      currentVersion,
      latestVersion,
    )
  ) {
    if (
      includeVersionTransition &&
      !messageAlreadyContainsVersionTransition(
        normalizedMessage,
        currentVersion,
        latestVersion,
      )
    ) {
      return appendVersionTransition(
        normalizedMessage,
        currentVersion,
        latestVersion,
      );
    }
    return normalizedMessage;
  }

  const fromVersion = String(currentVersion || "").trim();
  const toVersion = String(latestVersion || "").trim();
  if (fromVersion && toVersion) {
    const messageWithVersion = "A new Govard Desktop version is ready to install.";
    if (includeVersionTransition) {
      return appendVersionTransition(messageWithVersion, fromVersion, toVersion);
    }
    return messageWithVersion;
  }
  if (toVersion) {
    return `Version ${toVersion} is ready to install.`;
  }

  return "A new Govard Desktop version is available.";
};
