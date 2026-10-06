// Pure DOM Sidebar Navigation Component
import { el } from '../utils/dom.js';

export function createSidebar(onNavigate, onOpenModal) {
  // Navigation elements
  const navDashboard = el('div', {
    id: 'nav-dashboard',
    className: 'nav-item active',
    onClick: () => onNavigate('dashboard')
  },
    el('span', {}, '📥'),
    el('span', {}, 'Tiến Trình (Tasks)'),
    el('span', { id: 'sidebarTaskCount', className: 'nav-badge' }, '0')
  );

  const navSources = el('div', {
    id: 'nav-sources',
    className: 'nav-item',
    onClick: () => onNavigate('sources')
  },
    el('span', {}, '🌐'),
    el('span', {}, 'Kênh Theo Dõi'),
    el('span', { id: 'sidebarSourceCount', className: 'nav-badge' }, '0')
  );

  const navSettings = el('div', {
    id: 'nav-settings',
    className: 'nav-item',
    onClick: () => onNavigate('settings')
  },
    el('span', {}, '⚙️'),
    el('span', {}, 'Cài Đặt Hệ Thống')
  );

  // Quick Action items
  const navToolModel = el('div', {
    className: 'nav-item',
    onClick: () => onOpenModal('modalWhisperDownloader')
  },
    el('span', {}, '🧠'),
    el('span', {}, 'Tải Model Whisper')
  );

  const navToolExport = el('div', {
    className: 'nav-item',
    style: { color: 'var(--success)' },
    onClick: () => onOpenModal('modalExportConfirm')
  },
    el('span', {}, '⚡'),
    el('span', {}, 'Kích Hoạt Auto Export')
  );

  // Sidebar container
  const sidebar = el('aside', { className: 'sidebar' },
    el('div', { className: 'sidebar-logo' },
      el('div', { className: 'logo-badge' }, '⚡'),
      el('div', {},
        el('div', { className: 'logo-title' }, 'Reels Studio'),
        el('div', { className: 'logo-subtitle' }, 'CapCut Pipeline')
      )
    ),

    el('div', { className: 'nav-section-title' }, 'Menu Chính'),
    navDashboard,
    navSources,
    navSettings,

    el('div', { className: 'nav-section-title', style: { marginTop: '10px' } }, 'Công Cụ Nhanh'),
    navToolModel,
    navToolExport,

    // Footer
    el('div', { className: 'sidebar-footer' },
      el('div', { className: 'system-status-indicator' },
        el('div', { className: 'status-dot' }),
        el('span', {}, 'Hệ thống sẵn sàng')
      ),
      el('div', { id: 'sidebarEngine', style: { color: 'var(--text-muted)' } }, 'Engine: Groq API Turbo'),
      el('div', { id: 'sidebarDiskFree', style: { color: 'var(--text-muted)' } }, 'Ổ đĩa trống: -- GB')
    )
  );

  return sidebar;
}
