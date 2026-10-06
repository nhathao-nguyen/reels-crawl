import { el } from '../utils/dom.js';
import { API } from '../api.js';
import { Toast } from '../components/toast.js';
import { Modal } from '../components/modal.js';

export function createSettingsPage() {
  // Header
  const header = el('div', { className: 'page-header' },
    el('div', {},
      el('h1', { className: 'page-title' }, 'Cài Đặt & Cấu Hình Hệ Thống'),
      el('p', { className: 'page-subtitle' }, 'Thiết lập môi trường CapCut Desktop, API nhận diện giọng nói và thư mục lưu trữ.')
    ),
    el('button', { className: 'btn btn-primary', onClick: saveSettings }, '💾 Lưu Cấu Hình')
  );

  // Section 1: CapCut Desktop & Template
  const settingDraftPath = el('input', {
    type: 'text',
    className: 'form-input',
    placeholder: 'C:\\Users\\...\\AppData\\Local\\CapCut\\User Data\\Projects\\com.lveditor.draft'
  });

  const settingTemplateSelect = el('select', {
    className: 'form-select',
    style: { maxWidth: '320px' }
  },
    el('option', { value: 'template' }, 'Đang tìm template...')
  );

  const section1 = el('div', { className: 'card settings-section' },
    el('div', { className: 'settings-section-title' },
      el('span', {}, '🎬'),
      el('span', {}, '1. Cấu hình CapCut Desktop & Template')
    ),
    el('div', { className: 'form-group' },
      el('label', { className: 'form-label' }, 'Thư mục dự án CapCut (Draft Path):'),
      settingDraftPath,
      el('span', { className: 'form-hint' }, 'Để trống để ứng dụng tự động dò tìm qua Registry và AppData Windows.')
    ),
    el('div', { className: 'form-group' },
      el('label', { className: 'form-label' }, 'Template mẫu mặc định:'),
      settingTemplateSelect,
      el('span', { className: 'form-hint' }, 'Tự động nhận diện mọi dự án CapCut có chứa từ khóa "template" trong tên thư mục.')
    )
  );

  // Section 2: Speech-to-Text
  const sttModeCloud = el('input', { type: 'radio', name: 'stt_mode_pure', value: 'groq_api', checked: true });
  const sttModeLocal = el('input', { type: 'radio', name: 'stt_mode_pure', value: 'local_offline' });

  const settingGroqKey = el('input', {
    type: 'password',
    className: 'form-input',
    placeholder: 'gsk_...'
  });

  const section2 = el('div', { className: 'card settings-section' },
    el('div', { className: 'settings-section-title' },
      el('span', {}, '🧠'),
      el('span', {}, '2. Nhận Diện Giọng Nói (Speech-to-Text)')
    ),
    el('div', { className: 'form-group' },
      el('label', { className: 'form-label' }, 'Phương thức nhận diện:'),
      el('div', { className: 'stt-options-group' },
        el('label', { className: 'form-checkbox-label' },
          sttModeCloud,
          el('strong', {}, 'Cloud API (Groq Whisper Large v3)'),
          ' [Tốc độ siêu nhanh 1-2s]'
        ),
        el('label', { className: 'form-checkbox-label' },
          sttModeLocal,
          el('strong', {}, 'Local Offline (faster-whisper)'),
          ' [Miễn phí, không cần mạng]'
        ),
        el('div', { style: { marginTop: '4px' } },
          el('button', {
            className: 'btn btn-secondary',
            style: { padding: '5px 12px', fontSize: '12px' },
            onClick: () => Modal.open(Modal.createWhisperModal())
          }, '🧠 Quản Lý Tải Model Whisper Offline...')
        )
      )
    ),
    el('div', { className: 'form-group' },
      el('label', { className: 'form-label' }, 'Groq API Key:'),
      el('div', { className: 'api-verify-row' },
        settingGroqKey,
        el('button', { className: 'btn btn-secondary', onClick: verifyGroqKey }, 'Kiểm Tra Kết Nối')
      )
    )
  );

  // Section 3: Storage & Disk
  const settingDownloadDir = el('input', {
    type: 'text',
    className: 'form-input',
    value: 'downloads'
  });

  const settingAutoDeleteRaw = el('input', { type: 'checkbox', checked: true });

  const section3 = el('div', { className: 'card settings-section' },
    el('div', { className: 'settings-section-title' },
      el('span', {}, '💾'),
      el('span', {}, '3. Quản Lý Dung Lượng & Ổ Đĩa')
    ),
    el('div', { className: 'form-group' },
      el('label', { className: 'form-label' }, 'Thư mục tải video thành phẩm:'),
      settingDownloadDir
    ),
    el('label', { className: 'form-checkbox-label' },
      settingAutoDeleteRaw,
      ' Tự động xóa video gốc (raw_video.mp4) sau khi CapCut đã xuất thành phẩm thành công.'
    )
  );

  async function loadConfig() {
    try {
      const cfg = await API.getConfig();
      if (!cfg) return;

      settingDraftPath.value = cfg.capcut_draft_path || '';
      settingGroqKey.value = cfg.groq_api_key || '';
      settingDownloadDir.value = cfg.download_dir || 'downloads';
      settingAutoDeleteRaw.checked = cfg.auto_delete_raw ?? true;

      if (cfg.whisper_provider === 'local_offline') {
        sttModeLocal.checked = true;
      } else {
        sttModeCloud.checked = true;
      }

      // Populate templates dynamically
      const templates = await API.listAvailableTemplates();
      while (settingTemplateSelect.firstChild) {
        settingTemplateSelect.removeChild(settingTemplateSelect.firstChild);
      }
      const currentTpl = cfg.template_name || 'template';

      if (templates && templates.length > 0) {
        templates.forEach(t => {
          settingTemplateSelect.appendChild(el('option', { value: t, selected: t === currentTpl }, t));
        });
        if (!templates.includes(currentTpl)) {
          settingTemplateSelect.appendChild(el('option', { value: currentTpl, selected: true }, currentTpl));
        }
      } else {
        settingTemplateSelect.appendChild(el('option', { value: currentTpl }, currentTpl));
      }
    } catch (err) {
      console.error('Lỗi nạp config:', err);
    }
  }

  async function saveSettings() {
    try {
      const cfg = {
        app_mode: 'gui',
        capcut_draft_path: settingDraftPath.value.trim(),
        template_name: settingTemplateSelect.value || 'template',
        whisper_provider: sttModeCloud.checked ? 'groq_api' : 'local_offline',
        groq_api_key: settingGroqKey.value.trim(),
        local_whisper_model: 'base',
        export_resolution: '1080p',
        download_dir: settingDownloadDir.value.trim() || 'downloads',
        auto_delete_raw: settingAutoDeleteRaw.checked,
        skip_duplicates: true,
        extract_comment_links: true,
        max_chapters_per_story: 20,
        request_delay_seconds: 2
      };

      await API.saveConfig(cfg);
      Toast.success('Đã lưu cấu hình thành công!');
    } catch (err) {
      Toast.error(String(err));
    }
  }

  async function verifyGroqKey() {
    const key = settingGroqKey.value.trim();
    if (!key) {
      Toast.error('Vui lòng nhập API Key trước khi kiểm tra');
      return;
    }
    const res = await API.verifyGroqKey(key);
    Toast.info(res);
  }

  loadConfig();

  const container = el('div', { style: { display: 'flex', flexDirection: 'column', gap: '18px' } },
    header,
    section1,
    section2,
    section3
  );

  return container;
}
