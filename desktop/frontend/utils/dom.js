export const byId = (id) => document.getElementById(id)

export const setText = (element, value) => {
  if (!element) {
    return
  }
  element.textContent = String(value ?? "")
}

export const escapeHTML = (value) =>
  String(value ?? "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;")

export const clearChildren = (element) => {
  if (!element) {
    return
  }
  while (element.firstChild) {
    element.removeChild(element.firstChild)
  }
}

