package crawler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ReelMetadata defines the structured metadata saved for each reel
type ReelMetadata struct {
	ID             string    `json:"id"`
	Platform       string    `json:"platform"`
	OriginalURL    string    `json:"original_url"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Uploader       string    `json:"uploader"`
	Duration       float64   `json:"duration"`
	ViewCount      int64     `json:"view_count"`
	LikeCount      int64     `json:"like_count"`
	CommentCount   int64     `json:"comment_count"`
	ShareCount     int64     `json:"share_count"`
	UploadDate     string    `json:"upload_date"`
	ExtractedLinks []string  `json:"extracted_links"`
	Thumbnail      string    `json:"thumbnail"`
	DownloadedAt   time.Time `json:"downloaded_at"`
	RawVideoPath   string    `json:"raw_video_path"`
}

// DownloadProgress represents the real-time download status
type DownloadProgress struct {
	Percent float64 `json:"percent"`
	Speed   string  `json:"speed"`
	ETA     string  `json:"eta"`
}

// Downloader manages yt-dlp execution
type Downloader struct {
	BinaryPath  string
	BaseDataDir string
	CookiesPath string
}

// NewDownloader creates a new Downloader instance, auto-resolving yt-dlp path
func NewDownloader(binaryPath, baseDataDir, cookiesPath string) *Downloader {
	if binaryPath == "" {
		binaryPath = findYtDlpBinary()
	}
	if baseDataDir == "" {
		baseDataDir = "downloads"
	}
	return &Downloader{
		BinaryPath:  binaryPath,
		BaseDataDir: baseDataDir,
		CookiesPath: cookiesPath,
	}
}

func findYtDlpBinary() string {
	candidates := []string{
		filepath.Join(".", "bin", "yt-dlp.exe"),
		filepath.Join("..", "bin", "yt-dlp.exe"),
		filepath.Join("..", "..", "bin", "yt-dlp.exe"),
		filepath.Join(".", "bin", "yt-dlp"),
		filepath.Join("..", "bin", "yt-dlp"),
		filepath.Join("..", "..", "bin", "yt-dlp"),
		"yt-dlp.exe",
		"yt-dlp",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			abs, err := filepath.Abs(c)
			if err == nil {
				return abs
			}
			return c
		}
	}
	if p, err := exec.LookPath("yt-dlp"); err == nil {
		return p
	}
	return "yt-dlp"
}

// DetectPlatform determines the platform from URL
func DetectPlatform(rawURL string) string {
	lower := strings.ToLower(rawURL)
	if strings.Contains(lower, "tiktok.com") {
		return "tiktok"
	}
	if strings.Contains(lower, "facebook.com") || strings.Contains(lower, "fb.watch") || strings.Contains(lower, "fb.me") {
		return "facebook"
	}
	if strings.Contains(lower, "youtube.com") || strings.Contains(lower, "youtu.be") {
		return "youtube"
	}
	return "facebook"
}

// ExtractVideoID extracts an ID candidate from URL
func ExtractVideoID(rawURL string) string {
	rawURL = strings.TrimPrefix(strings.TrimSpace(rawURL), "\ufeff")
	reFB := regexp.MustCompile(`(?:\?v=|/(?:reel|videos|watch|share/[rv])/)([a-zA-Z0-9_-]+)`)
	if match := reFB.FindStringSubmatch(rawURL); len(match) > 1 {
		return match[1]
	}
	reTT := regexp.MustCompile(`/(?:video|v)/(\d+)`)
	if match := reTT.FindStringSubmatch(rawURL); len(match) > 1 {
		return match[1]
	}
	reShortTT := regexp.MustCompile(`(?:vt\.tiktok\.com|tiktok\.com/t)/([a-zA-Z0-9_-]+)`)
	if match := reShortTT.FindStringSubmatch(rawURL); len(match) > 1 {
		return match[1]
	}
	// Fallback
	clean := regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(rawURL, "_")
	if len(clean) > 20 {
		clean = clean[len(clean)-20:]
	}
	return clean
}

// IsProfileURL checks if the provided URL represents a profile or channel rather than a single video
func IsProfileURL(rawURL string) bool {
	lower := strings.ToLower(rawURL)
	if strings.Contains(lower, "tiktok.com/@") && !strings.Contains(lower, "/video/") {
		return true
	}
	if (strings.Contains(lower, "facebook.com/") || strings.Contains(lower, "fb.com/")) &&
		!strings.Contains(lower, "/reel/") &&
		!strings.Contains(lower, "/videos/") &&
		!strings.Contains(lower, "/watch") &&
		!strings.Contains(lower, "/share/") {
		return true
	}
	return false
}

// FetchChannelVideos scans a profile/channel URL and returns up to maxVideos latest reel/video URLs
func (d *Downloader) FetchChannelVideos(profileURL string, maxVideos int) ([]string, error) {
	if maxVideos <= 0 {
		maxVideos = 10
	}

	args := []string{
		"--flat-playlist",
		"--dump-json",
		"--playlist-end", strconv.Itoa(maxVideos),
		"--no-warnings",
	}
	if d.CookiesPath != "" {
		if _, err := os.Stat(d.CookiesPath); err == nil {
			args = append(args, "--cookies", d.CookiesPath)
		}
	}
	args = append(args, profileURL)

	cmd := exec.Command(d.BinaryPath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("không thể quét kênh/profile: %w", err)
	}

	var urls []string
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var item map[string]any
		if err := json.Unmarshal([]byte(line), &item); err == nil {
			targetURL, _ := item["url"].(string)
			if targetURL == "" {
				targetURL, _ = item["webpage_url"].(string)
			}
			if targetURL == "" {
				if id, ok := item["id"].(string); ok && id != "" {
					if DetectPlatform(profileURL) == "tiktok" {
						targetURL = fmt.Sprintf("https://www.tiktok.com/video/%s", id)
					} else {
						targetURL = fmt.Sprintf("https://www.facebook.com/reel/%s", id)
					}
				}
			}
			if targetURL != "" && !seen[targetURL] {
				seen[targetURL] = true
				urls = append(urls, targetURL)
			}
		}
	}
	return urls, nil
}

// FetchMetadata retrieves metadata without downloading the video
func (d *Downloader) FetchMetadata(url string) (*ReelMetadata, error) {
	platform := DetectPlatform(url)
	args := []string{"--dump-json", "--no-download", "--no-warnings"}
	if d.CookiesPath != "" {
		if _, err := os.Stat(d.CookiesPath); err == nil {
			args = append(args, "--cookies", d.CookiesPath)
		}
	}
	args = append(args, url)

	cmd := exec.Command(d.BinaryPath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp metadata failed: %w (output: %s)", err, string(output))
	}

	var raw map[string]any
	if err := json.Unmarshal(output, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse yt-dlp json: %w", err)
	}

	id, _ := raw["id"].(string)
	if id == "" {
		id, _ = raw["display_id"].(string)
	}
	if id == "" {
		id = ExtractVideoID(url)
	}
	title, _ := raw["title"].(string)
	desc, _ := raw["description"].(string)
	uploader, _ := raw["uploader"].(string)
	duration, _ := raw["duration"].(float64)
	thumbnail, _ := raw["thumbnail"].(string)
	uploadDate, _ := raw["upload_date"].(string)

	var viewCount, likeCount, commentCount, shareCount int64
	if v, ok := raw["view_count"].(float64); ok {
		viewCount = int64(v)
	}
	if l, ok := raw["like_count"].(float64); ok {
		likeCount = int64(l)
	}
	if c, ok := raw["comment_count"].(float64); ok {
		commentCount = int64(c)
	}
	if s, ok := raw["repost_count"].(float64); ok {
		shareCount = int64(s)
	}

	extractedLinks := ExtractLinksFromText(desc)
	// Also inspect author's comments if present in metadata
	if comments, ok := raw["comments"].([]any); ok {
		for _, c := range comments {
			if cMap, ok := c.(map[string]any); ok {
				cAuthor, _ := cMap["author"].(string)
				cText, _ := cMap["text"].(string)
				if cAuthor == uploader || uploader == "" {
					extractedLinks = append(extractedLinks, ExtractLinksFromText(cText)...)
				}
			}
		}
	}

	return &ReelMetadata{
		ID:             id,
		Platform:       platform,
		OriginalURL:    url,
		Title:          title,
		Description:    desc,
		Uploader:       uploader,
		Duration:       duration,
		ViewCount:      viewCount,
		LikeCount:      likeCount,
		CommentCount:   commentCount,
		ShareCount:     shareCount,
		UploadDate:     uploadDate,
		ExtractedLinks: extractedLinks,
		Thumbnail:      thumbnail,
		DownloadedAt:   time.Now(),
	}, nil
}

// ProbeDuration uses ffprobe to get video duration in seconds
func ProbeDuration(videoPath string) float64 {
	ffprobeBin := "ffprobe"
	candidates := []string{
		filepath.Join(".", "bin", "ffprobe.exe"),
		filepath.Join("..", "bin", "ffprobe.exe"),
		filepath.Join("..", "..", "bin", "ffprobe.exe"),
		filepath.Join(".", "bin", "ffprobe"),
		filepath.Join("..", "bin", "ffprobe"),
		filepath.Join("..", "..", "bin", "ffprobe"),
		"ffprobe.exe",
		"ffprobe",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			if abs, err := filepath.Abs(c); err == nil {
				ffprobeBin = abs
				break
			}
			ffprobeBin = c
			break
		}
	}

	cmd := exec.Command(ffprobeBin,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)
	out, err := cmd.Output()
	if err == nil {
		s := strings.TrimSpace(string(out))
		if val, err := strconv.ParseFloat(s, 64); err == nil && val > 0 {
			return val
		}
	}
	return 15.0 // fallback
}

// Download downloads the reel and emits progress updates via callback
func (d *Downloader) Download(url string, onProgress func(DownloadProgress)) (*ReelMetadata, error) {
	meta, err := d.FetchMetadata(url)
	if err != nil {
		// Resilient fallback: initialize synthetic metadata and proceed to direct yt-dlp download
		plat := DetectPlatform(url)
		vid := ExtractVideoID(url)
		if vid == "" {
			vid = fmt.Sprintf("%s_%d", plat, time.Now().UnixNano()%1000000)
		}
		meta = &ReelMetadata{
			ID:           vid,
			Platform:     plat,
			OriginalURL:  url,
			Title:        fmt.Sprintf("Video %s", vid),
			Description:  fmt.Sprintf("Tự động cào từ %s", url),
			DownloadedAt: time.Now(),
		}
	}

	targetDir := filepath.Join(d.BaseDataDir, fmt.Sprintf("%s_%s", meta.Platform, meta.ID))
	rawVideoPath := filepath.Join(targetDir, "raw_video.mp4")
	metaPath := filepath.Join(targetDir, "metadata.json")

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	// Deduplication check: if video already exists and has content
	if info, err := os.Stat(rawVideoPath); err == nil && info.Size() > 1024 {
		meta.RawVideoPath = rawVideoPath
		if onProgress != nil {
			onProgress(DownloadProgress{Percent: 100, Speed: "Deduplicated", ETA: "00:00"})
		}
		// ensure metadata.json is written
		metaBytes, _ := json.MarshalIndent(meta, "", "  ")
		_ = os.WriteFile(metaPath, metaBytes, 0644)
		return meta, nil
	}

	args := []string{
		"-f", "b/best",
		"--no-warnings",
		"--newline",
		"-o", rawVideoPath,
	}

	if d.CookiesPath != "" {
		if _, err := os.Stat(d.CookiesPath); err == nil {
			args = append(args, "--cookies", d.CookiesPath)
		}
	}
	args = append(args, url)

	cmd := exec.Command(d.BinaryPath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start yt-dlp: %w", err)
	}

	progressRegex := regexp.MustCompile(`\[download\]\s+(\d+\.?\d*)%\s+of\s+([^\s]+)\s+at\s+([^\s]+)\s+ETA\s+([^\s]+)`)

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		matches := progressRegex.FindStringSubmatch(line)
		if len(matches) >= 5 {
			pct, _ := strconv.ParseFloat(matches[1], 64)
			if onProgress != nil {
				onProgress(DownloadProgress{
					Percent: pct,
					Speed:   matches[3],
					ETA:     matches[4],
				})
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("download command failed: %w", err)
	}

	meta.RawVideoPath = rawVideoPath
	if meta.Duration <= 0 {
		meta.Duration = ProbeDuration(rawVideoPath)
	}
	metaBytes, err := json.MarshalIndent(meta, "", "  ")
	if err == nil {
		_ = os.WriteFile(metaPath, metaBytes, 0644)
	}

	if onProgress != nil {
		onProgress(DownloadProgress{Percent: 100, Speed: "Complete", ETA: "00:00"})
	}

	return meta, nil
}
