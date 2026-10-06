// Pure DOM Application Bootstrap & Controller
// Assembles the entire UI dynamically via JavaScript without raw HTML templates.

import { el, clear } from './utils/dom.js';
import { API } from './api.js';
import { Toast } from './components/toast.js';
import { Modal } from './components/modal.js';
import { createTitlebar } from './components/titlebar.js';
import { createSidebar } from './components/sidebar.js';
import { createDashboardPage } from './pages/dashboard.js';
import { createSourcesPage } from './pages/sources.js';
import { createSettingsPage } from './pages/settings.js';

let currentPage = 'dashboard';
let activePageNode = null;
let viewportNode = null;

// Modal opener dispatcher
function openModalById(modalId) {
  if (modalId === 'modalWhisperDownloader') {
    Modal.open(Modal.createWhisperModal());
  } else if (modalId === 'modalExportConfirm') {
    Modal.open(Modal.createExportModal(async () => {
      try {
        await API.startPhase2();
        Toast.success('Đã kích hoạt chế độ tự động xuất video hàng loạt (Pha 2)');
      } catch (err) {
        Toast.error(String(err));
      }
    }));
  }
}

// Router & View Switcher
export function navigateTo(pageName) {
  currentPage = pageName;
  if (!viewportNode) return;

  // Update sidebar active classes
  document.querySelectorAll('.nav-item').forEach(el => el.classList.remove('active'));
  const activeNav = document.getElementById(`nav-${pageName}`);
  if (activeNav) activeNav.classList.add('active');

  // Clear current viewport & mount new pure DOM page
  clear(viewportNode);

  if (pageName === 'dashboard') {
    activePageNode = createDashboardPage(() => openModalById('modalExportConfirm'));
  } else if (pageName === 'sources') {
    activePageNode = createSourcesPage();
  } else if (pageName === 'settings') {
    activePageNode = createSettingsPage();
  }

  if (activePageNode) {
    viewportNode.appendChild(activePageNode);
  }
}

// Periodic System Status Updater
async function updateSystemStatus() {
  try {
    const status = await API.getSystemStatus();
    if (!status) return;

    const diskEl = document.getElementById('sidebarDiskFree');
    if (diskEl) diskEl.innerText = `Ổ đĩa trống: ${status.free_disk_gb} GB`;

    const engineEl = document.getElementById('sidebarEngine');
    if (engineEl) {
      engineEl.innerText = `Engine: ${status.whisper_provider === 'local_offline' ? 'Whisper Local' : 'Groq API Turbo'}`;
    }
  } catch (err) {
    console.error('Lỗi đọc system status:', err);
  }
}

// Application Entry Point
window.addEventListener('DOMContentLoaded', () => {
  const root = document.getElementById('app');
  if (!root) return;

  clear(root);

  // 1. Build Titlebar
  const titlebar = createTitlebar();

  // 2. Build Viewport
  viewportNode = el('main', { className: 'app-viewport' });

  // 3. Build Sidebar
  const sidebar = createSidebar(navigateTo, openModalById);

  // 4. Build Layout Structure
  const appLayout = el('div', { className: 'app-layout' },
    sidebar,
    viewportNode
  );

  // 5. Mount everything to root #app
  root.appendChild(titlebar);
  root.appendChild(appLayout);

  // 6. Navigate to initial page & fetch status
  navigateTo('dashboard');
  updateSystemStatus();

  // 7. Background poller (refreshes active view every 3s)
  setInterval(() => {
    if (activePageNode && typeof activePageNode.refresh === 'function') {
      activePageNode.refresh();
    }
  }, 3000);
});
