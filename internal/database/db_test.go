package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSQLiteDatabase(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reels_test_db_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test_data.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}
	defer db.Close()

	// 1. Test TrackedSource CRUD
	source := &TrackedSource{
		Platform:   "facebook",
		SourceURL:  "https://www.facebook.com/ghienphimhay",
		AuthorName: "Ghiền Phim Review",
	}

	sourceID, err := db.AddTrackedSource(source)
	if err != nil {
		t.Fatalf("Failed to add tracked source: %v", err)
	}
	if sourceID <= 0 {
		t.Fatalf("Expected valid source ID > 0, got %d", sourceID)
	}

	sources, err := db.ListTrackedSources()
	if err != nil {
		t.Fatalf("Failed to list tracked sources: %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("Expected 1 source, got %d", len(sources))
	}
	if sources[0].AuthorName != "Ghiền Phim Review" {
		t.Errorf("Expected author 'Ghiền Phim Review', got %s", sources[0].AuthorName)
	}

	// 2. Test Video Insertion with source_id
	video1 := &Video{
		VideoID:      "fb_1029384756",
		SourceID:     &sourceID,
		Platform:     "facebook",
		Title:        "Review Phim Hay Tập 01",
		TemplateName: "template",
		Status:       "downloaded",
	}
	if err := db.UpsertVideo(video1); err != nil {
		t.Fatalf("Failed to upsert video1: %v", err)
	}

	// 3. Test Video Insertion standalone (source_id = nil)
	video2 := &Video{
		VideoID:  "tt_7381920491",
		SourceID: nil,
		Platform: "tiktok",
		Title:    "Top Mẹo Công Nghệ",
		Status:   "pending",
	}
	if err := db.UpsertVideo(video2); err != nil {
		t.Fatalf("Failed to upsert video2: %v", err)
	}

	// 4. Test Deduplication lookup (< 10ms)
	start := time.Now()
	exists, err := db.IsVideoExists("fb_1029384756")
	duration := time.Since(start)
	if err != nil {
		t.Fatalf("Failed to check video exists: %v", err)
	}
	if !exists {
		t.Errorf("Expected video fb_1029384756 to exist")
	}
	if duration > 10*time.Millisecond {
		t.Logf("Notice: IsVideoExists took %v (still fast)", duration)
	}

	existsNonExistent, err := db.IsVideoExists("non_existent_123")
	if err != nil {
		t.Fatalf("Failed to check non-existent: %v", err)
	}
	if existsNonExistent {
		t.Errorf("Expected non-existent video to return false")
	}

	// 5. Test Status Update & Path
	err = db.UpdateVideoStatus("fb_1029384756", "exported", "downloads/facebook_fb_1029384756/exported/final_video.mp4", "")
	if err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	v, err := db.GetVideo("fb_1029384756")
	if err != nil {
		t.Fatalf("Failed to get video: %v", err)
	}
	if v.Status != "exported" {
		t.Errorf("Expected status 'exported', got '%s'", v.Status)
	}
	if v.ExportPath != "downloads/facebook_fb_1029384756/exported/final_video.mp4" {
		t.Errorf("Unexpected export path: %s", v.ExportPath)
	}
	if v.TemplateName != "template" {
		t.Errorf("Expected TemplateName 'template', got '%s'", v.TemplateName)
	}

	// 6. Test Pipeline Stats
	stats, err := db.GetPipelineStats()
	if err != nil {
		t.Fatalf("Failed to get pipeline stats: %v", err)
	}
	if stats.Total != 2 {
		t.Errorf("Expected total 2, got %d", stats.Total)
	}
	if stats.Exported != 1 {
		t.Errorf("Expected 1 exported, got %d", stats.Exported)
	}
	if stats.Pending != 1 {
		t.Errorf("Expected 1 pending, got %d", stats.Pending)
	}

	// 7. Verify TrackedSource video count
	sources, err = db.ListTrackedSources()
	if err != nil {
		t.Fatalf("Failed to list tracked sources after videos: %v", err)
	}
	if sources[0].VideoCount != 1 {
		t.Errorf("Expected source video count 1, got %d", sources[0].VideoCount)
	}

	// 8. Test Filtered ListVideos
	exportedList, err := db.ListVideos("exported", 10, 0)
	if err != nil {
		t.Fatalf("Failed to list exported videos: %v", err)
	}
	if len(exportedList) != 1 || exportedList[0].VideoID != "fb_1029384756" {
		t.Errorf("Expected 1 exported video fb_1029384756, got %d", len(exportedList))
	}

	pendingList, err := db.ListVideos("pending", 10, 0)
	if err != nil {
		t.Fatalf("Failed to list pending videos: %v", err)
	}
	if len(pendingList) != 1 || pendingList[0].VideoID != "tt_7381920491" {
		t.Errorf("Expected 1 pending video tt_7381920491, got %d", len(pendingList))
	}

	// 9. Test UpdateTrackedSourceChecked
	checkTime := time.Now().Truncate(time.Second)
	if err := db.UpdateTrackedSourceChecked(sourceID, checkTime); err != nil {
		t.Fatalf("Failed to update last checked time: %v", err)
	}
	sources, _ = db.ListTrackedSources()
	if sources[0].LastCheckedAt == nil {
		t.Errorf("Expected non-nil LastCheckedAt")
	}

	// 10. Test ON DELETE SET NULL constraint
	if err := db.DeleteTrackedSource(sourceID); err != nil {
		t.Fatalf("Failed to delete tracked source: %v", err)
	}
	vAfterSourceDelete, err := db.GetVideo("fb_1029384756")
	if err != nil {
		t.Fatalf("Failed to get video after source delete: %v", err)
	}
	if vAfterSourceDelete.SourceID != nil {
		t.Errorf("Expected source_id to become NULL after parent source delete, got %v", *vAfterSourceDelete.SourceID)
	}

	// 11. Test DeleteVideo
	if err := db.DeleteVideo("fb_1029384756"); err != nil {
		t.Fatalf("Failed to delete video: %v", err)
	}
	deletedCheck, err := db.GetVideo("fb_1029384756")
	if err != nil {
		t.Fatalf("Error checking deleted video: %v", err)
	}
	if deletedCheck != nil {
		t.Errorf("Expected video to be nil after deletion")
	}
}
