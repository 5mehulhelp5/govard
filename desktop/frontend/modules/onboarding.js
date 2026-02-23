export const normalizeOnboardingRecipe = (recipe = "") => {
  const normalized = String(recipe || "")
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
    onStatus(`Status: selected project path ${path}`);
  };

  const addProject = async () => {
    const projectPath = String(refs.projectPath?.value || "").trim();
    if (!projectPath) {
      onStatus("Project path is required");
      onToast("Project path is required", "warning");
      return;
    }

    const recipe = normalizeOnboardingRecipe(refs.projectRecipe?.value || "");
    const message = String(
      (await bridge.onboardProject(projectPath, recipe)) || "",
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
