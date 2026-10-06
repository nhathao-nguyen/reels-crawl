package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"reels-crawl/internal/capcut"
	"reels-crawl/internal/config"
	"reels-crawl/internal/crawler"
	"reels-crawl/internal/database"
	"reels-crawl/internal/pipeline"
	"reels-crawl/internal/subtitles"
	"reels-crawl/internal/system"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct manages application state and exposes bridge methods to the Wails frontend.
type App struct {
	ctx      context.Context
	db       *database.DB
	cfg      *config.Config
	pipeline *pipeline.Service
	dbPath   string
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{
		dbPath: "data.db",
	}
}

// startup is called when Wails initializes the application.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 1. Load or initialize configuration
	cfg, err := config.Load("config.json")
	if err != nil {
		runtime.LogErrorf(ctx, "Failed to load config: %v", err)
		defaultCfg := config.Default()
		a.cfg = &defaultCfg
	} else {
		a.cfg = cfg
	}

	// 2. Open SQLite database & run migrations
	db, err := database.Open(a.dbPath)
	if err != nil {
		runtime.LogErrorf(ctx, "Failed to open SQLite database: %v", err)
	} else {
		a.db = db
		runtime.LogInfo(ctx, "SQLite database connected successfully")
	}

	// 3. Ensure download directory exists
	downloadDir := a.cfg.DownloadDir
	if downloadDir == "" {
		downloadDir = "downloads"
	}
	_ = os.MkdirAll(downloadDir, 0755)

	// 4. Initialize pipeline service
	a.pipeline = pipeline.NewService(a.db, a.cfg)
}

// shutdown is called when Wails terminates the application.
func (a *App) shutdown(ctx context.Context) {
	if a.pipeline != nil {
		a.pipeline.Cancel()
	}
	if a.db != nil {
		_ = a.db.Close()
	}
}

// --- Window Controls (Custom Titlebar) ---

func (a *App) WindowMinimise() {
	runtime.WindowMinimise(a.ctx)
}

func (a *App) WindowToggleMaximise() {
	runtime.WindowToggleMaximise(a.ctx)
}

func (a *App) WindowClose() {
	runtime.Quit(a.ctx)
}

// --- Pipeline & Video Management ---

// GetPipelineStats returns status count summaries for the Dashboard.
func (a *App) GetPipelineStats() (*database.PipelineStats, error) {
	if a.db == nil {
		return &database.PipelineStats{}, fmt.Errorf("database not connected")
	}
	return a.db.GetPipelineStats()
}

// ListVideos returns tasks filtered by status ("all", "downloaded", "draft_created", "exported", "failed").
func (a *App) ListVideos(statusFilter string) ([]*database.Video, error) {
	if a.db == nil {
		return []*database.Video{}, fmt.Errorf("database not connected")
	}
	return a.db.ListVideos(statusFilter, 100, 0)
}

// AddVideoTask adds a new video task to the pipeline.
func (a *App) AddVideoTask(urlStr string, templateName string, skipDuplicates bool, extractComments bool) (*database.Video, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	urlStr = strings.TrimSpace(urlStr)
	if urlStr == "" {
		return nil, fmt.Errorf("đường dẫn không được để trống")
	}

	// Detect platform & extract ID consistently
	platform := crawler.DetectPlatform(urlStr)
	videoID := crawler.ExtractVideoID(urlStr)
	if videoID == "" {
		videoID = fmt.Sprintf("%d", time.Now().UnixNano()%10000000)
	}

	// Check deduplication if requested
	if skipDuplicates {
		exists, err := a.db.IsVideoExists(videoID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("video %s đã có trong cơ sở dữ liệu (Deduplicated)", videoID)
		}
	}

	templateName = strings.TrimSpace(templateName)
	if templateName == "" && a.cfg != nil {
		templateName = a.cfg.TemplateName
	}
	if templateName == "" {
		templateName = "template"
	}

	video := &database.Video{
		VideoID:      videoID,
		Platform:     platform,
		Title:        fmt.Sprintf("Video %s", videoID),
		TemplateName: templateName,
		Status:       "pending",
	}

	if err := a.db.UpsertVideo(video); err != nil {
		return nil, fmt.Errorf("không thể thêm video: %w", err)
	}

	return video, nil
}

// StartPhase1 triggers Phase 1 Crawl & Draft creation in the background
func (a *App) StartPhase1(urls []string, templateName string, skipDuplicates bool, extractComments bool) error {
	if a.pipeline == nil {
		return fmt.Errorf("pipeline chưa sẵn sàng")
	}
	if a.pipeline.IsRunning() {
		return fmt.Errorf("tiến trình đang chạy, vui lòng đợi hoàn tất")
	}

	go func() {
		a.pipeline.BatchProcessURLs(
			context.Background(),
			urls,
			templateName,
			skipDuplicates,
			extractComments,
			func(event pipeline.ProgressEvent) {
				runtime.EventsEmit(a.ctx, "pipeline:progress", event)
			},
		)
		runtime.EventsEmit(a.ctx, "pipeline:done", true)
	}()

	return nil
}

// StartPhase2 triggers Phase 2 Auto-Export in the background
func (a *App) StartPhase2() error {
	if a.pipeline == nil {
		return fmt.Errorf("pipeline chưa sẵn sàng")
	}
	if a.pipeline.IsRunning() {
		return fmt.Errorf("tiến trình đang chạy, vui lòng đợi hoàn tất")
	}

	go func() {
		success, failed, err := a.pipeline.StartBatchExport(
			context.Background(),
			func(videoID string, percent float64, message string, status string) {
				runtime.EventsEmit(a.ctx, "export:progress", map[string]any{
					"video_id": videoID,
					"percent":  percent,
					"message":  message,
					"status":   status,
				})
			},
		)
		runtime.EventsEmit(a.ctx, "export:done", map[string]any{
			"success": success,
			"failed":  failed,
			"error":   fmt.Sprintf("%v", err),
		})
	}()

	return nil
}

// CancelPipeline stops any currently running Phase 1 or Phase 2 process
func (a *App) CancelPipeline() {
	if a.pipeline != nil {
		a.pipeline.Cancel()
	}
}

// IsPipelineRunning checks whether the pipeline is executing
func (a *App) IsPipelineRunning() bool {
	if a.pipeline == nil {
		return false
	}
	return a.pipeline.IsRunning()
}

// SelectLinksFile opens a Windows file dialog to pick a .txt file containing links
func (a *App) SelectLinksFile() ([]string, error) {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Chọn file danh sách link (.txt)",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Text Files (*.txt)",
				Pattern:     "*.txt",
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if filePath == "" {
		return nil, nil // user cancelled
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("không thể mở file: %w", err)
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		line = strings.TrimPrefix(line, "\ufeff")
		if line != "" && !strings.HasPrefix(line, "#") {
			urls = append(urls, line)
		}
	}

	return urls, nil
}

// OpenCapCutProject opens the project draft folder or launches CapCut
func (a *App) OpenCapCutProject(projectName string) error {
	draftPath := ""
	if a.cfg != nil {
		draftPath = a.cfg.CapcutDraftPath
	}
	if draftPath == "" {
		draftPath = config.DetectCapcutDraftPath()
	}
	projectDir := filepath.Join(draftPath, projectName)
	if _, err := os.Stat(projectDir); err == nil {
		_ = system.OpenFolder(projectDir)
	}
	return capcut.OpenProjectInCapCut("")
}

// DeleteVideoTask deletes a video record from the database.
func (a *App) DeleteVideoTask(videoID string) error {
	if a.db == nil {
		return fmt.Errorf("database not connected")
	}
	return a.db.DeleteVideo(videoID)
}

// ClearCompletedTasks removes tasks with status 'exported'.
func (a *App) ClearCompletedTasks() error {
	if a.db == nil {
		return fmt.Errorf("database not connected")
	}
	_, err := a.db.ClearCompletedVideos()
	return err
}

// --- Tracked Sources Management ---

// ListTrackedSources returns all tracked channels/profiles.
func (a *App) ListTrackedSources() ([]*database.TrackedSource, error) {
	if a.db == nil {
		return []*database.TrackedSource{}, fmt.Errorf("database not connected")
	}
	return a.db.ListTrackedSources()
}

// AddTrackedSource registers a new page/profile for continuous monitoring.
func (a *App) AddTrackedSource(sourceURL string, authorName string) (*database.TrackedSource, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	sourceURL = strings.TrimSpace(sourceURL)
	if sourceURL == "" {
		return nil, fmt.Errorf("đường dẫn kênh không được để trống")
	}

	platform := "facebook"
	if strings.Contains(sourceURL, "tiktok.com") {
		platform = "tiktok"
	}

	source := &database.TrackedSource{
		Platform:   platform,
		SourceURL:  sourceURL,
		AuthorName: authorName,
	}

	id, err := a.db.AddTrackedSource(source)
	if err != nil {
		return nil, fmt.Errorf("lỗi thêm kênh: %w", err)
	}
	source.ID = id
	return source, nil
}

// DeleteTrackedSource removes a tracked source.
func (a *App) DeleteTrackedSource(id int64) error {
	if a.db == nil {
		return fmt.Errorf("database not connected")
	}
	return a.db.DeleteTrackedSource(id)
}

// ScanTrackedSources scans all registered sources in the database for new videos
func (a *App) ScanTrackedSources() error {
	if a.pipeline == nil {
		return fmt.Errorf("pipeline chưa sẵn sàng")
	}
	if a.pipeline.IsRunning() {
		return fmt.Errorf("tiến trình đang chạy, vui lòng đợi hoàn tất")
	}
	if a.db == nil {
		return fmt.Errorf("database chưa kết nối")
	}

	sources, err := a.db.ListTrackedSources()
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		return fmt.Errorf("chưa có kênh nào được theo dõi")
	}

	go func() {
		for _, s := range sources {
			now := time.Now()
			_ = a.db.UpdateTrackedSourceChecked(s.ID, now)

			templateName := "template"
			if a.cfg != nil && a.cfg.TemplateName != "" {
				templateName = a.cfg.TemplateName
			}

			_, _ = a.pipeline.ProcessURLWithSource(
				context.Background(),
				s.SourceURL,
				templateName,
				true,
				true,
				&s.ID,
				func(event pipeline.ProgressEvent) {
					runtime.EventsEmit(a.ctx, "pipeline:progress", event)
				},
			)
		}
		runtime.EventsEmit(a.ctx, "pipeline:done", true)
	}()

	return nil
}

// --- Settings & System Info ---

// GetConfig returns the current configuration.
func (a *App) GetConfig() *config.Config {
	return a.cfg
}

// SaveConfig saves updated settings to config.json.
func (a *App) SaveConfig(cfg config.Config) error {
	err := config.Save("config.json", &cfg)
	if err == nil {
		a.cfg = &cfg
		if a.pipeline != nil {
			a.pipeline.UpdateConfig(&cfg)
		}
	}
	return err
}

// VerifyGroqKey tests connectivity to Groq API using the provided key.
func (a *App) VerifyGroqKey(apiKey string) string {
	_, msg := system.VerifyGroqAPIKey(apiKey)
	return msg
}

// ListWhisperModels returns available and installed local Whisper models.
func (a *App) ListWhisperModels() []subtitles.WhisperModelInfo {
	return subtitles.ListWhisperModels("models")
}

// DownloadWhisperModel downloads a model in background and emits events to Wails.
func (a *App) DownloadWhisperModel(modelName string) error {
	go func() {
		err := subtitles.DownloadWhisperModel(modelName, "models", func(pct float64, speed string) {
			runtime.EventsEmit(a.ctx, "whisper:download:progress", map[string]any{
				"model":   modelName,
				"percent": pct,
				"speed":   speed,
			})
		})
		runtime.EventsEmit(a.ctx, "whisper:download:done", map[string]any{
			"model": modelName,
			"error": fmt.Sprintf("%v", err),
		})
	}()
	return nil
}

// ListAvailableTemplates scans the CapCut drafts directory for all projects containing the word 'template' (case-insensitive).
func (a *App) ListAvailableTemplates() []string {
	draftPath := ""
	if a.cfg != nil {
		draftPath = a.cfg.CapcutDraftPath
	}
	if draftPath == "" {
		draftPath = config.DetectCapcutDraftPath()
	}
	templates := system.FindCapcutTemplates(draftPath)
	if len(templates) == 0 {
		return []string{"template"}
	}
	return templates
}

// GetSystemStatus returns system stats (disk space, download path, etc.).
func (a *App) GetSystemStatus() map[string]interface{} {
	downloadDir := "downloads"
	if a.cfg != nil && a.cfg.DownloadDir != "" {
		downloadDir = a.cfg.DownloadDir
	}
	absDownloads, _ := filepath.Abs(downloadDir)

	freeGB := system.GetFreeDiskSpaceGB(absDownloads)

	capcutDraftPath := ""
	if a.cfg != nil {
		capcutDraftPath = a.cfg.CapcutDraftPath
	}

	return map[string]interface{}{
		"free_disk_gb":       fmt.Sprintf("%.1f", freeGB),
		"download_dir":       absDownloads,
		"capcut_draft_path":  capcutDraftPath,
		"template_name":      a.cfg.TemplateName,
		"whisper_provider":   a.cfg.WhisperProvider,
	}
}

// OpenDownloadsFolder opens the downloads folder in Windows Explorer.
func (a *App) OpenDownloadsFolder() error {
	downloadDir := "downloads"
	if a.cfg != nil && a.cfg.DownloadDir != "" {
		downloadDir = a.cfg.DownloadDir
	}
	return system.OpenFolder(downloadDir)
}

// OpenTaskFolder opens a specific task folder in Windows Explorer.
func (a *App) OpenTaskFolder(platform string, videoID string) error {
	downloadDir := "downloads"
	if a.cfg != nil && a.cfg.DownloadDir != "" {
		downloadDir = a.cfg.DownloadDir
	}
	folderName := fmt.Sprintf("%s_%s", platform, videoID)
	if strings.HasPrefix(videoID, platform+"_") {
		folderName = videoID
	}
	target := filepath.Join(downloadDir, folderName)
	_ = os.MkdirAll(target, 0755)
	return system.OpenFolder(target)
}

// OpenExportedVideo opens the final video in the default media player.
func (a *App) OpenExportedVideo(exportPath string) error {
	if exportPath == "" {
		return fmt.Errorf("chưa có video xuất thành phẩm")
	}
	return system.OpenFile(exportPath)
}

// Helper: extractVideoID attempts regex matches for FB and TikTok URLs.
func extractVideoID(urlStr string) string {
	return crawler.ExtractVideoID(urlStr)
}
