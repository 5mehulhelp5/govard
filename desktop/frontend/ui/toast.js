export const createToast = (container) => {
  const activeMessages = new Set();

  const getIcon = (type) => {
    switch (type) {
      case "error":
        return "report";
      case "warning":
        return "warning";
      case "info":
        return "info";
      default:
        return "check_circle";
    }
  };

  const show = (message, type = "success") => {
    if (!container || !message) return;

    const msg = String(message).trim();
    if (activeMessages.has(msg)) return;
    activeMessages.add(msg);

    const item = document.createElement("div");
    item.className = `toast toast--${type} group`;

    const duration = type === "error" || type === "warning" ? 6000 : 4000;

    item.innerHTML = `
      <div class="toast-indicator"></div>
      <div class="toast-icon-wrapper">
        <span class="material-symbols-outlined toast-icon">${getIcon(type)}</span>
      </div>
      <div class="toast-content">
        <p class="toast-message">${msg}</p>
      </div>
      <button class="toast-close" aria-label="Close">
        <span class="material-symbols-outlined">close</span>
      </button>
      <div class="toast-progress" style="animation: toast-shrink ${duration}ms linear forwards"></div>
    `;

    container.appendChild(item);

    const close = () => {
      if (item.classList.contains("is-removing")) return;
      item.classList.add("is-removing");
      item.classList.remove("is-visible");
      setTimeout(() => {
        item.remove();
        activeMessages.delete(msg);
      }, 500);
    };

    item.querySelector(".toast-close").addEventListener("click", (e) => {
      e.stopPropagation();
      close();
    });

    requestAnimationFrame(() => item.classList.add("is-visible"));
    setTimeout(close, duration);
  };

  return { show };
};
