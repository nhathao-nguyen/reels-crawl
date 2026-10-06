// Pure DOM Dashboard Page Component (Zero HTML strings)
import { el, clear } from '../utils/dom.js';
import { API } from '../api.js';
import { Toast } from '../components/toast.js';

let activeFilter = 'all';

export function createDashboardPage(onOpenExportModal) {
  // --- Header ---
  const header = el('div', { className: 'page-header' },
    el('div', {},
      el('h1', { className: 'page-title' }, '📥 Bảng Điều Khiển & Tiến Trình Xử Lý'),
      el('p', { className: 'page-subtitle' }, 'Hệ thống tự động hóa Pipeline: Tải video Reels/TikTok sạch -> Bóc tách web -> Speech-to-Text -> CapCut Draft -> Xuất 1080p.')
    ),
    el('div', { style: { display: 'flex', gap: '8px' } },
      el('button', {
        className: 'btn btn-secondary',
        onClick: () => {
          API.openDownloadsFolder();
          Toast.info('Đang mở thư mục downloads');
        }
      }, '📁 Mở Thư Mục Downloads'),
      el('button', {
        className: 'btn btn-success',
        style: { fontWeight: '700' },
        onClick: () => onOpenExportModal()
      }, '⚡ Xuất Video Hàng Loạt (Pha 2)')
    )
  );

  // --- Input Card ---
  const taskUrlInput = el('textarea', {
    className: 'form-input',
    rows: 2,
    style: { resize: 'vertical', fontFamily: 'var(--font-mono)', fontSize: '13px', lineHeight: '1.4' },
    placeholder: 'Nhập một đường link hoặc dán danh sách nhiều link (mỗi dòng 1 link):\nVí dụ: https://www.facebook.com/reel/991736342092712'
  });

  const btnSelectFile = el('button', {
    className: 'btn btn-secondary',
    style: { fontSize: '12px', padding: '6px 12px' },
    onClick: async () => {
      try {
        const urls = await API.selectLinksFile();
        if (urls && urls.length > 0) {
          taskUrlInput.value = urls.join('\n');
          Toast.success(`Đã nạp ${urls.length} link từ file!`);
        }
      } catch (err) {
        Toast.error('Không thể đọc file: ' + err);
      }
    }
  }, '📄 Chọn file links.txt');

  const chkDeduplicate = el('input', { type: 'checkbox', checked: true });
  const chkExtractComments = el('input', { type: 'checkbox', checked: true });
  const selectTemplate = el('select', { className: 'form-select', style: { padding: '5px 10px', fontSize: '12.5px' } },
    el('option', { value: 'template' }, 'Đang tìm template...')
  );

  async function loadTemplates() {
    try {
      const templates = await API.listAvailableTemplates();
      const cfg = await API.getConfig();
      const defaultTpl = cfg?.template_name || 'template';

      clear(selectTemplate);
      if (templates && templates.length > 0) {
        templates.forEach(t => {
          const isSelected = (t === defaultTpl);
          selectTemplate.appendChild(el('option', { value: t, selected: isSelected }, `${t}${isSelected ? ' (Mặc định)' : ''}`));
        });
      } else {
        selectTemplate.appendChild(el('option', { value: 'template' }, 'template (Mặc định)'));
      }
    } catch (err) {
      console.error('Lỗi tải danh sách template:', err);
    }
  }
  loadTemplates();

  const btnStartPhase1 = el('button', {
    className: 'btn btn-primary',
    style: { padding: '10px 20px', fontWeight: '700', fontSize: '14px' },
    onClick: handleStartPhase1
  }, '▶ BẮT ĐẦU CÀO (PHA 1)');

  const btnStartPhase2 = el('button', {
    className: 'btn btn-success',
    style: { padding: '10px 20px', fontWeight: '700', fontSize: '14px' },
    onClick: () => onOpenExportModal()
  }, '⚡ XUẤT VIDEO HÀNG LOẠT (PHA 2)');

  const inputCard = el('div', { className: 'card' },
    el('div', { style: { marginBottom: '8px', fontSize: '13px', fontWeight: '600', color: 'var(--text-main)' } },
      '[+] Nhập liên kết (Profile Facebook / TikTok URL hoặc dán nhiều link):'
    ),
    taskUrlInput,
    el('div', { className: 'task-options-row', style: { marginTop: '10px' } },
      btnSelectFile,
      el('label', { className: 'form-checkbox-label' }, chkDeduplicate, ' Bỏ qua video trùng lặp'),
      el('label', { className: 'form-checkbox-label' }, chkExtractComments, ' Bóc tách link bài viết'),
      el('div', { className: 'template-picker' },
        el('span', { style: { fontSize: '12.5px', color: 'var(--text-muted)' } }, 'Template:'),
        selectTemplate
      )
    ),
    el('div', { style: { display: 'flex', gap: '12px', marginTop: '14px' } },
      btnStartPhase1,
      btnStartPhase2
    )
  );

  // --- Progress & Metrics Card ---
  const progressBar = el('div', { className: 'progress-fill', style: { width: '0%', transition: 'width 0.3s ease' } });
  const progressPercent = el('strong', { style: { color: 'var(--primary)' } }, '0%');
  const progressSummary = el('span', { style: { fontSize: '12px', color: 'var(--text-muted)' } }, 'Sẵn sàng');
  const liveStatusText = el('div', {
    style: { fontSize: '12px', color: 'var(--text-secondary)', marginTop: '6px', fontFamily: 'var(--font-mono)' }
  }, 'Hệ thống đang chờ lệnh...');

  const statTotal = el('span', { className: 'stat-number' }, '0');
  const statPending = el('span', { className: 'stat-number', style: { color: 'var(--warning)' } }, '0');
  const statDownloaded = el('span', { className: 'stat-number', style: { color: 'var(--cyan)' } }, '0');
  const statDraft = el('span', { className: 'stat-number', style: { color: 'var(--purple)' } }, '0');
  const statExported = el('span', { className: 'stat-number', style: { color: 'var(--success)' } }, '0');

  const progressCard = el('div', { className: 'card progress-card' },
    el('div', { className: 'progress-header-row' },
      el('span', { style: { color: 'var(--text-main)', fontWeight: '600' } }, 'TIẾN TRÌNH TỔNG QUAN: ', progressPercent),
      progressSummary
    ),
    el('div', { className: 'progress-track' }, progressBar),
    liveStatusText,
    el('div', { className: 'stats-grid', style: { marginTop: '14px' } },
      el('div', { className: 'stat-item' }, el('span', { className: 'stat-label' }, 'Tổng tác vụ'), statTotal),
      el('div', { className: 'stat-item' }, el('span', { className: 'stat-label' }, 'Chờ xử lý'), statPending),
      el('div', { className: 'stat-item' }, el('span', { className: 'stat-label' }, 'Đã tải gốc'), statDownloaded),
      el('div', { className: 'stat-item' }, el('span', { className: 'stat-label' }, 'Đã tạo Draft'), statDraft),
      el('div', { className: 'stat-item' }, el('span', { className: 'stat-label' }, 'Đã xuất MP4'), statExported)
    )
  );

  // --- Tasks Queue Table Card ---
  const tableBody = el('tbody', {},
    el('tr', {},
      el('td', { colSpan: 6, style: { textAlign: 'center', color: 'var(--text-dim)', padding: '24px' } }, 'Đang tải danh sách tác vụ...')
    )
  );

  const filterButtons = ['all', 'pending', 'downloaded', 'draft_created', 'exported', 'failed'].map(f => {
    const labels = {
      all: 'Tất cả',
      pending: 'Chờ xử lý',
      downloaded: 'Đã tải',
      draft_created: 'Đã tạo Draft',
      exported: 'Đã xuất video',
      failed: 'Lỗi'
    };
    const btn = el('button', {
      className: `filter-pill ${f === activeFilter ? 'active' : ''}`,
      onClick: () => {
        filterButtons.forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        activeFilter = f;
        loadTasks();
      }
    }, labels[f]);
    return btn;
  });

  const tableCard = el('div', { className: 'card', style: { padding: '16px 18px' } },
    el('div', { className: 'table-filter-bar' },
      el('div', { className: 'filter-pills' }, ...filterButtons),
      el('div', { style: { display: 'flex', gap: '8px' } },
        el('button', {
          className: 'btn btn-secondary',
          style: { padding: '5px 10px', fontSize: '12px' },
          onClick: () => {
            loadData();
            Toast.info('Đã làm mới dữ liệu');
          }
        }, '⟳ Làm mới'),
        el('button', {
          className: 'btn btn-secondary',
          style: { padding: '5px 10px', fontSize: '12px' },
          onClick: async () => {
            try {
              await API.clearCompletedTasks();
              Toast.success('Đã dọn dẹp các video xuất thành công');
              loadData();
            } catch (err) {
              Toast.error(String(err));
            }
          }
        }, '🗑 Dọn task đã xuất')
      )
    ),
    el('div', { className: 'table-wrapper' },
      el('table', {},
        el('thead', {},
          el('tr', {},
            el('th', { style: { width: '130px' } }, 'Nền Tảng'),
            el('th', {}, 'ID Video / Tiêu Đề'),
            el('th', {}, 'Bóc Tách Web'),
            el('th', {}, 'Phụ Đề (STT)'),
            el('th', {}, 'Trạng Thái'),
            el('th', { style: { width: '130px', textAlign: 'right' } }, 'Thao Tác')
          )
        ),
        tableBody
      )
    )
  );

  // --- Handlers & Data Loading ---
  async function handleStartPhase1() {
    const text = taskUrlInput.value.trim();
    const urls = text ? text.split('\n').map(s => s.trim()).filter(s => s.length > 0 && !s.startsWith('#')) : [];

    if (urls.length === 0) {
      Toast.error('Vui lòng nhập ít nhất 1 liên kết Reel hoặc chọn file links.txt');
      return;
    }

    try {
      btnStartPhase1.disabled = true;
      btnStartPhase1.innerText = '⏳ Đang Xử Lý Pha 1...';
      liveStatusText.innerText = `Đang khởi động cào dữ liệu cho ${urls.length} video...`;
      Toast.info(`Bắt đầu xử lý ${urls.length} link...`);

      await API.startPhase1(urls, selectTemplate.value, chkDeduplicate.checked, chkExtractComments.checked);
      taskUrlInput.value = '';
      loadData();
    } catch (err) {
      Toast.error('Lỗi khởi động: ' + err);
      btnStartPhase1.disabled = false;
      btnStartPhase1.innerText = '▶ BẮT ĐẦU CÀO (PHA 1)';
    }
  }

  async function loadStats() {
    try {
      const stats = await API.getPipelineStats();
      statTotal.innerText = stats.total || 0;
      statPending.innerText = stats.pending || 0;
      statDownloaded.innerText = stats.downloaded || 0;
      statDraft.innerText = stats.draft_created || 0;
      statExported.innerText = stats.exported || 0;

      const sidebarBadge = document.getElementById('sidebarTaskCount');
      if (sidebarBadge) sidebarBadge.innerText = stats.total || 0;

      const total = stats.total || 0;
      const completed = (stats.draft_created || 0) + (stats.exported || 0);
      const pct = total > 0 ? Math.round((completed / total) * 100) : 0;

      if (!progressBar.classList.contains('active-running')) {
        progressBar.style.width = `${pct}%`;
        progressPercent.innerText = `${pct}%`;
        progressSummary.innerText = `${stats.exported || 0} / ${total} hoàn tất`;
      }
    } catch (err) {
      console.error('Lỗi đọc stats:', err);
    }
  }

  async function loadTasks() {
    try {
      const videos = await API.listVideos(activeFilter);
      clear(tableBody);

      if (!videos || videos.length === 0) {
        tableBody.appendChild(
          el('tr', {},
            el('td', { colSpan: 6, style: { textAlign: 'center', color: 'var(--text-dim)', padding: '30px' } },
              'Chưa có tác vụ nào trong danh sách. Hãy nhập liên kết bên trên để bắt đầu!'
            )
          )
        );
        return;
      }

      videos.forEach(v => {
        const isFb = v.platform === 'facebook';
        const platformLabel = isFb ? 'ⓕ Facebook' : '♬ TikTok';
        const platformClass = isFb ? 'platform-fb' : 'platform-tt';

        let badge = el('span', { className: 'badge badge-pending' }, '⟳ Chờ xử lý');
        if (v.status === 'downloading') badge = el('span', { className: 'badge badge-downloaded' }, '⬇ Đang tải...');
        if (v.status === 'downloaded') badge = el('span', { className: 'badge badge-downloaded' }, '● Đã tải gốc');
        if (v.status === 'draft_created') badge = el('span', { className: 'badge badge-draft' }, '● Đã tạo Draft');
        if (v.status === 'exported') badge = el('span', { className: 'badge badge-exported' }, '✓ Đã xuất video');
        if (v.status === 'failed') badge = el('span', { className: 'badge badge-failed', title: v.error_message || '' }, '✕ Thất bại');

        const actions = [
          el('button', {
            className: 'btn-icon',
            title: 'Mở thư mục video',
            onClick: () => {
              API.openTaskFolder(v.platform, v.video_id);
              Toast.info(`Mở thư mục: ${v.platform}_${v.video_id}`);
            }
          }, '📁')
        ];

        // Button to open exported final video
        if (v.export_path || v.status === 'exported') {
          actions.push(
            el('button', {
              className: 'btn-icon',
              title: 'Xem video thành phẩm (1080p)',
              onClick: () => {
                API.openExportedVideo(v.export_path);
                Toast.info('Đang mở video thành phẩm');
              }
            }, '🎬')
          );
        }

        // Button to open CapCut project
        if (v.status === 'draft_created' || v.status === 'exported') {
          actions.push(
            el('button', {
              className: 'btn-icon',
              title: 'Mở dự án CapCut',
              onClick: () => {
                const projectName = `Dự án ${v.platform}_${v.video_id}`;
                API.openCapCutProject(projectName);
                Toast.info(`Đang mở CapCut: ${projectName}`);
              }
            }, '✂')
          );
        }

        // Error info button
        if (v.status === 'failed') {
          actions.push(
            el('button', {
              className: 'btn-icon',
              style: { color: 'var(--danger)' },
              title: v.error_message || 'Xem lỗi',
              onClick: () => alert(`Chi tiết lỗi của tác vụ ${v.video_id}:\n\n${v.error_message}`)
            }, '⚠️')
          );
        }

        // Delete button
        actions.push(
          el('button', {
            className: 'btn-icon',
            title: 'Xóa tác vụ',
            onClick: async () => {
              if (!confirm(`Bạn có chắc muốn xóa tác vụ ${v.video_id}?`)) return;
              try {
                await API.deleteVideoTask(v.video_id);
                Toast.success(`Đã xóa tác vụ ${v.video_id}`);
                loadData();
              } catch (err) {
                Toast.error(String(err));
              }
            }
          }, '🗑')
        );

        const row = el('tr', {},
          el('td', {}, el('span', { className: `platform-indicator ${platformClass}` }, platformLabel)),
          el('td', {},
            el('div', { className: 'task-cell-title' }, v.video_id),
            el('div', { className: 'task-cell-sub' }, `${v.title || 'Reel task'} • 🎬 ${v.template_name || 'template'}`)
          ),
          el('td', {}, el('span', { style: { fontSize: '12px', color: 'var(--text-muted)' } }, 'content/')),
          el('td', {}, el('span', { style: { fontSize: '12px', color: 'var(--success)', fontWeight: '600' } }, 'Groq / SRT')),
          el('td', {}, badge),
          el('td', { style: { textAlign: 'right' } }, el('div', { style: { display: 'inline-flex', gap: '4px' } }, ...actions))
        );

        tableBody.appendChild(row);
      });
    } catch (err) {
      console.error('Lỗi đọc tasks:', err);
    }
  }

  async function loadData() {
    await Promise.all([loadStats(), loadTasks()]);
  }

  // --- Real-time Listeners ---
  API.onPipelineProgress((event) => {
    progressBar.classList.add('active-running');
    progressBar.style.width = `${Math.round(event.percent)}%`;
    progressPercent.innerText = `${Math.round(event.percent)}%`;
    liveStatusText.innerText = `[${event.video_id}] ${event.message}`;
    loadData();
  });

  API.onPipelineDone(() => {
    progressBar.classList.remove('active-running');
    btnStartPhase1.disabled = false;
    btnStartPhase1.innerText = '▶ BẮT ĐẦU CÀO (PHA 1)';
    liveStatusText.innerText = '✅ Đã hoàn tất Pha 1! Các dự án CapCut đã sẵn sàng để xuất.';
    Toast.success('Hoàn thành Pha 1! Nhấn "⚡ Xuất Video Hàng Loạt" để xuất thành phẩm.');
    loadData();
  });

  API.onExportProgress((event) => {
    progressBar.classList.add('active-running');
    progressBar.style.width = `${Math.round(event.percent)}%`;
    progressPercent.innerText = `${Math.round(event.percent)}%`;
    liveStatusText.innerText = `[Export] ${event.message}`;
    loadData();
  });

  API.onExportDone((res) => {
    progressBar.classList.remove('active-running');
    liveStatusText.innerText = `🎉 Hoàn tất xuất video! Thành công: ${res.success}, Thất bại: ${res.failed}`;
    Toast.success(`Xuất hoàn tất: ${res.success} thành công!`);
    loadData();
  });

  // Load initial data
  loadData();

  // Periodic polling fallback for reliability
  const poller = setInterval(loadData, 4000);

  // Return root container element
  const container = el('div', { style: { display: 'flex', flexDirection: 'column', gap: '18px' } },
    header,
    inputCard,
    progressCard,
    tableCard
  );

  container.refresh = loadData;
  container.onDestroy = () => clearInterval(poller);

  return container;
}
