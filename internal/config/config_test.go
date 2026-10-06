package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadAndSave(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reels_test_config_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	// 1. Loading non-existent config should create default file
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}
	if cfg.TemplateName != "template-fb" {
		t.Errorf("Expected default template 'template-fb', got '%s'", cfg.TemplateName)
	}
	if cfg.WhisperProvider != "groq_api" {
		t.Errorf("Expected default whisper provider 'groq_api', got '%s'", cfg.WhisperProvider)
	}

	// 2. Modifying and saving config
	cfg.TemplateName = "custom-template"
	cfg.GroqAPIKey = "gsk_test123"
	if err := Save(configPath, cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// 3. Reloading config should reflect saved changes
	reloaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}
	if reloaded.TemplateName != "custom-template" {
		t.Errorf("Expected 'custom-template', got '%s'", reloaded.TemplateName)
	}
	if reloaded.GroqAPIKey != "gsk_test123" {
		t.Errorf("Expected 'gsk_test123', got '%s'", reloaded.GroqAPIKey)
	}
}
