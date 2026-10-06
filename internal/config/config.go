package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Config holds runtime configuration settings for Reels Automation Studio.
type Config struct {
	AppMode             string `json:"app_mode"`
	CapcutDraftPath     string `json:"capcut_draft_path"`
	TemplateName        string `json:"template_name"`
	WhisperProvider     string `json:"whisper_provider"` // groq_api | local_offline
	GroqAPIKey          string `json:"groq_api_key"`
	LocalWhisperModel   string `json:"local_whisper_model"` // tiny | base | small
	ExportResolution    string `json:"export_resolution"`
	DownloadDir         string `json:"download_dir"`
	MaxChaptersPerStory int    `json:"max_chapters_per_story"`
	RequestDelaySeconds int    `json:"request_delay_seconds"`
	AutoDeleteRaw       bool   `json:"auto_delete_raw"`
	SkipDuplicates      bool   `json:"skip_duplicates"`
	ExtractCommentLinks bool   `json:"extract_comment_links"`
}

var (
	defaultConfig = Config{
		AppMode:             "gui",
		CapcutDraftPath:     "",
		TemplateName:        "template-fb",
		WhisperProvider:     "groq_api",
		GroqAPIKey:          "",
		LocalWhisperModel:   "base",
		ExportResolution:    "1080p",
		DownloadDir:         "downloads",
		MaxChaptersPerStory: 20,
		RequestDelaySeconds: 2,
		AutoDeleteRaw:       true,
		SkipDuplicates:      true,
		ExtractCommentLinks: true,
	}
	cfgLock sync.RWMutex
	currentConfig *Config
)

// Default returns a copy of the default configuration.
func Default() Config {
	return defaultConfig
}

// Load loads the configuration from filePath or creates it with defaults.
func Load(filePath string) (*Config, error) {
	cfgLock.Lock()
	defer cfgLock.Unlock()

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		cfg := defaultConfig
		// Try to detect CapCut draft path if empty
		if detected := DetectCapcutDraftPath(); detected != "" {
			cfg.CapcutDraftPath = detected
		}
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err == nil {
			_ = os.WriteFile(filePath, data, 0644)
		}
		currentConfig = &cfg
		return &cfg, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	cfg := defaultConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	currentConfig = &cfg
	return &cfg, nil
}

// Save writes the configuration to filePath.
func Save(filePath string, cfg *Config) error {
	cfgLock.Lock()
	defer cfgLock.Unlock()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	currentConfig = cfg
	return nil
}

// DetectCapcutDraftPath checks common Windows CapCut draft paths across drives.
func DetectCapcutDraftPath() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	userProfile := os.Getenv("USERPROFILE")

	var candidates []string
	if localAppData != "" {
		candidates = append(candidates, filepath.Join(localAppData, "CapCut", "User Data", "Projects", "com.lveditor.draft"))
	}
	if userProfile != "" {
		candidates = append(candidates, filepath.Join(userProfile, "AppData", "Local", "CapCut", "User Data", "Projects", "com.lveditor.draft"))
	}
	// Multi-drive common setups
	candidates = append(candidates,
		`D:\CapCut\User Data\Projects\com.lveditor.draft`,
		`D:\CapCut Drafts\com.lveditor.draft`,
		`C:\CapCut\User Data\Projects\com.lveditor.draft`,
	)

	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return ""
}
