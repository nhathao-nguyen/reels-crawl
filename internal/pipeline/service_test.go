package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"reels-crawl/internal/config"
	"reels-crawl/internal/database"
)

func TestPipelineService(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_pipeline.db")

	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	cfg := config.Default()
	cfg.DownloadDir = filepath.Join(tempDir, "downloads")
	_ = os.MkdirAll(cfg.DownloadDir, 0755)

	svc := NewService(db, &cfg)
	if svc == nil {
		t.Fatalf("Expected non-nil service")
	}

	if svc.IsRunning() {
		t.Errorf("Service should not be running initially")
	}

	if svc.Downloader() == nil {
		t.Errorf("Expected non-nil Downloader")
	}

	// Test deduplication in ProcessSingleURL
	videoID := "test_dedup_001"
	existingVideo := &database.Video{
		VideoID:  videoID,
		Platform: "facebook",
		Title:    "Existing Video",
		Status:   "draft_created",
	}
	if err := db.UpsertVideo(existingVideo); err != nil {
		t.Fatalf("Failed to seed existing video: %v", err)
	}

	var events []ProgressEvent
	res, err := svc.ProcessSingleURL(
		context.Background(),
		"https://www.facebook.com/reel/test_dedup_001",
		"template",
		true, // skip duplicates
		true,
		func(e ProgressEvent) {
			events = append(events, e)
		},
	)

	if err != nil {
		t.Fatalf("ProcessSingleURL failed: %v", err)
	}
	if res == nil {
		t.Fatalf("Expected non-nil video record for deduplicated video")
	}
	if res.VideoID != videoID {
		t.Errorf("Expected video ID %s, got %s", videoID, res.VideoID)
	}
	if len(events) == 0 {
		t.Errorf("Expected progress events to be emitted")
	}

	// Verify last event is complete with Deduplication message
	lastEvent := events[len(events)-1]
	if lastEvent.Percent != 100 {
		t.Errorf("Expected 100 percent on dedup complete, got %f", lastEvent.Percent)
	}

	// Test StartBatchExport when no pending drafts exist
	success, failed, err := svc.StartBatchExport(
		context.Background(),
		nil,
	)
	if err != nil {
		t.Fatalf("StartBatchExport failed: %v", err)
	}
	// The existing video was already draft_created, but raw_video doesn't exist on disk so it handles cleanly
	t.Logf("StartBatchExport completed: success=%d, failed=%d", success, failed)
}

func TestEndToEndPipelineExecution(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "e2e_pipeline.db")

	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// 1. Add tracked source
	src := &database.TrackedSource{
		Platform:   "facebook",
		SourceURL:  "https://www.facebook.com/e2e_channel",
		AuthorName: "E2E Channel",
	}
	srcID, err := db.AddTrackedSource(src)
	if err != nil {
		t.Fatalf("Failed to add tracked source: %v", err)
	}

	cfg := config.Default()
	cfg.DownloadDir = filepath.Join(tempDir, "downloads")
	cfg.AutoDeleteRaw = false
	_ = os.MkdirAll(cfg.DownloadDir, 0755)

	svc := NewService(db, &cfg)

	// Setup mock downloaded files using real test_video.mp4 if present in project
	realVideo := filepath.Join("..", "..", "downloads", "test_video.mp4")
	if _, err := os.Stat(realVideo); err != nil {
		t.Skip("test_video.mp4 not found, skipping full E2E render")
	}

	videoID := "e2e_test_998877"
	targetDir := filepath.Join(cfg.DownloadDir, "facebook_"+videoID)
	_ = os.MkdirAll(targetDir, 0755)

	rawVideoDest := filepath.Join(targetDir, "raw_video.mp4")
	videoData, err := os.ReadFile(realVideo)
	if err != nil {
		t.Fatalf("Failed to read test_video: %v", err)
	}
	if err := os.WriteFile(rawVideoDest, videoData, 0644); err != nil {
		t.Fatalf("Failed to copy test video: %v", err)
	}

	// Run Phase 1 with existing raw_video.mp4 (Deduplication / pre-downloaded path)
	var events []ProgressEvent
	record, err := svc.ProcessURLWithSource(
		context.Background(),
		"https://www.facebook.com/reel/"+videoID,
		"template",
		false, // do not skip
		true,
		&srcID,
		func(e ProgressEvent) {
			events = append(events, e)
		},
	)

	if err != nil {
		t.Fatalf("ProcessURLWithSource failed: %v", err)
	}
	if record == nil {
		t.Fatalf("Expected valid video record")
	}
	if record.Status != "draft_created" {
		t.Errorf("Expected status 'draft_created', got '%s'", record.Status)
	}
	if record.SourceID == nil || *record.SourceID != srcID {
		t.Errorf("Expected source ID %d, got %v", srcID, record.SourceID)
	}

	// Verify Phase 1 file outputs
	if _, err := os.Stat(filepath.Join(targetDir, "subtitles.srt")); err != nil {
		t.Errorf("Expected subtitles.srt to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(targetDir, "subtitles.json")); err != nil {
		t.Errorf("Expected subtitles.json to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(targetDir, "metadata.json")); err != nil {
		t.Errorf("Expected metadata.json to exist: %v", err)
	}

	// 2. Run Phase 2: StartBatchExport
	success, failed, err := svc.StartBatchExport(
		context.Background(),
		nil,
	)
	if err != nil {
		t.Fatalf("StartBatchExport failed: %v", err)
	}
	if success != 1 {
		v, _ := db.GetVideo(videoID)
		if v != nil {
			t.Logf("Video error in DB: %s", v.ErrorMessage)
		}
		t.Errorf("Expected 1 successful export, got %d (failed: %d)", success, failed)
	}

	// Verify Phase 2 file output: final_video.mp4
	finalExport := filepath.Join(targetDir, "exported", "final_video.mp4")
	fi, err := os.Stat(finalExport)
	if err != nil || fi.Size() < 1024 {
		t.Fatalf("Expected valid final_video.mp4 to be exported: %v", err)
	}

	// Verify DB status is now 'exported'
	updatedRec, err := db.GetVideo(videoID)
	if err != nil || updatedRec == nil {
		t.Fatalf("Failed to fetch updated video: %v", err)
	}
	if updatedRec.Status != "exported" {
		t.Errorf("Expected DB status 'exported', got '%s'", updatedRec.Status)
	}
	if updatedRec.ExportPath == "" {
		t.Errorf("Expected non-empty ExportPath")
	}

	// 3. Test Deduplication Skip on re-run
	dedupRec, err := svc.ProcessSingleURL(
		context.Background(),
		"https://www.facebook.com/reel/"+videoID,
		"template",
		true, // skip duplicates
		true,
		nil,
	)
	if err != nil {
		t.Fatalf("Dedup re-run failed: %v", err)
	}
	if dedupRec.VideoID != videoID {
		t.Errorf("Expected video ID %s on dedup, got %s", videoID, dedupRec.VideoID)
	}
}
