// Pure DOM Tracked Sources Page Component (Zero HTML strings)
import { el, clear } from '../utils/dom.js';
import { API } from '../api.js';
import { Toast } from '../components/toast.js';

export function createSourcesPage() {
  // Header
  const header = el('div', { className: 'page-header' },
    el('div', {},
      el('h1', { className: 'page-title' }, 'Quản Lý Kênh & Nguồn Theo Dõi'),
      el('p', { className: 'page-subtitle' }, 'Hệ thống lưu danh sách các Page Facebook và Profile TikTok để tự động quét video mới.')
    ),
    el('button', {
      className: 'btn btn-primary',
      onClick: async () => {
        try {
          Toast.info('Đang quét kiểm tra video mới trên tất cả kênh theo dõi...');
          await API.scanTrackedSources();
          Toast.success('Đã kích hoạt quét kênh thành công!');
        } catch (err) {
          Toast.error(String(err));
        }
      }
    }, '⟳ Quét Tất Cả Kênh')
  );

  // Add source form
  const sourceUrlInput = el('input', {
    type: 'text',
    className: 'form-input source-url-input',
    placeholder: 'Dán link Page hoặc Profile (Ví dụ: https://www.facebook.com/ghienphimhay)...',
    onKeydown: (e) => { if (e.key === 'Enter') handleAddSource(); }
  });

  const sourceNameInput = el('input', {
    type: 'text',
    className: 'form-input source-name-input',
    placeholder: 'Tên tác giả / kênh (tùy chọn)'
  });

  const addCard = el('div', { className: 'card' },
    el('div', { className: 'source-add-row' },
      sourceUrlInput,
      sourceNameInput,
      el('button', { className: 'btn btn-primary', onClick: handleAddSource }, '+ Thêm Kênh')
    )
  );

  // Sources table
  const tableBody = el('tbody', {},
    el('tr', {},
      el('td', { colSpan: 6, style: { textAlign: 'center', color: 'var(--text-dim)', padding: '24px' } }, 'Đang tải danh sách kênh...')
    )
  );

  const tableCard = el('div', { className: 'card', style: { padding: '16px 18px' } },
    el('div', { className: 'table-wrapper' },
      el('table', {},
        el('thead', {},
          el('tr', {},
            el('th', { style: { width: '140px' } }, 'Nền Tảng'),
            el('th', {}, 'Tên Kênh / Profile'),
            el('th', {}, 'URL Theo Dõi'),
            el('th', { style: { width: '130px' } }, 'Số Video Đã Cào'),
            el('th', { style: { width: '150px' } }, 'Lần Quét Cuối'),
            el('th', { style: { width: '80px', textAlign: 'right' } }, 'Thao Tác')
          )
        ),
        tableBody
      )
    )
  );

  async function handleAddSource() {
    const url = sourceUrlInput.value.trim();
    const name = sourceNameInput.value.trim();

    if (!url) {
      Toast.error('Vui lòng nhập đường dẫn kênh hoặc Fanpage');
      return;
    }

    try {
      await API.addTrackedSource(url, name);
      Toast.success('Đã thêm kênh theo dõi');
      sourceUrlInput.value = '';
      sourceNameInput.value = '';
      loadSources();
    } catch (err) {
      Toast.error(String(err));
    }
  }

  async function loadSources() {
    try {
      const sources = await API.listTrackedSources();
      clear(tableBody);

      // Update sidebar counter
      const sidebarBadge = document.getElementById('sidebarSourceCount');
      if (sidebarBadge) sidebarBadge.innerText = sources ? sources.length : 0;

      if (!sources || sources.length === 0) {
        tableBody.appendChild(
          el('tr', {},
            el('td', { colSpan: 6, style: { textAlign: 'center', color: 'var(--text-dim)', padding: '30px' } },
              'Chưa theo dõi kênh nào. Thêm link Page/Profile bên trên!'
            )
          )
        );
        return;
      }

      sources.forEach(s => {
        const isFb = s.platform === 'facebook';
        const platformLabel = isFb ? 'ⓕ Facebook' : '♬ TikTok';
        const platformClass = isFb ? 'platform-fb' : 'platform-tt';
        const lastCheck = s.last_checked_at ? new Date(s.last_checked_at).toLocaleString('vi-VN') : 'Chưa quét';

        const row = el('tr', {},
          el('td', {}, el('span', { className: `platform-indicator ${platformClass}` }, platformLabel)),
          el('td', {}, el('strong', { style: { color: 'var(--text-main)' } }, s.author_name || 'Kênh Chưa Đặt Tên')),
          el('td', {}, el('span', { className: 'source-link-cell' }, s.source_url)),
          el('td', {}, el('strong', { style: { color: 'var(--text-main)' } }, `${s.video_count || 0}`), ' video'),
          el('td', { style: { color: 'var(--text-muted)', fontSize: '12px' } }, lastCheck),
          el('td', { style: { textAlign: 'right' } },
            el('button', {
              className: 'btn-icon',
              title: 'Xóa kênh',
              onClick: async () => {
                if (!confirm(`Bạn có chắc muốn xóa kênh ${s.author_name || s.source_url}?`)) return;
                try {
                  await API.deleteTrackedSource(s.id);
                  Toast.success('Đã xóa kênh theo dõi');
                  loadSources();
                } catch (err) {
                  Toast.error(String(err));
                }
              }
            }, '✕')
          )
        );

        tableBody.appendChild(row);
      });
    } catch (err) {
      console.error('Lỗi đọc kênh:', err);
    }
  }

  // Initial load
  loadSources();

  const container = el('div', { style: { display: 'flex', flexDirection: 'column', gap: '18px' } },
    header,
    addCard,
    tableCard
  );

  container.refresh = loadSources;
  return container;
}
