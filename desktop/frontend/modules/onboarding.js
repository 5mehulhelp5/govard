export const normalizeOnboardingFramework = (framework = "") => {
  const normalized = String(framework || "")
    .trim()
    .toLowerCase();

  if (["", "auto", "detect"].includes(normalized)) {
    return "";
  }
  if (normalized === "m2") {
    return "magento2";
  }
  if (normalized === "m1") {
    return "magento1";
  }
  if (normalized === "wp") {
    return "wordpress";
  }
  return normalized;
};

const inferProjectNameFromPath = (projectPath = "") => {
  const parts = String(projectPath || "")
    .split(/[\\/]+/)
    .map((part) => part.trim())
    .filter(Boolean);
  return parts.at(-1) || "";
};

const normalizeOnboardingDomain = (domain = "", projectPath = "") => {
  const trimmed = String(domain || "")
    .trim()
    .toLowerCase();
  const base = trimmed || inferProjectNameFromPath(projectPath).toLowerCase();
  if (!base) {
    return "";
  }
  if (base.includes(".")) {
    return base;
  }
  return `${base}.test`;
};

export const createOnboardingController = ({
  bridge,
  refs,
  onStatus,
  onToast,
  onProjectAdded,
}) => {
  const browseProject = async () => {
    const path = String((await bridge.pickProjectDirectory()) || "").trim();
    if (!path) {
      return;
    }
    if (refs.projectPath) {
      refs.projectPath.value = path;
    }
    if (refs.displayProjectPath) {
      refs.displayProjectPath.textContent = path;
    }
    if (refs.projectDomain && !String(refs.projectDomain.value || "").trim()) {
      refs.projectDomain.value = inferProjectNameFromPath(path);
    }
    onStatus(`Status: selected project path ${path}`);
  };

  const addProject = async () => {
    const projectPath = String(refs.projectPath?.value || "").trim();
    if (!projectPath) {
      onStatus("Project path is required");
      onToast("Project path is required", "warning");
      return;
    }

    const framework = normalizeOnboardingFramework(refs.projectFramework?.value || "");
    const domain = normalizeOnboardingDomain(refs.projectDomain?.value || "", projectPath);
    const serviceOptions = {
      varnish: Boolean(refs.onboardVarnish?.checked),
      redis: Boolean(refs.onboardRedis?.checked),
      rabbitmq: Boolean(refs.onboardRabbitMQ?.checked),
      elasticsearch: Boolean(refs.onboardElasticsearch?.checked),
    };
    const message = String(
      (await bridge.onboardProject(projectPath, framework, domain, serviceOptions)) ||
        "",
    );
    const isError = message.toLowerCase().includes("failed");
    onStatus(message || "Status: project onboarding finished");
    onToast(
      message || "Project onboarding finished",
      isError ? "error" : "success",
    );
    if (!isError && typeof onProjectAdded === "function") {
      await onProjectAdded();
    }
  };

  const toggleModal = (open) => {
    if (!refs.onboardingModal) {
      return;
    }
    if (open) {
      refs.onboardingModal.classList.remove("hidden");
      refs.onboardingModal.classList.add("flex");
    } else {
      refs.onboardingModal.classList.add("hidden");
      refs.onboardingModal.classList.remove("flex");
    }
  };

  return {
    browseProject,
    addProject,
    toggleModal,
  };
};
