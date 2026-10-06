// API Bridge Layer: Communicates directly with Wails v2 Go Backend
export const API = {
  isWails: () => typeof window !== 'undefined' && window.go && window.go.main && window.go.main.App,

  // Window Controls
  windowMinimise: () => {
    if (API.isWails()) window.go.main.App.WindowMinimise();
  },
  windowToggleMaximise: () => {
    if (API.isWails()) window.go.main.App.WindowToggleMaximise();
  },
  windowClose: () => {
    if (API.isWails()) window.go.main.App.WindowClose();
  },

  // Video / Task Management
  getPipelineStats: async () => {
    if (API.isWails()) return window.go.main.App.GetPipelineStats();
    return { total: 0, pending: 0, downloaded: 0, draft_created: 0, exported: 0, failed: 0 };
  },

  listVideos: async (filter = 'all') => {
    if (API.isWails()) return window.go.main.App.ListVideos(filter);
    return [];
  },

  addVideoTask: async (url, template, skipDup, extractComments) => {
    if (API.isWails()) {
      return window.go.main.App.AddVideoTask(url, template, skipDup, extractComments);
    }
    const vid = (url.includes('tiktok') ? 'tt_' : 'fb_') + Date.now();
    return { video_id: vid, platform: url.includes('tiktok') ? 'tiktok' : 'facebook', title: url, status: 'pending' };
  },

  deleteVideoTask: async (videoId) => {
    if (API.isWails()) return window.go.main.App.DeleteVideoTask(videoId);
  },

  clearCompletedTasks: async () => {
    if (API.isWails()) return window.go.main.App.ClearCompletedTasks();
  },

  // Pipeline Execution (Phase 1 & Phase 2)
  startPhase1: async (urls, template, skipDup, extractComments) => {
    if (API.isWails()) return window.go.main.App.StartPhase1(urls, template, skipDup, extractComments);
  },

  startPhase2: async () => {
    if (API.isWails()) return window.go.main.App.StartPhase2();
  },

  cancelPipeline: async () => {
    if (API.isWails()) return window.go.main.App.CancelPipeline();
  },

  isPipelineRunning: async () => {
    if (API.isWails()) return window.go.main.App.IsPipelineRunning();
    return false;
  },

  selectLinksFile: async () => {
    if (API.isWails()) return window.go.main.App.SelectLinksFile();
    return [];
  },

  openCapCutProject: async (projectName) => {
    if (API.isWails()) return window.go.main.App.OpenCapCutProject(projectName);
  },

  // Real-time Event Listeners
  onPipelineProgress: (callback) => {
    if (typeof window !== 'undefined' && window.runtime && window.runtime.EventsOn) {
      window.runtime.EventsOn('pipeline:progress', callback);
    }
  },

  onPipelineDone: (callback) => {
    if (typeof window !== 'undefined' && window.runtime && window.runtime.EventsOn) {
      window.runtime.EventsOn('pipeline:done', callback);
    }
  },

  onExportProgress: (callback) => {
    if (typeof window !== 'undefined' && window.runtime && window.runtime.EventsOn) {
      window.runtime.EventsOn('export:progress', callback);
    }
  },

  onExportDone: (callback) => {
    if (typeof window !== 'undefined' && window.runtime && window.runtime.EventsOn) {
      window.runtime.EventsOn('export:done', callback);
    }
  },

  // Tracked Sources Management
  listTrackedSources: async () => {
    if (API.isWails()) return window.go.main.App.ListTrackedSources();
    return [];
  },

  addTrackedSource: async (url, author) => {
    if (API.isWails()) return window.go.main.App.AddTrackedSource(url, author);
  },

  deleteTrackedSource: async (id) => {
    if (API.isWails()) return window.go.main.App.DeleteTrackedSource(id);
  },

  scanTrackedSources: async () => {
    if (API.isWails()) return window.go.main.App.ScanTrackedSources();
  },

  // System & Settings
  getConfig: async () => {
    if (API.isWails()) return window.go.main.App.GetConfig();
    return {
      capcut_draft_path: '',
      template_name: 'template',
      whisper_provider: 'groq_api',
      groq_api_key: '',
      download_dir: 'downloads',
      auto_delete_raw: true,
      skip_duplicates: true
    };
  },

  saveConfig: async (cfg) => {
    if (API.isWails()) return window.go.main.App.SaveConfig(cfg);
  },

  listAvailableTemplates: async () => {
    if (API.isWails()) return window.go.main.App.ListAvailableTemplates();
    return ['template'];
  },

  verifyGroqKey: async (key) => {
    if (API.isWails()) return window.go.main.App.VerifyGroqKey(key);
    return 'Chế độ xem trước (Bypass kết nối)';
  },

  listWhisperModels: async () => {
    if (API.isWails()) return window.go.main.App.ListWhisperModels();
    return [
      { name: 'Whisper Tiny', size_mb: 75, installed: false },
      { name: 'Whisper Base', size_mb: 142, installed: false },
      { name: 'Whisper Small', size_mb: 466, installed: false }
    ];
  },

  downloadWhisperModel: async (modelName) => {
    if (API.isWails()) return window.go.main.App.DownloadWhisperModel(modelName);
  },

  onWhisperDownloadProgress: (callback) => {
    if (typeof window !== 'undefined' && window.runtime && window.runtime.EventsOn) {
      window.runtime.EventsOn('whisper:download:progress', callback);
    }
  },

  onWhisperDownloadDone: (callback) => {
    if (typeof window !== 'undefined' && window.runtime && window.runtime.EventsOn) {
      window.runtime.EventsOn('whisper:download:done', callback);
    }
  },

  getSystemStatus: async () => {
    if (API.isWails()) return window.go.main.App.GetSystemStatus();
    return { free_disk_gb: '128.0', download_dir: 'downloads', capcut_draft_path: 'Tự động phát hiện' };
  },

  openDownloadsFolder: async () => {
    if (API.isWails()) return window.go.main.App.OpenDownloadsFolder();
  },

  openTaskFolder: async (platform, videoId) => {
    if (API.isWails()) return window.go.main.App.OpenTaskFolder(platform, videoId);
  },

  openExportedVideo: async (path) => {
    if (API.isWails()) return window.go.main.App.OpenExportedVideo(path);
  }
};
