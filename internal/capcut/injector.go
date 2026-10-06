package capcut

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"reels-crawl/internal/subtitles"
)

// GenerateUUID creates a standard UUID v4 string
func GenerateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant RFC 4122
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// DeepCopyMap creates a deep clone of a map via JSON serialize/deserialize
func DeepCopyMap(src map[string]any) (map[string]any, error) {
	data, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	var dst map[string]any
	err = json.Unmarshal(data, &dst)
	return dst, err
}

// CopyDir recursively copies a directory
func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".tmp") || strings.HasSuffix(info.Name(), ".bak") {
			return nil
		}
		targetPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}
		return CopyFile(path, targetPath)
	})
}

// CopyFile copies a single file
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// InjectResult contains metadata about the newly generated CapCut project
type InjectResult struct {
	DraftID      string `json:"draft_id"`
	ProjectName  string `json:"project_name"`
	ProjectDir   string `json:"project_dir"`
	DurationSec  float64 `json:"duration_sec"`
	DurationUs   int64   `json:"duration_us"`
	SubtitlesQty int     `json:"subtitles_qty"`
}

// EnsureTemplateDirectory locates or provisions a valid CapCut template folder
func EnsureTemplateDirectory(baseDraftDir, templateName string) (string, error) {
	if baseDraftDir == "" {
		baseDraftDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "CapCut", "User Data", "Projects", "com.lveditor.draft")
	}

	// 1. Direct match
	targetDir := filepath.Join(baseDraftDir, templateName)
	if _, err := os.Stat(targetDir); err == nil {
		return targetDir, nil
	}

	// 2. Alternative "template" or "template-fb" match
	altNames := []string{"template", "template-fb", "Template", "template_reels"}
	for _, alt := range altNames {
		candidate := filepath.Join(baseDraftDir, alt)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	// 3. Scan baseDraftDir for ANY folder with 'template' in its name
	if entries, err := os.ReadDir(baseDraftDir); err == nil {
		for _, e := range entries {
			if e.IsDir() && strings.Contains(strings.ToLower(e.Name()), "template") {
				return filepath.Join(baseDraftDir, e.Name()), nil
			}
		}
	}

	// 4. Provision from bundled templates if available
	bundledCandidates := []string{
		filepath.Join(".", "templates", templateName),
		filepath.Join("..", "templates", templateName),
		filepath.Join("..", "..", "templates", templateName),
		filepath.Join(".", "templates", "template-fb"),
		filepath.Join("..", "templates", "template-fb"),
		filepath.Join("..", "..", "templates", "template-fb"),
		filepath.Join(".", "templates", "template"),
		filepath.Join("..", "templates", "template"),
		filepath.Join("..", "..", "templates", "template"),
	}
	for _, b := range bundledCandidates {
		if info, err := os.Stat(b); err == nil && info.IsDir() {
			destDir := filepath.Join(baseDraftDir, templateName)
			if err := CopyDir(b, destDir); err == nil {
				return destDir, nil
			}
		}
	}

	return "", fmt.Errorf("Chưa phát hiện Template chuẩn! Vui lòng mở CapCut tạo 1 dự án mẫu tên '%s' có sẵn text/style mẫu trước khi chạy.", templateName)
}

// CreateDraftFromTemplate clones a template and injects dynamic video and subtitles
func CreateDraftFromTemplate(
	baseDraftDir string,
	templateName string,
	projectName string,
	rawVideoPath string,
	durationSec float64,
	subs []subtitles.SubtitleSegment,
) (*InjectResult, error) {
	if baseDraftDir == "" {
		baseDraftDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "CapCut", "User Data", "Projects", "com.lveditor.draft")
	}

	templateDir, err := EnsureTemplateDirectory(baseDraftDir, templateName)
	if err != nil {
		return nil, err
	}

	targetProjectDir := filepath.Join(baseDraftDir, projectName)
	_ = os.RemoveAll(targetProjectDir)

	if err := CopyDir(templateDir, targetProjectDir); err != nil {
		return nil, fmt.Errorf("lỗi sao chép template sang %s: %w", targetProjectDir, err)
	}

	// Extract real video frame to draft_cover.jpg for CapCut home screen preview
	ffmpegBin := subtitles.FindFFmpeg()
	if ffmpegBin != "" {
		coverPath := filepath.Join(targetProjectDir, "draft_cover.jpg")
		cmd := exec.Command(ffmpegBin, "-ss", "00:00:01", "-i", rawVideoPath, "-vframes", "1", "-q:v", "2", coverPath, "-y")
		_ = cmd.Run()
	}

	absVideoPath, err := filepath.Abs(rawVideoPath)
	if err != nil {
		absVideoPath = rawVideoPath
	}
	cleanVideoPath := filepath.ToSlash(absVideoPath)

	totalDurationUs := int64(durationSec * 1e6)
	if totalDurationUs <= 0 {
		totalDurationUs = 10000000 // default 10s
	}

	// 1. Process draft_content.json
	draftContentPath := filepath.Join(targetProjectDir, "draft_content.json")
	contentBytes, err := os.ReadFile(draftContentPath)
	if err != nil {
		return nil, fmt.Errorf("lỗi đọc draft_content.json: %w", err)
	}

	var root map[string]any
	if err := json.Unmarshal(contentBytes, &root); err != nil {
		return nil, fmt.Errorf("lỗi parse draft_content.json: %w", err)
	}

	root["duration"] = totalDurationUs

	materials, ok := root["materials"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("không tìm thấy materials trong draft_content.json")
	}

	// Replace video material
	if videos, ok := materials["videos"].([]any); ok && len(videos) > 0 {
		if v0, ok := videos[0].(map[string]any); ok {
			v0["path"] = cleanVideoPath
			v0["duration"] = totalDurationUs
			v0["material_name"] = filepath.Base(cleanVideoPath)
			v0["has_audio"] = true
		}
	}

	tracks, ok := root["tracks"].([]any)
	if !ok {
		return nil, fmt.Errorf("không tìm thấy tracks trong draft_content.json")
	}

	var textTrackIndex = -1
	var bestSubtitleScore = -1
	var textMaterialTemplate map[string]any
	var textSegmentTemplate map[string]any

	for i, tr := range tracks {
		trMap, ok := tr.(map[string]any)
		if !ok {
			continue
		}
		tType, _ := trMap["type"].(string)

		if tType == "video" {
			if segs, ok := trMap["segments"].([]any); ok && len(segs) > 0 {
				if seg0, ok := segs[0].(map[string]any); ok {
					if trgt, ok := seg0["target_timerange"].(map[string]any); ok {
						trgt["start"] = 0
						trgt["duration"] = totalDurationUs
					}
					if src, ok := seg0["source_timerange"].(map[string]any); ok {
						src["start"] = 0
						src["duration"] = totalDurationUs
					}
					// Ensure video audio volume is active so the newly injected video sound is played
					seg0["volume"] = 1.0
					seg0["last_nonzero_volume"] = 1.0
				}
			}
		} else if tType == "audio" {
			// Filter out any stale extracted audio segments from the old template video
			if segs, ok := trMap["segments"].([]any); ok {
				var preservedAudioSegs []any
				for _, seg := range segs {
					if sMap, ok := seg.(map[string]any); ok {
						matID, _ := sMap["material_id"].(string)
						isOldExtracted := false
						if audios, ok := materials["audios"].([]any); ok {
							for _, a := range audios {
								if aMap, ok := a.(map[string]any); ok && aMap["id"] == matID {
									if aType, _ := aMap["type"].(string); aType == "extract_music" {
										isOldExtracted = true
										break
									}
								}
							}
						}
						if !isOldExtracted {
							preservedAudioSegs = append(preservedAudioSegs, seg)
						}
					}
				}
				trMap["segments"] = preservedAudioSegs
			}
		} else if tType == "text" {
			segs, ok := trMap["segments"].([]any)
			if !ok || len(segs) == 0 {
				continue
			}

			score := len(segs)
			// Check transform.y of the first segment
			if s0, ok := segs[0].(map[string]any); ok {
				if clip, ok := s0["clip"].(map[string]any); ok {
					if tf, ok := clip["transform"].(map[string]any); ok {
						if y, ok := tf["y"].(float64); ok && y <= 0 {
							score += 100 // Boost track positioned in bottom half of screen
						}
					}
				}
			}

			if score > bestSubtitleScore {
				bestSubtitleScore = score
				textTrackIndex = i
				if s0, ok := segs[0].(map[string]any); ok {
					textSegmentTemplate, _ = DeepCopyMap(s0)
					// Locate matching material in materials["texts"]
					matID, _ := s0["material_id"].(string)
					if texts, ok := materials["texts"].([]any); ok {
						for _, t := range texts {
							if tMap, ok := t.(map[string]any); ok {
								if tMap["id"] == matID {
									textMaterialTemplate, _ = DeepCopyMap(tMap)
									break
								}
							}
						}
					}
				}
			}
		}
	}

	// Fallback to first text material if not directly matched
	if textMaterialTemplate == nil {
		if texts, ok := materials["texts"].([]any); ok && len(texts) > 0 {
			if t0, ok := texts[0].(map[string]any); ok {
				textMaterialTemplate, _ = DeepCopyMap(t0)
			}
		}
	}

	// Inject dynamic subtitles if template has text capabilities
	if textMaterialTemplate != nil && textSegmentTemplate != nil && len(subs) > 0 && textTrackIndex >= 0 {
		newTexts := make([]any, 0, len(subs))
		newTextSegments := make([]any, 0, len(subs))

		// Collect old material IDs that belonged to this specific subtitle track to replace them cleanly
		oldMatIDs := make(map[string]bool)
		trMap := tracks[textTrackIndex].(map[string]any)
		if oldSegs, ok := trMap["segments"].([]any); ok {
			for _, os := range oldSegs {
				if osMap, ok := os.(map[string]any); ok {
					if mid, ok := osMap["material_id"].(string); ok {
						oldMatIDs[mid] = true
					}
				}
			}
		}

		for _, sub := range subs {
			newMatID := GenerateUUID()
			newSegID := GenerateUUID()

			newMat, err := DeepCopyMap(textMaterialTemplate)
			if err != nil {
				continue
			}
			newMat["id"] = newMatID

			// Update text and formatting range
			if contentRaw, ok := newMat["content"].(string); ok {
				var contentObj map[string]any
				if err := json.Unmarshal([]byte(contentRaw), &contentObj); err == nil {
					contentObj["text"] = sub.Text
					runesCount := len([]rune(sub.Text))
					if styles, ok := contentObj["styles"].([]any); ok && len(styles) > 0 {
						if s0, ok := styles[0].(map[string]any); ok {
							s0["range"] = []int{0, runesCount}
						}
					}
					updatedContent, _ := json.Marshal(contentObj)
					newMat["content"] = string(updatedContent)
				}
			}
			newTexts = append(newTexts, newMat)

			newSeg, err := DeepCopyMap(textSegmentTemplate)
			if err != nil {
				continue
			}
			newSeg["id"] = newSegID
			newSeg["material_id"] = newMatID
			newSeg["target_timerange"] = map[string]any{
				"start":    sub.StartUs,
				"duration": sub.DurationUs,
			}
			newSeg["source_timerange"] = nil
			newTextSegments = append(newTextSegments, newSeg)
		}

		// Keep any other text materials belonging to OTHER tracks (e.g. titles, watermarks)
		var preservedTexts []any
		if texts, ok := materials["texts"].([]any); ok {
			for _, t := range texts {
				if tMap, ok := t.(map[string]any); ok {
					if id, ok := tMap["id"].(string); ok && !oldMatIDs[id] {
						preservedTexts = append(preservedTexts, t)
					}
				}
			}
		}
		materials["texts"] = append(preservedTexts, newTexts...)
		trMap["segments"] = newTextSegments
	}

	newContentData, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("lỗi đóng gói draft_content.json: %w", err)
	}
	if err := os.WriteFile(draftContentPath, newContentData, 0644); err != nil {
		return nil, fmt.Errorf("lỗi lưu draft_content.json: %w", err)
	}

	// Also sync draft_info.json if template includes it
	draftInfoPath := filepath.Join(targetProjectDir, "draft_info.json")
	if _, err := os.Stat(draftInfoPath); err == nil {
		_ = os.WriteFile(draftInfoPath, newContentData, 0644)
	}

	// 2. Process draft_meta_info.json
	draftMetaPath := filepath.Join(targetProjectDir, "draft_meta_info.json")
	newDraftID := GenerateUUID()
	nowMicro := time.Now().UnixMicro()

	metaBytes, err := os.ReadFile(draftMetaPath)
	var metaRoot map[string]any
	if err == nil {
		_ = json.Unmarshal(metaBytes, &metaRoot)
	}
	if metaRoot == nil {
		metaRoot = make(map[string]any)
	}

	cleanDirForJSON := filepath.ToSlash(targetProjectDir)
	metaRoot["draft_id"] = newDraftID
	metaRoot["draft_name"] = projectName
	metaRoot["draft_fold_path"] = cleanDirForJSON
	metaRoot["draft_json_file"] = cleanDirForJSON + "/draft_content.json"
	metaRoot["draft_cover"] = "draft_cover.jpg"
	metaRoot["tm_duration"] = float64(totalDurationUs)
	metaRoot["tm_draft_modified"] = float64(nowMicro)
	metaRoot["tm_draft_create"] = float64(nowMicro)

	newMetaData, _ := json.MarshalIndent(metaRoot, "", "  ")
	_ = os.WriteFile(draftMetaPath, newMetaData, 0644)

	// 3. Register into root_meta_info.json
	rootMetaPath := filepath.Join(baseDraftDir, "root_meta_info.json")
	rootMetaBak := filepath.Join(baseDraftDir, "root_meta_info.json.bak")

	var rootMeta map[string]any
	rootMetaBytes, err := os.ReadFile(rootMetaPath)
	if err == nil {
		_ = CopyFile(rootMetaPath, rootMetaBak)
		_ = json.Unmarshal(rootMetaBytes, &rootMeta)
	}
	if rootMeta == nil {
		rootMeta = make(map[string]any)
	}

	var allDraftStore []any
	if store, ok := rootMeta["all_draft_store"].([]any); ok {
		for _, item := range store {
			if itemMap, ok := item.(map[string]any); ok {
				if itemMap["draft_name"] != projectName {
					allDraftStore = append(allDraftStore, item)
				}
			}
		}
	}

	newEntry := map[string]any{
		"cloud_draft_cover":                 false,
		"cloud_draft_sync":                  false,
		"draft_cloud_last_action_download":  false,
		"draft_cloud_purchase_info":         "",
		"draft_cloud_template_id":           "",
		"draft_cloud_tutorial_info":         "",
		"draft_cloud_videocut_purchase_info": "",
		"draft_cover":                       cleanDirForJSON + "/draft_cover.jpg",
		"draft_fold_path":                   cleanDirForJSON,
		"draft_id":                          newDraftID,
		"draft_is_ai_shorts":                false,
		"draft_is_cloud_temp_draft":         false,
		"draft_is_invisible":                false,
		"draft_is_pippit_draft":             false,
		"draft_is_web_article_video":        false,
		"draft_json_file":                   cleanDirForJSON + "/draft_content.json",
		"draft_name":                        projectName,
		"draft_new_version":                 "",
		"draft_root_path":                   filepath.ToSlash(baseDraftDir),
		"draft_timeline_materials_size":     5500000,
		"draft_type":                        "",
		"draft_web_article_video_enter_from": "",
		"pippit_avatar_url":                 "",
		"pippit_extra_info":                 "",
		"pippit_id":                         "",
		"pippit_user_name":                  "",
		"streaming_edit_draft_ready":        true,
		"tm_draft_cloud_completed":          "",
		"tm_draft_cloud_entry_id":           -1,
		"tm_draft_cloud_modified":           0,
		"tm_draft_cloud_parent_entry_id":    -1,
		"tm_draft_cloud_space_id":           -1,
		"tm_draft_cloud_user_id":            -1,
		"tm_draft_create":                   nowMicro,
		"tm_draft_modified":                 nowMicro,
		"tm_draft_removed":                  0,
		"tm_duration":                       totalDurationUs,
	}

	allDraftStore = append([]any{newEntry}, allDraftStore...)
	rootMeta["all_draft_store"] = allDraftStore
	rootMeta["draft_ids"] = len(allDraftStore)

	updatedRootBytes, _ := json.Marshal(rootMeta)
	_ = os.WriteFile(rootMetaPath, updatedRootBytes, 0644)

	return &InjectResult{
		DraftID:      newDraftID,
		ProjectName:  projectName,
		ProjectDir:   targetProjectDir,
		DurationSec:  durationSec,
		DurationUs:   totalDurationUs,
		SubtitlesQty: len(subs),
	}, nil
}
