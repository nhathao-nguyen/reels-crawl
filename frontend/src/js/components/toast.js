// Pure DOM Toast Component
import { el } from '../utils/dom.js';

let toastContainer = null;

function getContainer() {
  if (!toastContainer) {
    toastContainer = document.getElementById('toast-container');
    if (!toastContainer) {
      toastContainer = el('div', { id: 'toast-container', className: 'toast-container' });
      document.body.appendChild(toastContainer);
    }
  }
  return toastContainer;
}

export const Toast = {
  show: (message, type = 'info') => {
    const container = getContainer();

    let icon = 'ℹ️';
    if (type === 'success') icon = '✓';
    if (type === 'error') icon = '✕';

    const toast = el('div', { className: `toast toast-${type}` },
      el('span', { style: { fontWeight: '700' } }, icon),
      el('span', {}, String(message))
    );

    container.appendChild(toast);

    setTimeout(() => {
      toast.style.opacity = '0';
      toast.style.transform = 'translateY(10px)';
      toast.style.transition = 'all 0.2s ease';
      setTimeout(() => toast.remove(), 250);
    }, 3500);
  },

  success: (msg) => Toast.show(msg, 'success'),
  error: (msg) => Toast.show(msg, 'error'),
  info: (msg) => Toast.show(msg, 'info')
};
