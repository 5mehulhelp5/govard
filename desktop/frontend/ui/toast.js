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

    // Simple deduplication to prevent flooding
    if (activeMessages.has(message)) return;
    activeMessages.add(message);

    const item = document.createElement("div");
    item.className = `toast toast--${type} group`;

    // Add progress bar for timing feedback
    const duration = type === "error" || type === "warning" ? 6000 : 4000;

    item.innerHTML = `
      <div class="toast-indicator"></div>
      <span class="material-symbols-outlined text-[20px] shrink-0 toast-icon">${getIcon(type)}</span>
      <div class="flex-1 min-w-0 py-0.5">
        <p class="text-sm font-medium leading-normal text-white/95">${String(message)}</p>
      </div>
      <button class="opacity-0 group-hover:opacity-100 p-1.5 -mr-1.5 rounded-full hover:bg-white/10 text-white/40 hover:text-white transition-all transform hover:scale-105" aria-label="Close">
        <span class="material-symbols-outlined text-[18px]">close</span>
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
        activeMessages.delete(message);
      }, 500);
    };

    item.querySelector("button").addEventListener("click", (e) => {
      e.stopPropagation();
      close();
    });

    requestAnimationFrame(() => item.classList.add("is-visible"));
    setTimeout(close, duration);
  };

  return { show };
};
