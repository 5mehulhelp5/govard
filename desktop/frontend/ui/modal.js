import { byId, setHTML, setText } from "../utils/dom.js";

const modal = byId("confirmModal");
const titleEl = byId("confirmTitle");
const messageEl = byId("confirmMessage");
const iconEl = byId("confirmIcon");
const cancelBtn = byId("confirmCancelBtn");
const confirmBtn = byId("confirmConfirmBtn");

let onConfirm = null;
let onCancel = null;

const hide = () => {
  if (!modal) return;
  modal.classList.add("opacity-0", "pointer-events-none");
  modal.classList.remove("opacity-100", "pointer-events-auto");
  const inner = modal.querySelector("div");
  if (inner) {
    inner.classList.add("scale-95");
    inner.classList.remove("scale-100");
  }
};

const show = () => {
  if (!modal) return;
  modal.classList.remove("opacity-0", "pointer-events-none");
  modal.classList.add("opacity-100", "pointer-events-auto");
  const inner = modal.querySelector("div");
  if (inner) {
    inner.classList.remove("scale-95");
    inner.classList.add("scale-100");
  }
};

if (cancelBtn) {
  cancelBtn.addEventListener("click", () => {
    hide();
    if (onCancel) onCancel();
  });
}

if (confirmBtn) {
  confirmBtn.addEventListener("click", () => {
    hide();
    if (onConfirm) onConfirm();
  });
}

/**
 * Show a themed confirmation dialog.
 * 
 * @param {Object} options
 * @param {string} options.title - Modal title
 * @param {string} options.message - Modal message
 * @param {string} options.icon - Material icon name
 * @param {string} options.confirmLabel - Label for confirm button
 * @param {string} options.cancelLabel - Label for cancel button
 * @returns {Promise<boolean>} - Resolves to true if confirmed, false if cancelled
 */
export const confirm = ({ title, message, icon, confirmLabel, cancelLabel }) => {
  return new Promise((resolve) => {
    setText(titleEl, title || "Confirm Action");
    setHTML(messageEl, message || "Are you sure you want to proceed?");
    setText(iconEl, icon || "help");
    setText(confirmBtn, confirmLabel || "Confirm");
    setText(cancelBtn, cancelLabel || "Cancel");
    
    onConfirm = () => resolve(true);
    onCancel = () => resolve(false);
    
    show();
  });
};
