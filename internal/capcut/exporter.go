package capcut

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"reels-crawl/internal/subtitles"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	shcore               = syscall.NewLazyDLL("shcore.dll")
	procFindWindow       = user32.NewProc("FindWindowW")
	procSetForeground    = user32.NewProc("SetForegroundWindow")
	procKeybdEvent       = user32.NewProc("keybd_event")
	procSetProcessDpi    = shcore.NewProc("SetProcessDpiAwareness")
)

// EnableDPIAwareness sets Per-Monitor DPI Awareness to ensure proper coordinate handling on Windows
func EnableDPIAwareness() {
	if procSetProcessDpi.Find() == nil {
		const PROCESS_PER_MONITOR_DPI_AWARE = 2
		_, _, _ = procSetProcessDpi.Call(uintptr(PROCESS_PER_MONITOR_DPI_AWARE))
	}
}

// FindCapCutExecutable locates CapCut.exe in standard Windows installation directories
func FindCapCutExecutable() string {
	localAppData := os.Getenv("LOCALAPPDATA")
	appsDir := filepath.Join(localAppData, "CapCut", "Apps")

	// 1. Check direct Apps/CapCut.exe
	directCandidate := filepath.Join(appsDir, "CapCut.exe")
	if _, err := os.Stat(directCandidate); err == nil {
		return directCandidate
	}

	// 2. Check version subdirectories in Apps
	if entries, err := os.ReadDir(appsDir); err == nil {
		for i := len(entries) - 1; i >= 0; i-- { // iterate newest version first
			if entries[i].IsDir() {
				candidate := filepath.Join(appsDir, entries[i].Name(), "CapCut.exe")
				if _, err := os.Stat(candidate); err == nil {
					return candidate
				}
			}
		}
	}

	// 3. Check PATH
	if p, err := exec.LookPath("CapCut.exe"); err == nil {
		return p
	}
	return ""
}

// CleanCapCutProcesses kills any running CapCut.exe processes to ensure clean automation
func CleanCapCutProcesses() error {
	cmd := exec.Command("taskkill", "/F", "/IM", "CapCut.exe")
	_ = cmd.Run() // Ignore error if no process was running
	time.Sleep(500 * time.Millisecond)
	return nil
}

// FocusCapCutWindow brings CapCut Desktop to the foreground
func FocusCapCutWindow() bool {
	titlePtr, _ := syscall.UTF16PtrFromString("CapCut")
	hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr)))
	if hwnd != 0 {
		procSetForeground.Call(hwnd)
		return true
	}
	return false
}

// DismissCapCutPopups sends ESC to close any unexpected update/marketing modals
func DismissCapCutPopups() {
	const VK_ESCAPE = 0x1B
	const KEYEVENTF_KEYUP = 0x0002
	procKeybdEvent.Call(VK_ESCAPE, 0, 0, 0)
	time.Sleep(50 * time.Millisecond)
	procKeybdEvent.Call(VK_ESCAPE, 0, KEYEVENTF_KEYUP, 0)
}

// SendExportShortcut sends Ctrl + E keyboard shortcut to trigger CapCut's export dialog
func SendExportShortcut() {
	const VK_CONTROL = 0x11
	const VK_E = 0x45
	const KEYEVENTF_KEYUP = 0x0002

	procKeybdEvent.Call(VK_CONTROL, 0, 0, 0)
	time.Sleep(50 * time.Millisecond)
	procKeybdEvent.Call(VK_E, 0, 0, 0)
	time.Sleep(50 * time.Millisecond)
	procKeybdEvent.Call(VK_E, 0, KEYEVENTF_KEYUP, 0)
	procKeybdEvent.Call(VK_CONTROL, 0, KEYEVENTF_KEYUP, 0)
}

// RenderExportedVideo renders the final video with dynamic subtitles burned-in to 1080p
// saving to downloads/{platform}_{video_id}/exported/final_video.mp4
func RenderExportedVideo(
	rawVideoPath string,
	srtSubtitlePath string,
	exportResolution string,
	outputFinalPath string,
) error {
	if _, err := os.Stat(rawVideoPath); err != nil {
		return fmt.Errorf("video gốc không tồn tại: %w", err)
	}

	outDir := filepath.Dir(outputFinalPath)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("không thể tạo thư mục xuất: %w", err)
	}

	ffmpegBin := subtitles.FindFFmpeg()
	if ffmpegBin == "" {
		return fmt.Errorf("không tìm thấy ffmpeg để render xuất video")
	}

	// Prepare resolution filter
	scaleFilter := "scale=1080:1920:force_original_aspect_ratio=decrease,pad=1080:1920:(ow-iw)/2:(oh-ih)/2"
	if exportResolution == "720p" {
		scaleFilter = "scale=720:1280:force_original_aspect_ratio=decrease,pad=720:1280:(ow-iw)/2:(oh-ih)/2"
	}

	// Check if SRT exists and has content to burn subtitles
	hasSubtitles := false
	if info, err := os.Stat(srtSubtitlePath); err == nil && info.Size() > 10 {
		hasSubtitles = true
	}

	var args []string
	if hasSubtitles {
		// Escape backslashes and colons for ffmpeg subtitles filter on Windows
		cleanSrtPath := filepath.ToSlash(srtSubtitlePath)
		cleanSrtPath = strings.ReplaceAll(cleanSrtPath, ":", "\\:")

		// Beautiful styled subtitle: Yellow text, black outline, rounded styling
		subFilter := fmt.Sprintf("subtitles='%s':force_style='FontSize=16,PrimaryColour=&H00FFFFFF,OutlineColour=&H00000000,BackColour=&H80000000,Outline=2,Shadow=1,MarginV=45,Alignment=2'", cleanSrtPath)
		combinedFilter := fmt.Sprintf("%s,%s", scaleFilter, subFilter)

		args = []string{
			"-i", rawVideoPath,
			"-vf", combinedFilter,
			"-c:v", "libx264",
			"-preset", "veryfast",
			"-crf", "22",
			"-c:a", "aac",
			"-b:a", "192k",
			outputFinalPath,
			"-y",
		}
	} else {
		args = []string{
			"-i", rawVideoPath,
			"-vf", scaleFilter,
			"-c:v", "libx264",
			"-preset", "veryfast",
			"-crf", "22",
			"-c:a", "aac",
			"-b:a", "192k",
			outputFinalPath,
			"-y",
		}
	}

	cmd := exec.Command(ffmpegBin, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("render video lỗi: %s (chi tiết: %w)", string(output), err)
	}

	return nil
}

// OpenProjectInCapCut launches CapCut Desktop
func OpenProjectInCapCut(capcutExePath string) error {
	if capcutExePath == "" {
		capcutExePath = FindCapCutExecutable()
	}
	if capcutExePath == "" {
		return fmt.Errorf("không tìm thấy CapCut.exe trên hệ thống")
	}

	cmd := exec.Command(capcutExePath)
	return cmd.Start()
}
