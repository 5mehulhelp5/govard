export const createToast = (container) => {
  const show = (message, type = "success") => {
    if (!container) {
      return
    }
    const item = document.createElement("div")
    item.className = `toast toast--${type}`
    item.textContent = String(message || "")
    container.appendChild(item)
    requestAnimationFrame(() => item.classList.add("is-visible"))
    setTimeout(() => {
      item.classList.remove("is-visible")
      setTimeout(() => item.remove(), 180)
    }, 2200)
  }

  return { show }
}

