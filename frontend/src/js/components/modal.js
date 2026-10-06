import { el, clear } from '../utils/dom.js';
import { API } from '../api.js';
import { Toast } from './toast.js';

export const Modal = {
  activeModal: null,

  open: (modalElement) => {
    if (Modal.activeModal) Modal.close();
    Modal.activeModal = modalElement;
    document.body.appendChild(modalElement);
    // Trigger animation next tick
    requestAnimationFrame(() => modalElement.classList.add('open'));
  },

  close: () => {
    if (!Modal.activeModal) return;
    const modal = Modal.activeModal;
    modal.classList.remove('open');
    setTimeout(() => {
      if (modal.parentNode) modal.parentNode.removeChild(modal);
      Modal.activeModal = null;
    }, 150);
  },

  createWhisperModal: () => {
    const listContainer = el('div', { style: { display: 'flex', flexDirection: 'column', gap: '10px' } });

    async function loadModels() {
      try {
        clear(listContainer);
        const models = await API.listWhisperModels();
        if (!models || models.length === 0) {
          listContainer.appendChild(el('div', { style: { textAlign: 'center', color: 'var(--text-dim)' } }, 'Không có model nào'));
          return;
        }

        models.forEach(m => {
          const key = m.file_name ? m.file_name.replace('ggml-', '').replace('.bin', '') : m.name.toLowerCase();
          const actionBtn = m.installed
            ? el('span', { className: 'badge badge-exported' }, '✓ Đã Cài Đặt')
            : el('button', {
                className: 'btn btn-secondary',
                style: { fontSize: '12px', padding: '5px 10px' },
                onClick: async () => {
                  try {
                    actionBtn.innerText = '⏳ Đang tải...';
                    actionBtn.disabled = true;
                    await API.downloadWhisperModel(key);
                    Toast.info(`Bắt đầu tải model ${m.name}...`);
                  } catch (e) {
                    Toast.error(String(e));
                  }
                }
              }, `⬇ Tải Về (${m.size_mb} MB)`);

          const itemCard = el('div', {
            className: 'card',
            id: `model-item-${key}`,
            style: { padding: '12px 14px', flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }
          },
            el('div', {},
              el('div', { style: { fontWeight: '600', color: 'var(--text-main)' } }, `${m.name} (~${m.size_mb} MB)`),
              el('div', { style: { fontSize: '11.5px', color: 'var(--text-muted)' } }, m.installed ? `Đã lưu tại models/${m.file_name}` : 'Tải về thư mục models/ để dùng offline')
            ),
            actionBtn
          );
          listContainer.appendChild(itemCard);
        });
      } catch (err) {
        console.error('Lỗi đọc models:', err);
      }
    }

    loadModels();

    API.onWhisperDownloadProgress((ev) => {
      const elItem = document.getElementById(`model-item-${ev.model}`);
      if (elItem) {
        const btn = elItem.querySelector('button');
        if (btn) {
          btn.innerText = `[${Math.round(ev.percent)}%] ${ev.speed}`;
        }
      }
    });

    API.onWhisperDownloadDone((ev) => {
      if (ev.error && ev.error !== '') {
        Toast.error(`Lỗi tải model ${ev.model}: ${ev.error}`);
      } else {
        Toast.success(`Tải thành công model ${ev.model}!`);
      }
      loadModels();
    });

    const modal = el('div', {
      className: 'modal-backdrop',
      onClick: (e) => { if (e.target === modal) Modal.close(); }
    },
      el('div', { className: 'modal-dialog' },
        el('div', { className: 'modal-header' },
          el('div', { className: 'modal-title' },
            el('span', {}, '🧠'),
            el('span', {}, 'Trình Quản Lý Model Whisper Offline')
          ),
          el('button', { className: 'btn-icon', onClick: () => Modal.close() }, '✕')
        ),
        el('div', { className: 'modal-body' },
          el('p', { style: { color: 'var(--text-secondary)', fontSize: '13px', marginBottom: '12px' } },
            'Tải model speech-to-text về máy để sử dụng hoàn toàn offline. Các file model được lưu vào thư mục models/.'
          ),
          listContainer
        ),
        el('div', { className: 'modal-footer' },
          el('button', { className: 'btn btn-secondary', onClick: () => Modal.close() }, 'Đóng')
        )
      )
    );
    return modal;
  },

  createExportModal: (onConfirm) => {
    const modal = el('div', {
      className: 'modal-backdrop',
      onClick: (e) => { if (e.target === modal) Modal.close(); }
    },
      el('div', { className: 'modal-dialog' },
        el('div', { className: 'modal-header' },
          el('div', { className: 'modal-title', style: { color: 'var(--warning)' } },
            el('span', {}, '⚠️'),
            el('span', {}, 'Xác Nhận Xuất Hàng Loạt (Pha 2)')
          ),
          el('button', { className: 'btn-icon', onClick: () => Modal.close() }, '✕')
        ),
        el('div', { className: 'modal-body' },
          el('p', { style: { lineHeight: '1.5' } },
            'Hệ thống sẽ kích hoạt ',
            el('strong', {}, 'Windows UI Automation'),
            ' để điều khiển CapCut Desktop xuất tự động tất cả các dự án đã tạo Draft.'
          ),
          el('div', {
            style: {
              background: 'var(--warning-subtle)',
              border: '1px solid var(--warning-border)',
              padding: '12px',
              borderRadius: 'var(--radius-md)',
              fontSize: '12px',
              color: '#92400e'
            }
          },
            el('strong', {}, 'Lưu ý quan trọng:'),
            el('ul', { style: { marginLeft: '18px', marginTop: '4px', lineHeight: '1.5' } },
              el('li', {}, 'Nếu CapCut đang mở, ứng dụng sẽ đóng để thiết lập lại môi trường.'),
              el('li', {}, 'Không dùng chuột/bàn phím đè lên cửa sổ CapCut khi đang render.'),
              el('li', {}, 'Các video đã xuất thành công trước đó sẽ tự động được bỏ qua.')
            )
          )
        ),
        el('div', { className: 'modal-footer' },
          el('button', { className: 'btn btn-secondary', onClick: () => Modal.close() }, 'Hủy Bỏ'),
          el('button', {
            className: 'btn btn-success',
            onClick: () => {
              Modal.close();
              if (onConfirm) onConfirm();
            }
          }, '⚡ Bắt Đầu Xuất Ngay')
        )
      )
    );
    return modal;
  }
};
