package subtitles

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// SubtitleSegment represents a timestamped subtitle segment
type SubtitleSegment struct {
	ID         int     `json:"id"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Text       string  `json:"text"`
	StartUs    int64   `json:"start_us"`
	DurationUs int64   `json:"duration_us"`
}

// WhisperVerboseResponse maps the Groq / OpenAI Whisper verbose_json output
type WhisperVerboseResponse struct {
	Task     string  `json:"task"`
	Language string  `json:"language"`
	Duration float64 `json:"duration"`
	Text     string  `json:"text"`
	Segments []struct {
		ID    int     `json:"id"`
		Start float64 `json:"start"`
		End   float64 `json:"end"`
		Text  string  `json:"text"`
	} `json:"segments"`
}

// FindFFmpeg locates the ffmpeg executable on the system
// FindFFmpeg locates the ffmpeg executable on the system
func FindFFmpeg() string {
	candidates := []string{
		filepath.Join(".", "bin", "ffmpeg.exe"),
		filepath.Join("..", "bin", "ffmpeg.exe"),
		filepath.Join("..", "..", "bin", "ffmpeg.exe"),
		filepath.Join(".", "bin", "ffmpeg"),
		filepath.Join("..", "bin", "ffmpeg"),
		filepath.Join("..", "..", "bin", "ffmpeg"),
		"ffmpeg.exe",
		"ffmpeg",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			if abs, err := filepath.Abs(c); err == nil {
				return abs
			}
			return c
		}
	}
	if p, err := exec.LookPath("ffmpeg"); err == nil {
		return p
	}
	return "ffmpeg"
}

// ExtractAudio extracts audio from video file to 16kHz mono WAV or m4a
func ExtractAudio(ffmpegPath, videoPath, audioPath string) error {
	if ffmpegPath == "" {
		ffmpegPath = FindFFmpeg()
	}

	// First try: 16kHz mono PCM WAV as recommended by Whisper specs
	cmd := exec.Command(ffmpegPath,
		"-i", videoPath,
		"-vn",
		"-ar", "16000",
		"-ac", "1",
		"-c:a", "pcm_s16le",
		audioPath,
		"-y",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback: direct audio stream copy
		cmd2 := exec.Command(ffmpegPath,
			"-i", videoPath,
			"-vn",
			"-c:a", "copy",
			audioPath,
			"-y",
		)
		out2, err2 := cmd2.CombinedOutput()
		if err2 != nil {
			return fmt.Errorf("ffmpeg failed: %s (fallback: %s)", string(output), string(out2))
		}
	}
	return nil
}

// FormatSRTTime converts seconds (e.g. 1.25) to SRT timestamp "00:00:01,250"
func FormatSRTTime(seconds float64) string {
	hrs := int(seconds) / 3600
	mins := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	millis := int(math.Round((seconds - float64(int(seconds))) * 1000))
	if millis >= 1000 {
		millis = 999
	}
	return fmt.Sprintf("%02d:%02d:%02d,%03d", hrs, mins, secs, millis)
}

// TranscribeWithGroq sends audio to Groq Whisper API
func TranscribeWithGroq(audioPath string, apiKey string) ([]SubtitleSegment, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("groq API key is empty")
	}

	audioFile, err := os.Open(audioPath)
	if err != nil {
		return nil, err
	}
	defer audioFile.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(audioPath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, audioFile); err != nil {
		return nil, err
	}

	_ = writer.WriteField("model", "whisper-large-v3")
	_ = writer.WriteField("response_format", "verbose_json")
	_ = writer.WriteField("temperature", "0")
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/audio/transcriptions", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("groq API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("groq API error (status %d): %s", resp.StatusCode, string(respBytes))
	}

	var wResp WhisperVerboseResponse
	if err := json.Unmarshal(respBytes, &wResp); err != nil {
		return nil, fmt.Errorf("failed to parse groq response: %w", err)
	}

	segments := make([]SubtitleSegment, 0, len(wResp.Segments))
	for i, s := range wResp.Segments {
		txt := strings.TrimSpace(s.Text)
		if txt == "" {
			continue
		}
		segments = append(segments, SubtitleSegment{
			ID:         i + 1,
			Start:      s.Start,
			End:        s.End,
			Text:       txt,
			StartUs:    int64(s.Start * 1e6),
			DurationUs: int64((s.End - s.Start) * 1e6),
		})
	}

	return segments, nil
}

// WhisperModelInfo describes a downloadable/installed local model
type WhisperModelInfo struct {
	Name        string `json:"name"`
	SizeMB      int    `json:"size_mb"`
	FileName    string `json:"file_name"`
	DownloadURL string `json:"download_url"`
	Installed   bool   `json:"installed"`
	Path        string `json:"path"`
}

// AvailableWhisperModels provides metadata for the supported offline Whisper models
var AvailableWhisperModels = map[string]WhisperModelInfo{
	"tiny": {
		Name:        "Whisper Tiny",
		SizeMB:      75,
		FileName:    "ggml-tiny.bin",
		DownloadURL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.bin",
	},
	"base": {
		Name:        "Whisper Base",
		SizeMB:      142,
		FileName:    "ggml-base.bin",
		DownloadURL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin",
	},
	"small": {
		Name:        "Whisper Small",
		SizeMB:      466,
		FileName:    "ggml-small.bin",
		DownloadURL: "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin",
	},
}

// ListWhisperModels returns the status of all supported local Whisper models
func ListWhisperModels(modelsDir string) []WhisperModelInfo {
	if modelsDir == "" {
		modelsDir = "models"
	}
	_ = os.MkdirAll(modelsDir, 0755)

	keys := []string{"tiny", "base", "small"}
	result := make([]WhisperModelInfo, 0, len(keys))

	for _, k := range keys {
		info := AvailableWhisperModels[k]
		targetPath := filepath.Join(modelsDir, info.FileName)
		if fi, err := os.Stat(targetPath); err == nil && fi.Size() > 1024*1024 {
			info.Installed = true
			info.Path = targetPath
		}
		result = append(result, info)
	}
	return result
}

// DownloadWhisperModel downloads a model to models/ with live progress callback
func DownloadWhisperModel(modelName string, modelsDir string, onProgress func(pct float64, speed string)) error {
	info, ok := AvailableWhisperModels[strings.ToLower(modelName)]
	if !ok {
		return fmt.Errorf("model không hợp lệ (hỗ trợ: tiny, base, small): %s", modelName)
	}

	if modelsDir == "" {
		modelsDir = "models"
	}
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		return err
	}

	targetPath := filepath.Join(modelsDir, info.FileName)
	partPath := targetPath + ".part"

	resp, err := http.Get(info.DownloadURL)
	if err != nil {
		return fmt.Errorf("không thể kết nối tải model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("lỗi tải model (HTTP %d)", resp.StatusCode)
	}

	totalBytes := resp.ContentLength
	out, err := os.Create(partPath)
	if err != nil {
		return err
	}
	defer out.Close()

	buf := make([]byte, 64*1024)
	var downloaded int64
	startTime := time.Now()
	lastReport := time.Now()

	for {
		n, rErr := resp.Body.Read(buf)
		if n > 0 {
			if _, wErr := out.Write(buf[:n]); wErr != nil {
				return wErr
			}
			downloaded += int64(n)

			if time.Since(lastReport) > 200*time.Millisecond || rErr != nil {
				lastReport = time.Now()
				elapsed := time.Since(startTime).Seconds()
				var speedStr string
				if elapsed > 0 {
					speedMB := float64(downloaded) / (1024 * 1024) / elapsed
					speedStr = fmt.Sprintf("%.1f MB/s", speedMB)
				}
				var pct float64
				if totalBytes > 0 {
					pct = float64(downloaded) / float64(totalBytes) * 100.0
				}
				if onProgress != nil {
					onProgress(pct, speedStr)
				}
			}
		}
		if rErr != nil {
			if rErr == io.EOF {
				break
			}
			return rErr
		}
	}

	_ = out.Close()
	_ = os.Remove(targetPath)
	return os.Rename(partPath, targetPath)
}

// GenerateDynamicSubtitles generates dynamic subtitle segments based on real video metadata, duration and caption text.
// It breaks text into natural reading clauses and maps timestamps dynamically across the actual video duration.
func GenerateDynamicSubtitles(caption string, totalDurationSec float64) []SubtitleSegment {
	return GenerateDynamicSubtitlesFromMeta("", caption, "", totalDurationSec)
}

// GenerateDynamicSubtitlesFromMeta creates dynamically timestamped subtitles using actual video metadata.
// Entirely dynamic, never using hardcoded filler phrases.
func GenerateDynamicSubtitlesFromMeta(title string, caption string, uploader string, totalDurationSec float64) []SubtitleSegment {
	if totalDurationSec <= 0 {
		totalDurationSec = 15.0
	}

	// 1. Gather all dynamic textual content from video
	var rawText string
	cleanCaption := strings.TrimSpace(caption)
	// Remove URLs from caption
	words := strings.Fields(cleanCaption)
	var filteredWords []string
	for _, w := range words {
		if !strings.HasPrefix(w, "http://") && !strings.HasPrefix(w, "https://") && !strings.HasPrefix(w, "#") {
			filteredWords = append(filteredWords, w)
		}
	}
	cleanedCaption := strings.Join(filteredWords, " ")

	if cleanedCaption != "" {
		rawText = cleanedCaption
	} else if title != "" {
		rawText = strings.TrimSpace(title)
	}

	var phrases []string
	if rawText != "" {
		// Split by multiple punctuation marks and clause boundaries
		delimiters := []string{". ", "! ", "? ", "\n", "; ", " - ", " / ", " | "}
		currentParts := []string{rawText}
		for _, d := range delimiters {
			var nextParts []string
			for _, p := range currentParts {
				split := strings.Split(p, d)
				for _, s := range split {
					trimmed := strings.TrimSpace(s)
					if len(trimmed) > 0 {
						nextParts = append(nextParts, trimmed)
					}
				}
			}
			currentParts = nextParts
		}

		// Further split any overly long phrase (> 8 words or > 60 chars) into natural chunks
		for _, part := range currentParts {
			partWords := strings.Fields(part)
			if len(partWords) > 8 {
				chunkSize := 6
				for i := 0; i < len(partWords); i += chunkSize {
					end := i + chunkSize
					if end > len(partWords) {
						end = len(partWords)
					}
					phrases = append(phrases, strings.Join(partWords[i:end], " "))
				}
			} else {
				phrases = append(phrases, part)
			}
		}
	}

	// If phrases are still empty, build dynamic phrases from actual video identity
	if len(phrases) == 0 {
		if title != "" {
			phrases = append(phrases, title)
		}
		if uploader != "" {
			phrases = append(phrases, fmt.Sprintf("@%s", uploader))
		}
		if len(phrases) == 0 {
			phrases = append(phrases, "Reel Content")
		}
	}

	// Calculate timing: each subtitle ideally displays for 2 to 4 seconds
	numSegments := len(phrases)
	maxSegments := int(totalDurationSec / 2.0)
	if maxSegments < 1 {
		maxSegments = 1
	}
	if numSegments > maxSegments {
		numSegments = maxSegments
		phrases = phrases[:numSegments]
	}

	step := totalDurationSec / float64(numSegments)
	segments := make([]SubtitleSegment, 0, numSegments)

	for i, phrase := range phrases {
		start := float64(i) * step
		end := start + step
		if i == numSegments-1 {
			end = totalDurationSec
		}
		durUs := int64((end - start) * 1e6)
		startUs := int64(start * 1e6)

		segments = append(segments, SubtitleSegment{
			ID:         i + 1,
			Start:      start,
			End:        end,
			Text:       phrase,
			StartUs:    startUs,
			DurationUs: durUs,
		})
	}

	return segments
}

// SaveSubtitles writes both subtitles.srt and subtitles.json to the target video directory
func SaveSubtitles(segments []SubtitleSegment, targetDir string) error {
	if len(segments) == 0 {
		return fmt.Errorf("no subtitle segments to save")
	}

	var srtBuilder strings.Builder
	for _, seg := range segments {
		srtBuilder.WriteString(fmt.Sprintf("%d\n%s --> %s\n%s\n\n",
			seg.ID,
			FormatSRTTime(seg.Start),
			FormatSRTTime(seg.End),
			seg.Text,
		))
	}

	srtPath := filepath.Join(targetDir, "subtitles.srt")
	if err := os.WriteFile(srtPath, []byte(srtBuilder.String()), 0644); err != nil {
		return err
	}

	jsonPath := filepath.Join(targetDir, "subtitles.json")
	jsonData, err := json.MarshalIndent(segments, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(jsonPath, jsonData, 0644)
}

// FindWhisperExecutable searches for whisper-cli or whisper binary in bin/ and system PATH
func FindWhisperExecutable() string {
	candidates := []string{
		filepath.Join(".", "bin", "whisper-cli.exe"),
		filepath.Join("..", "bin", "whisper-cli.exe"),
		filepath.Join("..", "..", "bin", "whisper-cli.exe"),
		filepath.Join(".", "bin", "whisper.exe"),
		filepath.Join("..", "bin", "whisper.exe"),
		filepath.Join("..", "..", "bin", "whisper.exe"),
		filepath.Join(".", "bin", "whisper-cli"),
		filepath.Join(".", "bin", "whisper"),
		"whisper-cli.exe",
		"whisper.exe",
		"whisper-cli",
		"whisper",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			if abs, err := filepath.Abs(c); err == nil {
				return abs
			}
			return c
		}
	}
	if p, err := exec.LookPath("whisper-cli"); err == nil {
		return p
	}
	if p, err := exec.LookPath("whisper"); err == nil {
		return p
	}
	return ""
}

var srtTimeRegex = regexp.MustCompile(`(\d{2}):(\d{2}):(\d{2})[,.](\d{3})\s*-->\s*(\d{2}):(\d{2}):(\d{2})[,.](\d{3})`)

// parseSRTTimestamp parses "00:00:01,250" to seconds float64
func parseSRTTimestamp(h, m, s, ms string) float64 {
	hi, _ := strconv.Atoi(h)
	mi, _ := strconv.Atoi(m)
	si, _ := strconv.Atoi(s)
	msi, _ := strconv.Atoi(ms)
	return float64(hi*3600+mi*60+si) + float64(msi)/1000.0
}

// ParseSRTContent parses SRT text into a slice of SubtitleSegment
func ParseSRTContent(content string) []SubtitleSegment {
	var segments []SubtitleSegment
	scanner := bufio.NewScanner(strings.NewReader(content))

	var currentStart, currentEnd float64
	var textLines []string
	inSegment := false
	segID := 1

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			if inSegment && len(textLines) > 0 {
				txt := strings.Join(textLines, " ")
				segments = append(segments, SubtitleSegment{
					ID:         segID,
					Start:      currentStart,
					End:        currentEnd,
					Text:       txt,
					StartUs:    int64(currentStart * 1e6),
					DurationUs: int64((currentEnd - currentStart) * 1e6),
				})
				segID++
			}
			inSegment = false
			textLines = nil
			continue
		}

		// Check if timestamp line
		if matches := srtTimeRegex.FindStringSubmatch(line); len(matches) == 9 {
			currentStart = parseSRTTimestamp(matches[1], matches[2], matches[3], matches[4])
			currentEnd = parseSRTTimestamp(matches[5], matches[6], matches[7], matches[8])
			inSegment = true
			textLines = nil
			continue
		}

		// If in segment and not just numeric ID line
		if inSegment {
			textLines = append(textLines, line)
		}
	}

	// Flush trailing segment if file did not end with blank line
	if inSegment && len(textLines) > 0 {
		txt := strings.Join(textLines, " ")
		segments = append(segments, SubtitleSegment{
			ID:         segID,
			Start:      currentStart,
			End:        currentEnd,
			Text:       txt,
			StartUs:    int64(currentStart * 1e6),
			DurationUs: int64((currentEnd - currentStart) * 1e6),
		})
	}

	return segments
}

// ParseSRTFile reads and parses an SRT file
func ParseSRTFile(srtPath string) ([]SubtitleSegment, error) {
	data, err := os.ReadFile(srtPath)
	if err != nil {
		return nil, err
	}
	return ParseSRTContent(string(data)), nil
}

// TranscribeWithLocalWhisper runs whisper-cli against audioPath with modelPath
func TranscribeWithLocalWhisper(whisperBin, modelPath, audioPath string) ([]SubtitleSegment, error) {
	if whisperBin == "" {
		whisperBin = FindWhisperExecutable()
	}
	if whisperBin == "" {
		return nil, fmt.Errorf("không tìm thấy file thực thi whisper-cli")
	}
	if _, err := os.Stat(modelPath); err != nil {
		return nil, fmt.Errorf("model không tồn tại: %w", err)
	}

	outPrefix := audioPath + "_whisper"
	args := []string{
		"-m", modelPath,
		"-f", audioPath,
		"-osrt",
		"-of", outPrefix,
	}

	cmd := exec.Command(whisperBin, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("whisper-cli execution failed: %s (err: %w)", string(output), err)
	}

	srtFile := outPrefix + ".srt"
	defer os.Remove(srtFile)

	return ParseSRTFile(srtFile)
}
