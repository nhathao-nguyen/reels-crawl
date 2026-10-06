// Pure DOM Custom Titlebar Component
import { el } from '../utils/dom.js';
import { API } from '../api.js';

export function createTitlebar() {
  return el('header', { className: 'titlebar' },
    el('div', { className: 'titlebar-brand' },
      el('span', {}, '🎬'),
      el('span', {}, 'Reels Automation Studio'),
      el('span', { className: 'titlebar-badge' }, 'Wails Desktop')
    ),
    el('div', { className: 'titlebar-controls' },
      el('div', {
        className: 'titlebar-btn btn-min',
        title: 'Thu nhỏ',
        onClick: () => API.windowMinimise()
      }),
      el('div', {
        className: 'titlebar-btn btn-max',
        title: 'Phóng to',
        onClick: () => API.windowToggleMaximise()
      }),
      el('div', {
        className: 'titlebar-btn btn-close',
        title: 'Đóng',
        onClick: () => API.windowClose()
      })
    )
  );
}
